package openapi31

import (
	"encoding/json"
	"strings"
	"testing"

	endpointpkg "github.com/zebodotdev/httpapi/endpoint"
)

func TestTranscribeSkipsInternalRoutesAndEmitsMetadata(t *testing.T) {
	group := endpointpkg.EndpointGroup{PathPrefix: "/ops"}
	group.Add(endpointpkg.NewEndpoint(
		endpointpkg.POST,
		"/public",
		noopOpenAPI31Handler,
		endpointpkg.WithRequiredAuthorization(endpointpkg.AuthorizationKindBearer),
		endpointpkg.WithPriority(endpointpkg.EndpointPriorityHigh),
		endpointpkg.WithRouteSpec(endpointpkg.RouteSpec{
			OperationID: "public_operation",
			Summary:     "Public operation",
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
		Version:   "2026-07-18",
		ServerURL: "https://api.example.com",
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
