package httpapi

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestEndpointMetadataOptions(t *testing.T) {
	endpoint := NewIdempotentEndpoint(
		POST,
		"/internal/sync",
		noopTranscriptionHandler,
		WithInternal(),
		WithRequiredAuthorization(AuthorizationKindService),
	)

	if !endpoint.IsInternal() {
		t.Fatal("endpoint is not internal")
	}
	if !endpoint.IsIdempotent() {
		t.Fatal("endpoint is not idempotent")
	}
	auth := endpoint.Authorization()
	if !auth.Required {
		t.Fatal("endpoint authorization is not required")
	}
	if auth.Kind != AuthorizationKindService {
		t.Fatalf("authorization kind = %q, want %q", auth.Kind, AuthorizationKindService)
	}
}

func TestEndpointRouteSpecOptions(t *testing.T) {
	endpoint := NewEndpoint(
		POST,
		"/orders/new",
		noopTranscriptionHandler,
		WithRouteSpec(RouteSpec{
			OperationID: "create_order",
			Summary:     "Create order",
		}),
		WithRouteBackend(RouteBackend{
			Address:  " https://service.example.internal ",
			PathMode: RoutePathModeConstant,
			Timeout:  30 * time.Second,
		}),
	)

	route := endpoint.RouteSpec()
	if route.OperationID != "create_order" {
		t.Fatalf("operation id = %q", route.OperationID)
	}
	if route.Summary != "Create order" {
		t.Fatalf("summary = %q", route.Summary)
	}
	if route.Backend.Address != "https://service.example.internal" {
		t.Fatalf("backend address = %q", route.Backend.Address)
	}
	if route.Backend.PathMode != RoutePathModeConstant {
		t.Fatalf("backend path mode = %q", route.Backend.PathMode)
	}
	if route.Backend.Timeout != 30*time.Second {
		t.Fatalf("backend timeout = %s", route.Backend.Timeout)
	}
}

func TestEndpointGroupPropagatesInternalAndAuthorization(t *testing.T) {
	group := EndpointGroup{
		PathPrefix: "/ops",
		Internal:   true,
	}
	group.RequireAuthorization(AuthorizationKindService)

	group.Add(NewEndpoint(POST, "/sync", noopTranscriptionHandler))
	group.Add(NewEndpoint(
		POST,
		"/bearer",
		noopTranscriptionHandler,
		WithRequiredAuthorization(AuthorizationKindBearer),
	))

	first := group.Endpoints[0]
	if !first.IsInternal() {
		t.Fatal("group did not mark endpoint internal")
	}
	if first.Authorization().Kind != AuthorizationKindService {
		t.Fatalf(
			"propagated authorization kind = %q, want %q",
			first.Authorization().Kind,
			AuthorizationKindService,
		)
	}

	second := group.Endpoints[1]
	if second.Authorization().Kind != AuthorizationKindBearer {
		t.Fatalf(
			"endpoint authorization kind = %q, want %q",
			second.Authorization().Kind,
			AuthorizationKindBearer,
		)
	}
}

func TestEndpointGroupMetadataSettersUpdateExistingEndpoints(t *testing.T) {
	group := EndpointGroup{PathPrefix: "/ops"}
	group.Add(NewEndpoint(POST, "/sync", noopTranscriptionHandler))
	group.Add(NewEndpoint(
		POST,
		"/bearer",
		noopTranscriptionHandler,
		WithRequiredAuthorization(AuthorizationKindBearer),
	))

	group.MarkInternal()
	group.RequireAuthorization(AuthorizationKindService)

	first := group.Endpoints[0]
	if !first.IsInternal() {
		t.Fatal("existing endpoint was not marked internal")
	}
	if first.Authorization().Kind != AuthorizationKindService {
		t.Fatalf(
			"existing endpoint authorization kind = %q, want %q",
			first.Authorization().Kind,
			AuthorizationKindService,
		)
	}

	second := group.Endpoints[1]
	if !second.IsInternal() {
		t.Fatal("existing endpoint with endpoint auth was not marked internal")
	}
	if second.Authorization().Kind != AuthorizationKindBearer {
		t.Fatalf(
			"endpoint authorization kind = %q, want %q",
			second.Authorization().Kind,
			AuthorizationKindBearer,
		)
	}
}

