# HTTP API

`github.com/zebodotdev/httpapi` is a reusable HTTP contract layer for Go
services.

It helps service teams define endpoint contracts, parse request bodies into
domain-ready values, render caller-aware responses, enforce runtime endpoint
requirements, and transcribe endpoint groups into OpenAPI documents.

The module is intentionally infrastructure-neutral. It does not know about your
database, auth provider, logging stack, deployment platform, or service
configuration system. Applications plug those pieces in at the package
boundaries.

## Install

```sh
go get github.com/zebodotdev/httpapi
```

## Mental Model

Use the narrow package that owns the behavior you need:

- `caller` defines stable labels for trusted request sources.
- `request` turns an incoming `*http.Request` into a safe `Req` with buffered
  body, request ID, caller, session, and redacted audit output.
- `param` parses JSON request bodies into endpoint-owned domain parameters.
- `response` builds responses, projects response shapes, filters
  caller-specific attributes, and writes HTTP responses.
- `endpoint` defines endpoint contracts and enforces method, content type,
  auth, caller availability, idempotency, timeout, and request-size policy.
- `openapi/*` transcribes endpoint metadata into target documents.

The important separation is:

- request parsing prepares an acceptable request;
- endpoint definitions describe what a route requires;
- param parsing accepts and normalizes endpoint payload data;
- response shapes describe what an endpoint returns.

Do not put endpoint requirements on `request.Req`. `Req` should not know which
endpoint will handle it.

## Callers

A `caller.Caller` is an application-defined label for the trusted source of a
request. It is not an auth scheme. The same caller values can restrict endpoint
access, request parameters, and response attributes.

Define callers once and reuse the values:

```go
package tasksapi

import "github.com/zebodotdev/httpapi/caller"

var (
	PublicAPI = caller.Define("public-api")
	Worker    = caller.Define("worker")
	Dashboard = caller.Define("dashboard")
)
```

Trusted middleware should attach the active caller with
`request.ContextWithCaller` before the endpoint runtime builds `request.Req`.

When no caller restriction is configured, the object is available to every
caller. This applies consistently to endpoint groups, endpoints, params, and
response attributes.

## Request Params

Use `param` to parse, not to perform a separate validation pass. Parsing should
return a complete value that downstream code can safely use, or a `*param.Error`
that is safe to convert into an API error response.

Define parsers once near the endpoint:

```go
type createTaskParams struct {
	Title      string
	Tags       []string
	AssigneeID string
	Automation automationParams
}

type automationParams struct {
	Name  string
	RunAt *time.Time
}

var createTaskRequest = param.JSON[createTaskParams]().
	Param(param.Required("title", param.String()).
		Null(param.NullRejected).
		MinSize(1).
		MaxSize(160).
		Parse(param.NonEmptyTrimmedString)).
	Param(param.Optional("tag_names", param.Array[string]()).
		MaxItems(20).
		Parse(param.TrimmedStringList)).
	Param(param.Optional("assignee_id", param.String()).
		Parse(parseTaskAssigneeID)).
	Param(param.Optional("automation",
		param.Object[automationParams]().
			Param(param.Required("name", param.String()).
				Parse(param.NonEmptyTrimmedString)).
			Param(param.Optional("run_at", param.String()).
				Parse(param.OptionalRFC3339TimestampPointer)).
			Parse(parseAutomation),
	).AvailableTo(Worker)).
	AtMostOne("assignee_id", "automation").
	Parse(parseCreateTask)
```

Inside the final parser, use `param.Must` for required parameters and
`param.Get` for optional parameters:

```go
func parseCreateTask(values param.Values) (createTaskParams, error) {
	params := createTaskParams{
		Title: param.Must[string](values, "title"),
	}

	if tags, ok := param.Get[[]string](values, "tag_names"); ok {
		params.Tags = tags
	}
	if assigneeID, ok := param.Get[string](values, "assignee_id"); ok {
		params.AssigneeID = assigneeID
	}
	if automation, ok := param.Get[automationParams](values, "automation"); ok {
		params.Automation = automation
	}

	return params, nil
}
```

Runtime usage stays direct:

```go
params, err := createTaskRequest.Parse(
	r.Body,
	param.WithRequestCaller(r),
)
if err != nil {
	renderParamError(r, err)
	return
}
```

Services choose their own public error envelope. A small adapter is usually
enough:

