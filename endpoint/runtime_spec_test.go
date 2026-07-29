package endpoint

import (
	"net/http"
	"testing"
	"time"

	e "github.com/zebodotdev/httpapi/erreur"
	"github.com/zebodotdev/httpapi/param"
	"github.com/zebodotdev/httpapi/response"
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
		Limits: EndpointLimitsSpec{
			MaxRequestBytes: 1024,
		},
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
	if endpoint.LimitsSpec().MaxRequestBytes != 1024 {
		t.Fatalf("max request bytes = %d, want 1024", endpoint.LimitsSpec().MaxRequestBytes)
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
	if endpoint.LimitsSpec() != (EndpointLimitsSpec{}) {
		t.Fatalf("limits = %#v, want zero value", endpoint.LimitsSpec())
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

func TestDefineEndpointStoresContractMetadata(t *testing.T) {
	request := RequestBody(contractRequestParser())
	responseContract := ResponseBody(
		http.StatusCreated,
		"Created order.",
		contractResponseShape(),
	)

	endpoint := DefineEndpoint(EndpointSpec{
		Method:    POST,
		Path:      "/orders/new",
		Handler:   noopTranscriptionHandler,
		Request:   request,
		Responses: []ResponseContract{responseContract},
	})

	gotRequest := endpoint.RequestContract()
	if !gotRequest.Required {
		t.Fatal("request contract is not required")
	}
	if gotRequest.Body.Type != param.TypeObject {
		t.Fatalf("request body type = %q, want object", gotRequest.Body.Type)
	}
	if len(gotRequest.Body.Parameters) != 2 {
		t.Fatalf("request parameters = %d, want 2", len(gotRequest.Body.Parameters))
	}
	if gotRequest.Body.Parameters[0].Name != "order_id" {
		t.Fatalf("first request parameter = %q", gotRequest.Body.Parameters[0].Name)
	}
	if gotRequest.Body.Parameters[1].Shape.Type != param.TypeArray ||
		gotRequest.Body.Parameters[1].Shape.Item == nil ||
		gotRequest.Body.Parameters[1].Shape.Item.Type != param.TypeString {
		t.Fatalf("tags parameter shape = %#v", gotRequest.Body.Parameters[1].Shape)
	}

	gotResponses := endpoint.ResponseContracts()
	if len(gotResponses) != 1 {
		t.Fatalf("response contracts = %d, want 1", len(gotResponses))
	}
	if gotResponses[0].Status != http.StatusCreated {
		t.Fatalf("response status = %d, want %d", gotResponses[0].Status, http.StatusCreated)
	}
	if gotResponses[0].Description != "Created order." {
		t.Fatalf("response description = %q", gotResponses[0].Description)
	}
	if gotResponses[0].ContentType != ApplicationJson {
		t.Fatalf("response content type = %q, want application/json", gotResponses[0].ContentType)
	}
	if gotResponses[0].Body.Type != response.TypeObject {
		t.Fatalf("response body type = %q, want object", gotResponses[0].Body.Type)
	}
	if len(gotResponses[0].Body.Attributes) != 2 {
		t.Fatalf("response attributes = %d, want 2", len(gotResponses[0].Body.Attributes))
	}
}

func TestDefineEndpointContractAccessorsDoNotLeakMutableState(t *testing.T) {
	minSize := int64(2)
	request := RequestContract{
		Required: true,
		Body: param.ShapeSpec{
			Type: param.TypeObject,
			Parameters: []param.ParameterSpec{
				{
					Name:     "order_id",
					Required: true,
					Shape:    param.ShapeSpec{Type: param.TypeString},
					MinSize:  &minSize,
				},
			},
			Rules: []param.RuleSpec{
				{Names: []string{"order_id"}, MinPresent: 1},
			},
		},
	}
	responses := []ResponseContract{
		{
			Status:      http.StatusOK,
			Description: "Order.",
			Body: response.ShapeSpec{
				Type: response.TypeObject,
				Attributes: []response.AttributeSpec{
					{
						Name:     "id",
						Required: true,
						Shape:    response.ShapeSpec{Type: response.TypeString},
					},
				},
			},
		},
	}

	endpoint := DefineEndpoint(EndpointSpec{
		Method:    POST,
		Handler:   noopTranscriptionHandler,
		Request:   request,
		Responses: responses,
	})

	request.Body.Parameters[0].Name = "mutated"
	request.Body.Rules[0].Names[0] = "mutated"
	*request.Body.Parameters[0].MinSize = 99
	responses[0].Description = "mutated"
	responses[0].Body.Attributes[0].Name = "mutated"

	gotRequest := endpoint.RequestContract()
	gotResponses := endpoint.ResponseContracts()
	gotRequest.Body.Parameters[0].Name = "mutated"
	gotRequest.Body.Rules[0].Names[0] = "mutated"
	*gotRequest.Body.Parameters[0].MinSize = 77
	gotResponses[0].Description = "mutated"
	gotResponses[0].Body.Attributes[0].Name = "mutated"

	gotRequest = endpoint.RequestContract()
	if gotRequest.Body.Parameters[0].Name != "order_id" {
		t.Fatalf("request parameter leaked mutation: %#v", gotRequest.Body.Parameters[0])
	}
	if gotRequest.Body.Rules[0].Names[0] != "order_id" {
		t.Fatalf("request rule leaked mutation: %#v", gotRequest.Body.Rules[0])
	}
	if gotRequest.Body.Parameters[0].MinSize == nil ||
		*gotRequest.Body.Parameters[0].MinSize != 2 {
		t.Fatalf("request min size leaked mutation: %#v", gotRequest.Body.Parameters[0].MinSize)
	}

	gotResponses = endpoint.ResponseContracts()
	if gotResponses[0].Description != "Order." {
		t.Fatalf("response description leaked mutation: %#v", gotResponses[0])
	}
	if gotResponses[0].Body.Attributes[0].Name != "id" {
		t.Fatalf("response attribute leaked mutation: %#v", gotResponses[0].Body.Attributes[0])
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

type contractRequest struct {
	OrderID string
	Tags    []string
}

type contractResponse struct {
	ID     string
	Status string
}

func contractRequestParser() *param.Request[contractRequest] {
	return param.JSON[contractRequest]().
		Param(param.Required("order_id", param.String())).
		Param(param.Optional("tags", param.ArrayOf(param.String()))).
		Parse(func(values param.Values) (contractRequest, error) {
			tags, _ := param.Get[[]string](values, "tags")
			return contractRequest{
				OrderID: param.Must[string](values, "order_id"),
				Tags:    tags,
			}, nil
		})
}

func contractResponseShape() response.Shape[contractResponse] {
	return response.Object[contractResponse](
		response.Required("id", response.String(), func(contract contractResponse) string {
			return contract.ID
		}),
		response.Required("status", response.String(), func(contract contractResponse) string {
			return contract.Status
		}),
	)
}