func TestEndpointGroupAuthorizationDefaultCanBeChanged(t *testing.T) {
	group := EndpointGroup{PathPrefix: "/ops"}
	group.RequireAuthorization(AuthorizationKindBearer)
	group.Add(NewEndpoint(POST, "/sync", noopTranscriptionHandler))
	group.Add(NewEndpoint(
		POST,
		"/bearer",
		noopTranscriptionHandler,
		WithRequiredAuthorization(AuthorizationKindBearer),
	))

	group.RequireAuthorization(AuthorizationKindService)

	first := group.Endpoints[0]
	if first.Authorization().Kind != AuthorizationKindService {
		t.Fatalf(
			"inherited authorization kind = %q, want %q",
			first.Authorization().Kind,
			AuthorizationKindService,
		)
	}

	second := group.Endpoints[1]
	if second.Authorization().Kind != AuthorizationKindBearer {
		t.Fatalf(
			"endpoint authorization kind = %q, want %q",
			second.Authorization().Kind,
			AuthorizationKindBearer,
		)
	}
}

func TestEndpointPriorityOptionsAndGroupDefaults(t *testing.T) {
	group := EndpointGroup{
		PathPrefix: "/ops",
		Priority:   " HIGH ",
	}
	group.Add(NewEndpoint(POST, "/sync", noopTranscriptionHandler))
	group.Add(NewEndpoint(
		POST,
		"/health",
		noopTranscriptionHandler,
		WithPriority(EndpointPriorityLow),
	))

	first := group.Endpoints[0]
	if first.Priority() != EndpointPriorityHigh {
		t.Fatalf("inherited priority = %q, want %q", first.Priority(), EndpointPriorityHigh)
	}
	second := group.Endpoints[1]
	if second.Priority() != EndpointPriorityLow {
		t.Fatalf("endpoint priority = %q, want %q", second.Priority(), EndpointPriorityLow)
	}

	group.SetPriority(EndpointPriorityCritical)
	if group.Endpoints[0].Priority() != EndpointPriorityCritical {
		t.Fatalf("changed inherited priority = %q", group.Endpoints[0].Priority())
	}
	if group.Endpoints[1].Priority() != EndpointPriorityLow {
		t.Fatalf("explicit endpoint priority changed to %q", group.Endpoints[1].Priority())
	}

	paths, err := group.TranscribePublicOpenAPI()
	if err != nil {
		t.Fatalf("TranscribePublicOpenAPI() error = %v", err)
	}
	if paths["/ops/sync"].Post.XHTTPAPIPriority != EndpointPriorityCritical {
		t.Fatalf("sync priority = %q", paths["/ops/sync"].Post.XHTTPAPIPriority)
	}
	if paths["/ops/health"].Post.XHTTPAPIPriority != EndpointPriorityLow {
		t.Fatalf("health priority = %q", paths["/ops/health"].Post.XHTTPAPIPriority)
	}
}

func TestEndpointGroupLiteralAuthorizationIsNormalized(t *testing.T) {
	group := EndpointGroup{
		PathPrefix: "/ops",
		Auth: AuthorizationRequirement{
			Required: true,
			Kind:     " SERVICE ",
		},
	}
	group.Add(NewEndpoint(POST, "/sync", noopTranscriptionHandler))

	auth := group.Endpoints[0].Authorization()
	if auth.Kind != AuthorizationKindService {
		t.Fatalf("endpoint authorization kind = %q, want %q", auth.Kind, AuthorizationKindService)
	}

	paths, err := group.TranscribePublicOpenAPI()
	if err != nil {
		t.Fatalf("TranscribePublicOpenAPI() error = %v", err)
	}
	operation := paths["/ops/sync"].Post
	if operation.XHTTPAPIAuthorization == nil ||
		operation.XHTTPAPIAuthorization.Kind != AuthorizationKindService {
		t.Fatalf("authorization metadata = %#v", operation.XHTTPAPIAuthorization)
	}
}