```go
func renderParamError(r response.Target, err *param.Error) {
	if err == nil {
		response.RenderErr(r, nil)
		return
	}

	response.RenderErr(r, erreur.InvalidParam(
		string(err.Code),
		err.Message,
		"",
	))
}
```

Restricted params deliberately fail as unexpected parameters when sent by a
caller that cannot use them. That avoids revealing that a hidden parameter
exists or which caller is allowed to send it.

## Responses

Handlers can render ordinary payloads:

```go
response.RenderJSON(r, http.StatusOK, map[string]any{
	"status": "ready",
})

response.RenderNoContent(r)
response.RenderRedirect(r, http.StatusSeeOther, "https://example.test/next")
response.RenderStream(r, http.StatusOK, "application/pdf", header, reader)
```

Use response shapes when the JSON object is part of the endpoint contract or
when attributes have caller availability rules. A response shape is explicit:
each emitted JSON key is named with its JSON type and extractor.

For stored/domain values that need nil guards, defaulting, or redaction before
attribute extraction, define a small response view and adapt the domain value
with `response.Project`:

```go
type taskView struct {
	task Task
}

func taskValue(task *Task) taskView {
	if task == nil {
		return taskView{}
	}
	return taskView{task: *task}
}

func (task taskView) ID() string { return task.task.ID }
func (task taskView) Status() string { return task.task.Status }
func (task taskView) CreatedAt() time.Time { return task.task.CreatedAt }
func (task taskView) InternalNote() (string, bool) {
	return task.task.InternalNote, task.task.InternalNote != ""
}

var taskResponse = response.Project(
	response.Object[taskView](
		response.Required("id", response.String(), taskView.ID),
		response.Required("status", response.String(), taskView.Status),
		response.Required("created_at", response.Time(), taskView.CreatedAt),
		response.Optional("internal_note", response.String(), taskView.InternalNote).
			AvailableTo(Worker, Dashboard),
	),
	taskValue,
)
```

Render the shaped body from the handler:

```go
response.RenderJSON(r, http.StatusCreated, taskResponse.Body(task))
```

When the target is `*request.Req` or `*endpoint.Req`, `response` derives the
caller from `r.RequestCaller()`. Do not pass the caller manually from handlers.

Use `response.Object[T]` directly when the handler already has a response view
value. Use `response.Project(shape, prepare)` when the handler has a domain
value and the shape should read against a prepared response view.

## Endpoint Definitions

Use `endpoint.DefineEndpoint(endpoint.EndpointSpec{...})` for new endpoints.
The struct keeps the endpoint contract in one readable place.

```go
var CreateTask = endpoint.DefineEndpoint(endpoint.EndpointSpec{
	Method: endpoint.POST,
	Path:   "/tasks/create",
	Handler: func(r *endpoint.Req) {
		params, err := createTaskRequest.Parse(r.Body, param.WithRequestCaller(r))
		if err != nil {
			renderParamError(r, err)
			return
		}

		task, appErr := service.CreateTask(r.Req.Context(), params)
		if appErr != nil {
			response.RenderErr(r, appErr)
			return
		}

		response.RenderJSON(r, http.StatusCreated, taskResponse.Body(task))
	},
	Access: endpoint.EndpointAccessSpec{
		Authorization: endpoint.RequiredAuthorization(endpoint.AuthorizationKindBearer),
		Callers:       []caller.Caller{PublicAPI, Dashboard},
	},
	Idempotency: endpoint.EndpointIdempotencySpec{
		Enabled: true,
	},
	Priority: endpoint.PriorityHigh,
	Timeout: endpoint.EndpointTimeoutSpec{
		ReadBody: 2 * time.Second,
		Handler:  10 * time.Second,
		Write:    2 * time.Second,
	},
	Limits: endpoint.EndpointLimitsSpec{
		MaxRequestBytes: 64 << 10,
	},
	Route: endpoint.RouteSpec{
		OperationID: "create_task",
		Summary:     "Create task",
		Backend: endpoint.RouteBackend{
			Address:  "https://tasks.example.internal",
			PathMode: endpoint.RoutePathModeAppend,
			Timeout:  30 * time.Second,
		},
	},
})
```

Prefer endpoint-level `Timeout`, `Limits`, `Priority`, `Access`, and `Route`
metadata over service-local side tables. Transcribers can only produce complete
documents when the endpoint contract carries the relevant metadata.

