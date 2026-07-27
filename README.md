# HTTP API

`github.com/zebodotdev/httpapi` is a reusable HTTP contract layer for Go
services. It is intended to be safe for service teams to import as a standalone
package.

The package owns:

- endpoint definitions with `DefineEndpoint(EndpointSpec{...})`,
- endpoint groups and mounting,
- request parsing/audit via `github.com/zebodotdev/httpapi/request`,
- response rendering/writing via `github.com/zebodotdev/httpapi/response`,
- auth/caller/internal/priority/route/timeout metadata,
- endpoint access enforcement,
- idempotency orchestration,
- redacted request audit serialization,
- error response rendering,
- and provider-neutral metadata consumed by OpenAPI / gateway transcribers.

The package does not own service infrastructure. Importing services must provide
adapters for authentication, durable audit persistence, and idempotency storage.
There are no dependencies on product services, auth servers, databases, logging
systems, or service configuration packages.

## Endpoint Definitions

Define new endpoints with `EndpointSpec`:

```go
var TasksUI = caller.Define("tasks-ui")

createTask := endpoint.DefineEndpoint(endpoint.EndpointSpec{
	Method:  endpoint.POST,
	Path:    "/tasks/create",
	Handler: handler,
	Access: endpoint.EndpointAccessSpec{
		Authorization: endpoint.RequiredAuthorization(
			endpoint.AuthorizationKindService,
		),
		Callers: []caller.Caller{TasksUI},
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
	Timeout: endpoint.EndpointTimeoutSpec{
		ReadBody: 2 * time.Second,
		Handler:  10 * time.Second,
		Write:    2 * time.Second,
	},
	TimeoutHandler: func(req *endpoint.Req) {
		response.RenderJSON(req, http.StatusAccepted, map[string]string{
			"status": "queued_after_timeout",
		})
	},
	Priority: endpoint.EndpointPriorityHigh,
})
```

`NewEndpoint`, `NewIdempotentEndpoint`, and
`NewIdempotentEndpointWithScopeResolver` remain as compatibility constructors,
but new code should prefer the declarative spec.

Caller availability uses stable definitions from
`github.com/zebodotdev/httpapi/caller`, for example
`TasksUI := caller.Define("tasks-ui")`. Services attach the active caller with
`request.ContextWithCaller` or equivalent trusted middleware before mounting
endpoints.

`Timeout.ReadBody`, `Timeout.Handler`, and `Timeout.Write` are runtime budgets
for request body parsing, endpoint execution, and response writing. If the
handler budget expires before a response is produced, httpapi calls
`TimeoutHandler`. When `TimeoutHandler` is unset, httpapi renders the default
`request_timeout` error response and terminates the request.

## Service Wiring

Authentication is injected with `ConfigureAuthenticator`:

```go
restore := request.ConfigureAuthenticator(request.AuthenticatorFunc(
	func(ctx context.Context, req *request.Req, auth request.AuthenticationRequest) (*request.Session, error) {
		// Call the service's auth boundary here, then return a session.
		return &request.Session{
			ID:        "sess_...",
			App:       request.App{ID: "app_..."},
			AuthMode:  auth.Type,
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}, nil
	},
))
defer restore()
```

`NewReq` recognizes `Bearer` and `Service` authorization schemes by default.
Services can rename either scheme:

```go
restore := request.ConfigureAuthorizationSchemes(request.AuthorizationSchemes{
	Service: "Internal-Service",
})
defer restore()
```

Completed request audits are injected with `endpoint.ConfigureAuditSink`. The
default sink is a no-op.

Idempotency storage is injected with `endpoint.ConfigureIdempotencyStore`.
Idempotent endpoints fail closed with `idempotency_storage_unavailable` until a
store is configured. Services should also call
`endpoint.ConfigureIdempotencyScopeNamespace` with their service name so default
scopes do not collide across services.

## Transcription

Endpoint metadata is provider-neutral. Target-specific writers translate
`endpoint.RouteSpec` and `endpoint.RouteBackend` into their own document fields.
Pass endpoints or endpoint groups directly to the transcriber for the OpenAPI
target you need.

```go
doc, err := gcpapigateway.Transcriber{
	Version:        "2026-07-18",
	Host:           "api.example.gateway.dev",
	BackendAddress: "https://service.example.run.app",
}.TranscribeGroupDocument(group)
```

Public OpenAPI 3.1 generation uses
`github.com/zebodotdev/httpapi/openapi/openapi31`. GCP API Gateway generation
uses `github.com/zebodotdev/httpapi/openapi/gcpapigateway`. Shared document
shapes live in `github.com/zebodotdev/httpapi/openapi/spec`.

## Package Boundaries

The root package is doc-only. Import the package that owns the behavior you need:

- `endpoint` contains endpoint contract primitives such as method, content
  types, route metadata, priority, and timeout specs.
- `request` contains `Req`, request parsing, auth attachment, and audit-safe
  request serialization.
- `response` contains `Res`, render helpers, response encoding, streaming, and
  HTTP response writing.
- `caller` contains stable, provider-neutral request source definitions used by
  endpoint, request, and param availability.
- `param` contains request payload parsing, parameter null policy, size checks,
  relationship rules, and parameter caller availability.

## Extraction Boundary

Do not add imports from application-specific modules to this package.
Service-specific behavior belongs in adapters:

- Auth provider calls belong behind `Authenticator`.
- Database-backed audit writes belong behind `AuditSink`.
- Database-backed idempotency reservation belongs behind `IdempotencyStore`.
- Structured logging, tracing, and metrics belong in service middleware.
- Concrete service endpoint registries stay in their service modules.