func TestPublicOpenAPIRejectsInternalEndpoint(t *testing.T) {
	endpoint := NewEndpoint(
		POST,
		"/internal/sync",
		noopTranscriptionHandler,
		WithInternal(),
	)

	_, err := endpoint.TranscribePublicOpenAPI()
	if !errors.Is(err, ErrInternalEndpointPublicOpenAPI) {
		t.Fatalf("error = %v, want ErrInternalEndpointPublicOpenAPI", err)
	}
}

func TestPublicOpenAPIExcludesInternalGroupEndpoints(t *testing.T) {
	group := EndpointGroup{PathPrefix: "/ops"}
	group.Add(NewEndpoint(POST, "/public", noopTranscriptionHandler))
	group.Add(NewEndpoint(
		POST,
		"/internal",
		noopTranscriptionHandler,
		WithInternal(),
	))

	paths, err := group.TranscribePublicOpenAPI()
	if err != nil {
		t.Fatalf("TranscribePublicOpenAPI() error = %v", err)
	}
	if _, ok := paths["/ops/public"]; !ok {
		t.Fatal("public endpoint path missing")
	}
	if _, ok := paths["/ops/internal"]; ok {
		t.Fatal("internal endpoint path was emitted")
	}
	if len(paths) != 1 {
		t.Fatalf("path count = %d, want 1", len(paths))
	}
}

func TestPublicOpenAPITranscriptionDoesNotEmitGatewayShape(t *testing.T) {
	endpoint := NewEndpoint(
		POST,
		"/orders/pay",
		noopTranscriptionHandler,
		WithRequiredAuthorization(AuthorizationKindBearer),
		WithRouteSpec(RouteSpec{
			OperationID: "pay_order",
			Summary:     "Pay order",
			Backend: RouteBackend{
				Address: "https://service.example.internal",
			},
		}),
	)

	paths, err := endpoint.TranscribePublicOpenAPI()
	if err != nil {
		t.Fatalf("TranscribePublicOpenAPI() error = %v", err)
	}

	operation := paths["/orders/pay"].Post
	if operation == nil {
		t.Fatal("post operation missing")
	}
	if operation.XGoogleBackend != nil {
		t.Fatal("public operation emitted x-google-backend")
	}
	if operation.OperationID != "pay_order" {
		t.Fatalf("operation id = %q", operation.OperationID)
	}
	if operation.Summary != "Pay order" {
		t.Fatalf("summary = %q", operation.Summary)
	}
	if operation.XHTTPAPIAuthorization == nil ||
		operation.XHTTPAPIAuthorization.Kind != AuthorizationKindBearer {
		t.Fatalf("authorization metadata = %#v", operation.XHTTPAPIAuthorization)
	}

	encoded, err := json.Marshal(paths)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}
	if strings.Contains(string(encoded), `"consumes"`) {
		t.Fatalf("public openapi operation emitted consumes: %s", encoded)
	}
	if strings.Contains(string(encoded), `"produces"`) {
		t.Fatalf("public openapi operation emitted produces: %s", encoded)
	}
}

func TestPublicOpenAPIDocumentIncludesVersionAndSkipsInternalEndpoints(t *testing.T) {
	group := EndpointGroup{PathPrefix: "/orders"}
	group.Add(NewEndpoint(POST, "/new", noopTranscriptionHandler))
	group.Add(NewEndpoint(POST, "/internal/sync", noopTranscriptionHandler, WithInternal()))

	doc, err := group.TranscribePublicOpenAPIDocument(
		WithOpenAPIVersion("2026-07-16"),
		WithOpenAPIServerURL("https://api.example.com"),
	)
	if err != nil {
		t.Fatalf("TranscribePublicOpenAPIDocument() error = %v", err)
	}

	if doc.OpenAPI != "3.1.1" {
		t.Fatalf("openapi = %q, want 3.1.1", doc.OpenAPI)
	}
	if doc.Info.Version != "2026-07-16" {
		t.Fatalf("info.version = %q", doc.Info.Version)
	}
	if len(doc.Servers) != 1 || doc.Servers[0].URL != "https://api.example.com" {
		t.Fatalf("servers = %#v", doc.Servers)
	}
	if _, ok := doc.Paths["/orders/new"]; !ok {
		t.Fatal("public path missing from document")
	}
	if _, ok := doc.Paths["/orders/internal/sync"]; ok {
		t.Fatal("internal path emitted in public document")
	}
}

