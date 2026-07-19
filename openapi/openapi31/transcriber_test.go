package openapi31

import (
	"encoding/json"
	"strings"
	"testing"

	httpapi "github.com/zebodotdev/httpapi"
)

func TestTranscribeSkipsInternalRoutesAndEmitsMetadata(t *testing.T) {
	group := httpapi.EndpointGroup{PathPrefix: "/ops"}
	group.Add(httpapi.NewEndpoint(
		httpapi.POST,
		"/public",
		noopOpenAPI31Handler,
		httpapi.WithRequiredAuthorization(httpapi.AuthorizationKindBearer),
		httpapi.WithPriority(httpapi.EndpointPriorityHigh),
		httpapi.WithRouteSpec(httpapi.RouteSpec{
			OperationID: "public_operation",
			Summary:     "Public operation",
			Backend: httpapi.RouteBackend{
				Address: "https://service.example.internal",
			},
		}),
	))
	group.Add(httpapi.NewEndpoint(
		httpapi.POST,
		"/internal",
		noopOpenAPI31Handler,
		httpapi.WithInternal(),
	))

	routes, err := httpapi.RoutesFromGroup(group)
	if err != nil {
		t.Fatalf("RoutesFromGroup() error = %v", err)
	}

	paths, err := Transcriber{}.Transcribe(routes)
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
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
	if operation.XHTTPAPIAuthorization == nil ||
		operation.XHTTPAPIAuthorization.Kind != httpapi.AuthorizationKindBearer {
		t.Fatalf("authorization metadata = %#v", operation.XHTTPAPIAuthorization)
	}
	if operation.XHTTPAPIPriority != httpapi.EndpointPriorityHigh {
		t.Fatalf("priority = %q", operation.XHTTPAPIPriority)
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
	routes, err := httpapi.RoutesFromEndpoint(
		httpapi.NewEndpoint(httpapi.POST, "/orders/new", noopOpenAPI31Handler),
	)
	if err != nil {
		t.Fatalf("RoutesFromEndpoint() error = %v", err)
	}

	doc, err := Transcriber{
		Version:   "2026-07-18",
		ServerURL: "https://api.example.com",
	}.TranscribeDocument(routes)
	if err != nil {
		t.Fatalf("TranscribeDocument() error = %v", err)
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
	routes, err := httpapi.RoutesFromEndpoint(
		httpapi.NewEndpoint(httpapi.POST, "/orders/new", noopOpenAPI31Handler),
	)
	if err != nil {
		t.Fatalf("RoutesFromEndpoint() error = %v", err)
	}

	_, err = Transcriber{}.TranscribeDocument(routes)
	if err != ErrDocumentVersionRequired {
		t.Fatalf("error = %v, want ErrDocumentVersionRequired", err)
	}
}

func TestTranscribeWithPathPrefix(t *testing.T) {
	group := httpapi.EndpointGroup{PathPrefix: "orders"}
	group.Add(httpapi.NewEndpoint(httpapi.POST, "", noopOpenAPI31Handler))
	group.Add(httpapi.NewEndpoint(httpapi.GET, "/lookup", noopOpenAPI31Handler))

	routes, err := httpapi.RoutesFromGroup(group)
	if err != nil {
		t.Fatalf("RoutesFromGroup() error = %v", err)
	}

	paths, err := Transcriber{PathPrefix: "/v1"}.Transcribe(routes)
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}

	if paths["/v1/orders"].Post == nil {
		t.Fatal("post operation missing for /v1/orders")
	}
	if paths["/v1/orders/lookup"].Get == nil {
		t.Fatal("get operation missing for /v1/orders/lookup")
	}
}

func noopOpenAPI31Handler(*httpapi.Req) {}