`NewEndpoint`, `NewIdempotentEndpoint`, and
`NewIdempotentEndpointWithScopeResolver` remain for compatibility. New code
should prefer `DefineEndpoint`.

## Endpoint Groups

Use endpoint groups for shared defaults and mounting:

```go
group := endpoint.EndpointGroup{
	PathPrefix: "/v1",
	Route: endpoint.RouteSpec{
		Backend: endpoint.RouteBackend{
			Address: "https://tasks.example.internal",
		},
	},
	Timeout: endpoint.EndpointTimeoutSpec{
		Handler: 5 * time.Second,
	},
	Endpoints: []endpoint.Endpoint{
		CreateTask,
	},
}

group.RequireAuthorization(endpoint.AuthorizationKindBearer)
group.AvailableTo(PublicAPI, Dashboard)
group.Mount(mux)
```

Group availability narrows endpoint availability; it does not widen a restricted
endpoint. An unrestricted group or endpoint is available to all callers.

## Service Wiring

Applications own the real authentication, audit, and idempotency systems.
`httpapi` provides injection points.

Configure authentication:

```go
restore := request.ConfigureAuthenticator(request.AuthenticatorFunc(
	func(ctx context.Context, req *request.Req, auth request.AuthenticationRequest) (*request.Session, error) {
		// Call your auth boundary here, then return a session.
		return &request.Session{
			ID:       "sess_...",
			App:      request.App{ID: "app_..."},
			AuthMode: auth.Type,
		}, nil
	},
))
defer restore()
```

Configure auth scheme names when your API uses non-default authorization
prefixes:

```go
restore := request.ConfigureAuthorizationSchemes(request.AuthorizationSchemes{
	Service: "Internal-Service",
})
defer restore()
```

Configure request audit persistence:

```go
restore := endpoint.ConfigureAuditSink(endpoint.AuditSinkFunc(
	func(ctx context.Context, req *endpoint.Req) error {
		return auditStore.Save(ctx, req)
	},
))
defer restore()
```

Configure idempotency before mounting idempotent endpoints:

```go
restoreStore := endpoint.ConfigureIdempotencyStore(store)
defer restoreStore()

restoreNamespace := endpoint.ConfigureIdempotencyScopeNamespace("tasks")
defer restoreNamespace()
```

Idempotent endpoints fail closed with `idempotency_storage_unavailable` until a
store is configured.

## Transcription

Endpoint metadata is provider-neutral. Target packages decide how to translate
that metadata.

Public OpenAPI 3.1 documents:

```go
doc, err := openapi31.Transcriber{
	Version:   "2026-07-28",
	ServerURL: "https://api.example.com",
}.TranscribeGroupDocument(group)
```

GCP API Gateway documents:

```go
doc, err := gcpapigateway.Transcriber{
	Version:        "2026-07-28",
	Host:           "api.example.gateway.dev",
	BackendAddress: "https://tasks.example.run.app",
}.TranscribeGroupDocument(group)
```

Shared OpenAPI document shapes live in `openapi/spec`. Target-specific packages
should be added under `openapi/<target>` and should translate from endpoint and
response metadata rather than adding provider fields to endpoint structs.

## Best Practices

- Define callers once and reuse the values everywhere.
- Put endpoint requirements on `EndpointSpec`, not on `request.Req`.
- Parse request bodies once with `param`; do not add a second validation phase.
- Return domain-ready params from parser boundaries.
- Treat unavailable restricted params as unexpected params.
- Use response shapes for public contracts and caller-sensitive attributes.
- Use `response.Project` when a domain value needs a response view.
- Keep service-specific auth, audit, storage, logging, and configuration behind
  the injected interfaces.
- Prefer provider-neutral route metadata; let transcribers speak provider
  dialects.
- Make timeout, request-size limit, and priority choices explicit on endpoints
  when they matter.

## Package Boundaries

The root package is doc-only. Application code should import subpackages:

- `github.com/zebodotdev/httpapi/auth`
- `github.com/zebodotdev/httpapi/caller`
- `github.com/zebodotdev/httpapi/endpoint`
- `github.com/zebodotdev/httpapi/erreur`
- `github.com/zebodotdev/httpapi/param`
- `github.com/zebodotdev/httpapi/request`
- `github.com/zebodotdev/httpapi/response`
- `github.com/zebodotdev/httpapi/openapi/openapi31`
- `github.com/zebodotdev/httpapi/openapi/gcpapigateway`
- `github.com/zebodotdev/httpapi/openapi/spec`
