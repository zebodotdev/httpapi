package endpoint

import (
	"testing"
	"time"

	e "github.com/zebodotdev/httpapi/erreur"
)

func TestDefineEndpointBuildsEndpointFromSpec(t *testing.T) {
	resolver := func(*Req) (string, *e.ErrInvalidParam) { return "orders:new", nil }
	authKeys := map[string]bool{"secret": true}

	endpoint := DefineEndpoint(EndpointSpec{
		Method:  " post ",
		Path:    "/orders/new",
		Handler: noopTranscriptionHandler,
		Accepts: " application/json ",
		Access: EndpointAccessSpec{
			Internal:      true,
			Authorization: RequiredAuthorization(" service "),
		},
		Idempotency: EndpointIdempotencySpec{
			ScopeResolver: resolver,
		},
		Route: RouteSpec{
			OperationID: " create_order ",
			Summary:     " Create order ",
			Backend: RouteBackend{
				Address:  " https://service.example.internal ",
				PathMode: " constant ",
				Timeout:  15 * time.Second,
			},
		},
		Priority: " high ",
		AuthKeys: authKeys,
	})
	authKeys["secret"] = false
	authKeys["new"] = true

	if endpoint.Method() != POST {
		t.Fatalf("method = %q, want %q", endpoint.Method(), POST)
	}
	if endpoint.Pattern() != "/orders/new" {
		t.Fatalf("pattern = %q", endpoint.Pattern())
	}
	if endpoint.Accepts() != ApplicationJson {
		t.Fatalf("accepts = %q, want %q", endpoint.Accepts(), ApplicationJson)
	}
	if !endpoint.IsInternal() {
		t.Fatal("endpoint is not internal")
	}
	if !endpoint.IsIdempotent() {
		t.Fatal("resolver did not mark endpoint idempotent")
	}
	if endpoint.resolver == nil {
		t.Fatal("idempotency resolver was not set")
	}
	auth := endpoint.Authorization()
	if !auth.Required {
		t.Fatal("authorization is not required")
	}
	if auth.Kind != AuthorizationKindService {
		t.Fatalf("authorization kind = %q, want %q", auth.Kind, AuthorizationKindService)
	}
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
	if route.Backend.Timeout != 15*time.Second {
		t.Fatalf("backend timeout = %s", route.Backend.Timeout)
	}
	if endpoint.Priority() != EndpointPriorityHigh {
		t.Fatalf("priority = %q, want %q", endpoint.Priority(), EndpointPriorityHigh)
	}
	if keys := endpoint.AuthKeys(); !keys["secret"] || keys["new"] {
		t.Fatalf("auth keys were not cloned: %#v", keys)
	}
	keys := endpoint.AuthKeys()
	keys["secret"] = false
	keys["new"] = true
	if keys := endpoint.AuthKeys(); !keys["secret"] || keys["new"] {
		t.Fatalf("auth keys accessor exposed internal map: %#v", keys)
	}
}

func TestDefineEndpointDefaults(t *testing.T) {
	endpoint := DefineEndpoint(EndpointSpec{
		Method:  GET,
		Handler: noopTranscriptionHandler,
	})

	if endpoint.Pattern() != "" {
		t.Fatalf("pattern = %q, want empty", endpoint.Pattern())
	}
	if endpoint.Accepts() != ApplicationJson {
		t.Fatalf("accepts = %q, want %q", endpoint.Accepts(), ApplicationJson)
	}
	if endpoint.RequiresAuthorization() {
		t.Fatal("endpoint unexpectedly requires authorization")
	}
	if endpoint.IsIdempotent() {
		t.Fatal("endpoint is unexpectedly idempotent")
	}
}

func TestDefineEndpointAcceptsMultipleContentTypes(t *testing.T) {
	endpoint := DefineEndpoint(EndpointSpec{
		Method: GET,
		AcceptsAny: []ContentType{
			TextPlain,
			ApplicationJson,
			TextPlain,
		},
		Handler: noopTranscriptionHandler,
	})

	if endpoint.Accepts() != TextPlain {
		t.Fatalf("primary accepts = %q, want %q", endpoint.Accepts(), TextPlain)
	}
	want := []ContentType{TextPlain, ApplicationJson}
	got := endpoint.AcceptedContentTypes()
	if len(got) != len(want) {
		t.Fatalf("accepted content types = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("accepted content types = %#v, want %#v", got, want)
		}
	}
	if err := validateEndpointContentType("text/plain; charset=utf-8", got); err != nil {
		t.Fatalf("validate text/plain: %v", err)
	}
}

func TestLegacyConstructorsMatchEndpointSpec(t *testing.T) {
	resolver := func(*Req) (string, *e.ErrInvalidParam) { return "sync", nil }
	route := RouteSpec{
		OperationID: "sync_internal",
		Summary:     "Sync internal state",
		Backend: RouteBackend{
			Address:  "https://service.example.internal",
			PathMode: RoutePathModeAppend,
			Timeout:  10 * time.Second,
		},
	}

	legacy := NewIdempotentEndpointWithScopeResolver(
		POST,
		"/internal/sync",
		resolver,
		noopTranscriptionHandler,
		WithInternal(),
		WithRequiredAuthorization(AuthorizationKindService),
		WithRouteSpec(route),
		WithPriority(EndpointPriorityCritical),
	)
	defined := DefineEndpoint(EndpointSpec{
		Method:  POST,
		Path:    "/internal/sync",
		Handler: noopTranscriptionHandler,
		Access: EndpointAccessSpec{
			Internal:      true,
			Authorization: RequiredAuthorization(AuthorizationKindService),
		},
		Idempotency: EndpointIdempotencySpec{
			Enabled:       true,
			ScopeResolver: resolver,
		},
		Route:    route,
		Priority: EndpointPriorityCritical,
	})

	if legacy.Method() != defined.Method() {
		t.Fatalf("method mismatch: %q != %q", legacy.Method(), defined.Method())
	}
	if legacy.Pattern() != defined.Pattern() {
		t.Fatalf("pattern mismatch: %q != %q", legacy.Pattern(), defined.Pattern())
	}
	if legacy.Accepts() != defined.Accepts() {
		t.Fatalf("accepts mismatch: %q != %q", legacy.Accepts(), defined.Accepts())
	}
	if legacy.IsInternal() != defined.IsInternal() {
		t.Fatalf("internal mismatch: %v != %v", legacy.IsInternal(), defined.IsInternal())
	}
	if legacy.IsIdempotent() != defined.IsIdempotent() {
		t.Fatalf("idempotent mismatch: %v != %v", legacy.IsIdempotent(), defined.IsIdempotent())
	}
	if legacy.Authorization() != defined.Authorization() {
		t.Fatalf("authorization mismatch: %#v != %#v", legacy.Authorization(), defined.Authorization())
	}
	if legacy.RouteSpec() != defined.RouteSpec() {
		t.Fatalf("route mismatch: %#v != %#v", legacy.RouteSpec(), defined.RouteSpec())
	}
	if legacy.Priority() != defined.Priority() {
		t.Fatalf("priority mismatch: %q != %q", legacy.Priority(), defined.Priority())
	}
}

func TestDefineEndpointRequiresHandler(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("DefineEndpoint did not panic")
		}
	}()

	DefineEndpoint(EndpointSpec{Method: POST})
}