func TestOpenAPIDocumentRequiresVersion(t *testing.T) {
	endpoint := NewEndpoint(POST, "/orders/new", noopTranscriptionHandler)

	_, err := endpoint.TranscribePublicOpenAPIDocument()
	if !errors.Is(err, ErrOpenAPIDocumentVersionRequired) {
		t.Fatalf("error = %v, want ErrOpenAPIDocumentVersionRequired", err)
	}
}

func TestGCPGatewayTranscriptionEmitsBackendAndDefaultResponse(t *testing.T) {
	endpoint := NewEndpoint(
		POST,
		"/internal/sync",
		noopTranscriptionHandler,
		WithInternal(),
		WithRequiredAuthorization(AuthorizationKindService),
	)

	paths, err := endpoint.TranscribeGCPGateway(
		WithGCPGatewayBackendAddress("https://service.example.internal"),
	)
	if err != nil {
		t.Fatalf("TranscribeGCPGateway() error = %v", err)
	}

	operation := paths["/internal/sync"].Post
	if operation == nil {
		t.Fatal("post operation missing")
	}
	if operation.XGoogleBackend == nil {
		t.Fatal("x-google-backend missing")
	}
	if operation.XGoogleBackend.Address != "https://service.example.internal" {
		t.Fatalf("backend address = %q", operation.XGoogleBackend.Address)
	}
	if operation.XGoogleBackend.PathTranslation != gcpGatewayPathTranslationAppend {
		t.Fatalf("path translation = %q", operation.XGoogleBackend.PathTranslation)
	}
	if len(operation.Consumes) != 1 || operation.Consumes[0] != ApplicationJson {
		t.Fatalf("consumes = %#v, want application/json", operation.Consumes)
	}
	if len(operation.Produces) != 1 || operation.Produces[0] != ApplicationJson {
		t.Fatalf("produces = %#v, want application/json", operation.Produces)
	}
	if operation.Responses["default"].Description != placeholderResponseDescription {
		t.Fatalf("default response = %#v", operation.Responses["default"])
	}
	if !operation.XHTTPAPIInternal {
		t.Fatal("internal metadata missing from operation")
	}
	if operation.XHTTPAPIAuthorization == nil ||
		operation.XHTTPAPIAuthorization.Kind != AuthorizationKindService {
		t.Fatalf("authorization metadata = %#v", operation.XHTTPAPIAuthorization)
	}

	encoded, err := json.Marshal(paths)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}
	if !strings.Contains(string(encoded), `"x-google-backend"`) {
		t.Fatalf("encoded gateway path does not contain x-google-backend: %s", encoded)
	}
}

func TestGCPGatewayTranscriptionUsesRouteSpecBackend(t *testing.T) {
	endpoint := NewEndpoint(
		POST,
		"/orders/new",
		noopTranscriptionHandler,
		WithRouteSpec(RouteSpec{
			OperationID: "create_order",
			Summary:     "Create order",
			Backend: RouteBackend{
				Address:  "https://tasks.example.internal",
				PathMode: RoutePathModeConstant,
				Timeout:  30 * time.Second,
			},
		}),
	)

	paths, err := endpoint.TranscribeGCPGateway()
	if err != nil {
		t.Fatalf("TranscribeGCPGateway() error = %v", err)
	}

	operation := paths["/orders/new"].Post
	if operation == nil {
		t.Fatal("post operation missing")
	}
	if operation.OperationID != "create_order" {
		t.Fatalf("operation id = %q", operation.OperationID)
	}
	if operation.Summary != "Create order" {
		t.Fatalf("summary = %q", operation.Summary)
	}
	if operation.XGoogleBackend == nil {
		t.Fatal("x-google-backend missing")
	}
	if operation.XGoogleBackend.Address != "https://tasks.example.internal" {
		t.Fatalf("backend address = %q", operation.XGoogleBackend.Address)
	}
	if operation.XGoogleBackend.PathTranslation != gcpGatewayPathTranslationConstant {
		t.Fatalf("path translation = %q", operation.XGoogleBackend.PathTranslation)
	}
	if operation.XGoogleBackend.Deadline == nil || *operation.XGoogleBackend.Deadline != 30 {
		t.Fatalf("deadline = %#v, want 30", operation.XGoogleBackend.Deadline)
	}
}

