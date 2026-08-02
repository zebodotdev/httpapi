package openapi31

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	endpointpkg "github.com/zebodotdev/httpapi/endpoint"
	"github.com/zebodotdev/httpapi/openapi/spec"
	"github.com/zebodotdev/httpapi/param"
	"github.com/zebodotdev/httpapi/response"
)

func TestTranscribeSkipsInternalRoutesAndEmitsMetadata(t *testing.T) {
	group := endpointpkg.EndpointGroup{PathPrefix: "/ops"}
	group.Add(endpointpkg.NewEndpoint(
		endpointpkg.POST,
		"/public",
		noopOpenAPI31Handler,
		endpointpkg.WithRequiredAuthorization(endpointpkg.AuthorizationKindBearer),
		endpointpkg.WithPriority(endpointpkg.EndpointPriorityHigh),
		endpointpkg.WithOperationSpec(endpointpkg.OperationSpec{
			ID:      "public_operation",
			Summary: "Public operation",
		}),
		endpointpkg.WithRouteSpec(endpointpkg.RouteSpec{
			Backend: endpointpkg.RouteBackend{
				Address: "https://service.example.internal",
			},
		}),
	))
	group.Add(endpointpkg.NewEndpoint(
		endpointpkg.POST,
		"/internal",
		noopOpenAPI31Handler,
		endpointpkg.WithInternal(),
	))

	paths, err := Transcriber{}.TranscribeGroup(group)
	if err != nil {
		t.Fatalf("TranscribeGroup() error = %v", err)
	}

	if _, ok := paths["/ops/internal"]; ok {
		t.Fatal("internal route was emitted in public OpenAPI")
	}
	operation := paths["/ops/public"].Post
	if operation == nil {
		t.Fatal("post operation missing")
	}
	if operation.OperationID != "public_operation" {
		t.Fatalf("operation id = %q", operation.OperationID)
	}
	if operation.Summary != "Public operation" {
		t.Fatalf("summary = %q", operation.Summary)
	}
	authValue, ok := operation.Extension(HTTPAPIAuthorizationExtensionName)
	if !ok {
		t.Fatal("authorization metadata missing")
	}
	auth, ok := authValue.(endpointpkg.AuthorizationRequirement)
	if !ok || auth.Kind != endpointpkg.AuthorizationKindBearer {
		t.Fatalf("authorization metadata = %#v", authValue)
	}
	priority, ok := operation.Extension(HTTPAPIPriorityExtensionName)
	if !ok || priority != endpointpkg.EndpointPriorityHigh {
		t.Fatalf("priority = %#v", priority)
	}
	encoded, err := json.Marshal(paths)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}
	if strings.Contains(string(encoded), `"x-google-backend"`) {
		t.Fatalf("public openapi operation emitted x-google-backend: %s", encoded)
	}
	if strings.Contains(string(encoded), `"consumes"`) {
		t.Fatalf("public operation emitted consumes: %s", encoded)
	}
	if strings.Contains(string(encoded), `"produces"`) {
		t.Fatalf("public operation emitted produces: %s", encoded)
	}
}

func TestTranscribeDocumentIncludesVersionAndServer(t *testing.T) {
	doc, err := Transcriber{
		Info: spec.Info{
			Version: "2026-07-18",
		},
		Servers: []spec.Server{{URL: "https://api.example.com"}},
	}.TranscribeEndpointDocument(
		endpointpkg.NewEndpoint(endpointpkg.POST, "/orders/new", noopOpenAPI31Handler),
	)
	if err != nil {
		t.Fatalf("TranscribeEndpointDocument() error = %v", err)
	}

	if doc.OpenAPI != DocumentSpecVersion {
		t.Fatalf("openapi = %q, want %q", doc.OpenAPI, DocumentSpecVersion)
	}
	if doc.Info.Version != "2026-07-18" {
		t.Fatalf("info.version = %q", doc.Info.Version)
	}
	if len(doc.Servers) != 1 || doc.Servers[0].URL != "https://api.example.com" {
		t.Fatalf("servers = %#v", doc.Servers)
	}
	if _, ok := doc.Paths["/orders/new"]; !ok {
		t.Fatal("public path missing from document")
	}
}

func TestTranscribeDocumentRequiresVersion(t *testing.T) {
	_, err := Transcriber{}.TranscribeEndpointDocument(
		endpointpkg.NewEndpoint(endpointpkg.POST, "/orders/new", noopOpenAPI31Handler),
	)
	if err != ErrDocumentVersionRequired {
		t.Fatalf("error = %v, want ErrDocumentVersionRequired", err)
	}
}

func TestTranscribeUsesEndpointRequestAndResponseContracts(t *testing.T) {
	paths, err := Transcriber{}.TranscribeEndpoint(endpointpkg.DefineEndpoint(endpointpkg.EndpointSpec{
		Method:    endpointpkg.POST,
		Path:      "/orders/new",
		Handler:   noopOpenAPI31Handler,
		Request:   endpointpkg.RequestBody(openAPI31OrderRequestParser()),
		Responses: []endpointpkg.ResponseContract{endpointpkg.ResponseBody(http.StatusCreated, "Created order.", openAPI31OrderResponseShape())},
	}))
	if err != nil {
		t.Fatalf("TranscribeEndpoint() error = %v", err)
	}

	operation := paths["/orders/new"].Post
	if operation == nil {
		t.Fatal("post operation missing")
	}
	if operation.RequestBody == nil {
		t.Fatal("request body missing")
	}
	requestSchema := operation.RequestBody.Content[string(endpointpkg.ApplicationJson)].Schema
	if requestSchema == nil {
		t.Fatal("request schema missing")
	}
	if requestSchema.Properties["order_id"].Type != "string" {
		t.Fatalf("order_id schema = %#v", requestSchema.Properties["order_id"])
	}
	if len(requestSchema.Required) != 1 || requestSchema.Required[0] != "order_id" {
		t.Fatalf("request required = %#v", requestSchema.Required)
	}

	created := operation.Responses["201"]
	if created.Description != "Created order." {
		t.Fatalf("created response = %#v", created)
	}
	responseSchema := created.Content[string(endpointpkg.ApplicationJson)].Schema
	if responseSchema == nil {
		t.Fatal("response schema missing")
	}
	if responseSchema.Properties["id"].Type != "string" {
		t.Fatalf("id response schema = %#v", responseSchema.Properties["id"])
	}
	if responseSchema.Properties["created_at"].Format != "date-time" {
		t.Fatalf("created_at response schema = %#v", responseSchema.Properties["created_at"])
	}
}

func TestTranscribeWithPathPrefix(t *testing.T) {
	group := endpointpkg.EndpointGroup{PathPrefix: "orders"}
	group.Add(endpointpkg.NewEndpoint(endpointpkg.POST, "", noopOpenAPI31Handler))
	group.Add(endpointpkg.NewEndpoint(endpointpkg.GET, "/lookup", noopOpenAPI31Handler))

	paths, err := Transcriber{PathPrefix: "/v1"}.TranscribeGroup(group)
	if err != nil {
		t.Fatalf("TranscribeGroup() error = %v", err)
	}

	if paths["/v1/orders"].Post == nil {
		t.Fatal("post operation missing for /v1/orders")
	}
	if paths["/v1/orders/lookup"].Get == nil {
		t.Fatal("get operation missing for /v1/orders/lookup")
	}
}

func noopOpenAPI31Handler(*endpointpkg.Req) {}

type openAPI31OrderRequest struct {
	OrderID string
}

type openAPI31OrderResponse struct {
	ID string
}

func openAPI31OrderRequestParser() *param.Request[openAPI31OrderRequest] {
	return param.JSON[openAPI31OrderRequest]().
		Param(param.Required("order_id", param.String())).
		Parse(func(values param.Values) (openAPI31OrderRequest, error) {
			return openAPI31OrderRequest{
				OrderID: param.Must[string](values, "order_id"),
			}, nil
		})
}

func openAPI31OrderResponseShape() response.Shape[openAPI31OrderResponse] {
	return response.Object[openAPI31OrderResponse](
		response.Required("id", response.String(), func(order openAPI31OrderResponse) string {
			return order.ID
		}),
		response.Required("created_at", response.Time(), func(openAPI31OrderResponse) time.Time {
			return time.Time{}
		}),
	)
}