func TestEndpointGroupRouteSpecDefaultsAndEndpointOverrides(t *testing.T) {
	group := EndpointGroup{PathPrefix: "/orders"}
	group.Add(NewEndpoint(POST, "/new", noopTranscriptionHandler))
	group.Add(NewEndpoint(
		POST,
		"/lookup",
		noopTranscriptionHandler,
		WithRouteSpec(RouteSpec{
			OperationID: "lookup_order",
			Backend: RouteBackend{
				Address:  "https://lookup.example.internal",
				PathMode: RoutePathModeConstant,
				Timeout:  45 * time.Second,
			},
		}),
	))

	group.ConfigureRouteSpec(RouteSpec{
		OperationID: "orders_group",
		Summary:     "Orders endpoint",
	})
	group.ConfigureRouteBackend(RouteBackend{
		Address: "https://tasks.example.internal",
		Timeout: 20 * time.Second,
	})

	paths, err := group.TranscribeGCPGateway(
		WithGCPGatewayBackendAddress("https://fallback.example.internal"),
	)
	if err != nil {
		t.Fatalf("TranscribeGCPGateway() error = %v", err)
	}

	createOperation := paths["/orders/new"].Post
	if createOperation == nil {
		t.Fatal("create operation missing")
	}
	if createOperation.Summary != "Orders endpoint" {
		t.Fatalf("create summary = %q", createOperation.Summary)
	}
	if createOperation.OperationID == "orders_group" {
		t.Fatal("group operation id was inherited by endpoint")
	}
	if createOperation.XGoogleBackend == nil ||
		createOperation.XGoogleBackend.Address != "https://tasks.example.internal" {
		t.Fatalf("create backend = %#v", createOperation.XGoogleBackend)
	}
	if createOperation.XGoogleBackend.PathTranslation != gcpGatewayPathTranslationAppend {
		t.Fatalf("create path translation = %q", createOperation.XGoogleBackend.PathTranslation)
	}
	if createOperation.XGoogleBackend.Deadline == nil || *createOperation.XGoogleBackend.Deadline != 20 {
		t.Fatalf("create deadline = %#v, want 20", createOperation.XGoogleBackend.Deadline)
	}

	lookupOperation := paths["/orders/lookup"].Post
	if lookupOperation == nil {
		t.Fatal("lookup operation missing")
	}
	if lookupOperation.OperationID != "lookup_order" {
		t.Fatalf("lookup operation id = %q", lookupOperation.OperationID)
	}
	if lookupOperation.Summary != "Orders endpoint" {
		t.Fatalf("lookup summary = %q", lookupOperation.Summary)
	}
	if lookupOperation.XGoogleBackend == nil ||
		lookupOperation.XGoogleBackend.Address != "https://lookup.example.internal" {
		t.Fatalf("lookup backend = %#v", lookupOperation.XGoogleBackend)
	}
	if lookupOperation.XGoogleBackend.PathTranslation != gcpGatewayPathTranslationConstant {
		t.Fatalf("lookup path translation = %q", lookupOperation.XGoogleBackend.PathTranslation)
	}
	if lookupOperation.XGoogleBackend.Deadline == nil || *lookupOperation.XGoogleBackend.Deadline != 45 {
		t.Fatalf("lookup deadline = %#v, want 45", lookupOperation.XGoogleBackend.Deadline)
	}
}

func TestGCPGatewayTranscriptionRejectsBackendTimeoutAboveGatewayLimit(t *testing.T) {
	endpoint := NewEndpoint(
		POST,
		"/exports/start",
		noopTranscriptionHandler,
		WithRouteBackend(RouteBackend{
			Address: "https://service.example.internal",
			Timeout: gcpGatewayBackendDeadlineMax + time.Second,
		}),
	)

	_, err := endpoint.TranscribeGCPGateway()
	if !errors.Is(err, ErrGCPGatewayBackendDeadlineExceeded) {
		t.Fatalf("error = %v, want ErrGCPGatewayBackendDeadlineExceeded", err)
	}
}

func TestGCPGatewayDocumentIncludesSwaggerShape(t *testing.T) {
	endpoint := NewEndpoint(POST, "/orders/new", noopTranscriptionHandler)

	doc, err := endpoint.TranscribeGCPGatewayDocument(
		WithOpenAPIVersion("0.1.0-beta5"),
		WithGCPGatewayHost("api.example.gateway.dev"),
		WithGCPGatewayBackendAddress("https://service.example.internal"),
	)
	if err != nil {
		t.Fatalf("TranscribeGCPGatewayDocument() error = %v", err)
	}

	if doc.Swagger != "2.0" {
		t.Fatalf("swagger = %q, want 2.0", doc.Swagger)
	}
	if doc.Info.Version != "0.1.0-beta5" {
		t.Fatalf("info.version = %q", doc.Info.Version)
	}
	if doc.Host != "api.example.gateway.dev" {
		t.Fatalf("host = %q", doc.Host)
	}
	if len(doc.Schemes) != 1 || doc.Schemes[0] != "https" {
		t.Fatalf("schemes = %#v", doc.Schemes)
	}
	if len(doc.Produces) != 1 || doc.Produces[0] != ApplicationJson {
		t.Fatalf("produces = %#v", doc.Produces)
	}
	if doc.Paths["/orders/new"].Post.XGoogleBackend == nil {
		t.Fatal("gateway backend missing from document operation")
	}
}

func TestGatewayTranscriptionNeutralAliases(t *testing.T) {
	endpoint := NewEndpoint(POST, "/orders/new", noopTranscriptionHandler)

	doc, err := endpoint.TranscribeGatewayDocument(
		WithOpenAPIVersion("0.1.0-beta5"),
		WithGatewayHost("api.example.gateway.dev"),
		WithGatewayBackendAddress("https://service.example.internal"),
	)
	if err != nil {
		t.Fatalf("TranscribeGatewayDocument() error = %v", err)
	}

	if doc.Swagger != "2.0" {
		t.Fatalf("swagger = %q, want 2.0", doc.Swagger)
	}
	if doc.Host != "api.example.gateway.dev" {
		t.Fatalf("host = %q", doc.Host)
	}
	if doc.Paths["/orders/new"].Post.XGoogleBackend == nil {
		t.Fatal("gateway backend missing from document operation")
	}
}

func TestEndpointGroupTranscriptionPrefixesPaths(t *testing.T) {
	group := EndpointGroup{PathPrefix: "orders"}
	group.Add(NewEndpoint(POST, "", noopTranscriptionHandler))
	group.Add(NewEndpoint(POST, "pay", noopTranscriptionHandler))
	group.Add(NewEndpoint(GET, "/lookup", noopTranscriptionHandler))

	paths, err := group.TranscribePublicOpenAPI(
		WithOpenAPIPathPrefix("/v1"),
	)
	if err != nil {
		t.Fatalf("TranscribePublicOpenAPI() error = %v", err)
	}

	tests := []struct {
		path   string
		method HttpMethod
	}{
		{path: "/v1/orders", method: POST},
		{path: "/v1/orders/pay", method: POST},
		{path: "/v1/orders/lookup", method: GET},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			item, ok := paths[tt.path]
			if !ok {
				t.Fatalf("path missing: %s", tt.path)
			}

			switch tt.method {
			case GET:
				if item.Get == nil {
					t.Fatal("get operation missing")
				}
			case POST:
				if item.Post == nil {
					t.Fatal("post operation missing")
				}
			}
		})
	}
}

func noopTranscriptionHandler(*Req) {}
