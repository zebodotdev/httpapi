package gcpapigateway

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	httpapi "github.com/zebodotdev/httpapi"
	"github.com/zebodotdev/httpapi/openapi/spec"
)

func TestTranscribeEmitsBackendAndDefaultResponse(t *testing.T) {
	routes, err := httpapi.RoutesFromEndpoint(httpapi.NewEndpoint(
		httpapi.POST,
		"/internal/sync",
		noopGCPGatewayHandler,
		httpapi.WithInternal(),
		httpapi.WithRequiredAuthorization(httpapi.AuthorizationKindService),
	))
	if err != nil {
		t.Fatalf("RoutesFromEndpoint() error = %v", err)
	}

	paths, err := Transcriber{
		BackendAddress: "https://service.example.internal",
	}.Transcribe(routes)
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}

	operation := paths["/internal/sync"].Post
	if operation == nil {
		t.Fatal("post operation missing")
	}
	backend := operationBackend(t, *operation)
	if backend.Address != "https://service.example.internal" {
		t.Fatalf("backend address = %q", backend.Address)
	}
	if backend.PathTranslation != PathTranslationAppend {
		t.Fatalf("path translation = %q", backend.PathTranslation)
	}
	if len(operation.Consumes) != 1 || operation.Consumes[0] != httpapi.ApplicationJson {
		t.Fatalf("consumes = %#v, want application/json", operation.Consumes)
	}
	if len(operation.Produces) != 1 || operation.Produces[0] != httpapi.ApplicationJson {
		t.Fatalf("produces = %#v, want application/json", operation.Produces)
	}
	if operation.Responses["default"].Description != PlaceholderResponseDescription {
		t.Fatalf("default response = %#v", operation.Responses["default"])
	}
	if !operation.XHTTPAPIInternal {
		t.Fatal("internal metadata missing from operation")
	}
	if operation.XHTTPAPIAuthorization == nil ||
		operation.XHTTPAPIAuthorization.Kind != httpapi.AuthorizationKindService {
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

func TestTranscribeUsesRouteSpecBackendAndGroupDefaults(t *testing.T) {
	group := httpapi.EndpointGroup{PathPrefix: "/orders"}
	group.Add(httpapi.NewEndpoint(httpapi.POST, "/new", noopGCPGatewayHandler))
	group.Add(httpapi.NewEndpoint(
		httpapi.POST,
		"/lookup",
		noopGCPGatewayHandler,
		httpapi.WithRouteSpec(httpapi.RouteSpec{
			OperationID: "lookup_order",
			Backend: httpapi.RouteBackend{
				Address:  "https://lookup.example.internal",
				PathMode: httpapi.RoutePathModeConstant,
				Timeout:  45 * time.Second,
			},
		}),
	))

	group.ConfigureRouteSpec(httpapi.RouteSpec{
		OperationID: "orders_group",
		Summary:     "Orders endpoint",
	})
	group.ConfigureRouteBackend(httpapi.RouteBackend{
		Address: "https://tasks.example.internal",
		Timeout: 20 * time.Second,
	})

	routes, err := httpapi.RoutesFromGroup(group)
	if err != nil {
		t.Fatalf("RoutesFromGroup() error = %v", err)
	}

	paths, err := Transcriber{
		BackendAddress: "https://fallback.example.internal",
	}.Transcribe(routes)
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
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
	createBackend := operationBackend(t, *createOperation)
	if createBackend.Address != "https://tasks.example.internal" {
		t.Fatalf("create backend = %#v", createBackend)
	}
	if createBackend.PathTranslation != PathTranslationAppend {
		t.Fatalf("create path translation = %q", createBackend.PathTranslation)
	}
	if createBackend.Deadline == nil ||
		*createBackend.Deadline != 20 {
		t.Fatalf("create deadline = %#v, want 20", createBackend.Deadline)
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
	lookupBackend := operationBackend(t, *lookupOperation)
	if lookupBackend.Address != "https://lookup.example.internal" {
		t.Fatalf("lookup backend = %#v", lookupBackend)
	}
	if lookupBackend.PathTranslation != PathTranslationConstant {
		t.Fatalf("lookup path translation = %q", lookupBackend.PathTranslation)
	}
	if lookupBackend.Deadline == nil ||
		*lookupBackend.Deadline != 45 {
		t.Fatalf("lookup deadline = %#v, want 45", lookupBackend.Deadline)
	}
}

func TestTranscribeRejectsBackendTimeoutAboveGatewayLimit(t *testing.T) {
	routes, err := httpapi.RoutesFromEndpoint(httpapi.NewEndpoint(
		httpapi.POST,
		"/exports/start",
		noopGCPGatewayHandler,
		httpapi.WithRouteBackend(httpapi.RouteBackend{
			Address: "https://service.example.internal",
			Timeout: BackendDeadlineMax + time.Second,
		}),
	))
	if err != nil {
		t.Fatalf("RoutesFromEndpoint() error = %v", err)
	}

	_, err = Transcriber{}.Transcribe(routes)
	if !errors.Is(err, ErrBackendDeadlineExceeded) {
		t.Fatalf("error = %v, want ErrBackendDeadlineExceeded", err)
	}
}

func TestTranscribeDocumentIncludesSwaggerShape(t *testing.T) {
	routes, err := httpapi.RoutesFromEndpoint(
		httpapi.NewEndpoint(httpapi.POST, "/orders/new", noopGCPGatewayHandler),
	)
	if err != nil {
		t.Fatalf("RoutesFromEndpoint() error = %v", err)
	}

	doc, err := Transcriber{
		Version:        "0.1.0-beta5",
		Host:           "api.example.gateway.dev",
		BackendAddress: "https://service.example.internal",
	}.TranscribeDocument(routes)
	if err != nil {
		t.Fatalf("TranscribeDocument() error = %v", err)
	}

	if doc.Swagger != DocumentSpecVersion {
		t.Fatalf("swagger = %q, want %q", doc.Swagger, DocumentSpecVersion)
	}
	if doc.Info.Version != "0.1.0-beta5" {
		t.Fatalf("info.version = %q", doc.Info.Version)
	}
	if doc.Host != "api.example.gateway.dev" {
		t.Fatalf("host = %q", doc.Host)
	}
	if len(doc.Schemes) != 1 || doc.Schemes[0] != DefaultScheme {
		t.Fatalf("schemes = %#v", doc.Schemes)
	}
	if len(doc.Produces) != 1 || doc.Produces[0] != httpapi.ApplicationJson {
		t.Fatalf("produces = %#v", doc.Produces)
	}
	if operationBackend(t, *doc.Paths["/orders/new"].Post) == (Backend{}) {
		t.Fatal("gateway backend missing from document operation")
	}
}

func TestTranscribeDocumentRequiresVersion(t *testing.T) {
	routes, err := httpapi.RoutesFromEndpoint(
		httpapi.NewEndpoint(httpapi.POST, "/orders/new", noopGCPGatewayHandler),
	)
	if err != nil {
		t.Fatalf("RoutesFromEndpoint() error = %v", err)
	}

	_, err = Transcriber{BackendAddress: "https://service.example.internal"}.TranscribeDocument(routes)
	if err != ErrDocumentVersionRequired {
		t.Fatalf("error = %v, want ErrDocumentVersionRequired", err)
	}
}

func noopGCPGatewayHandler(*httpapi.Req) {}

func operationBackend(t *testing.T, operation spec.Operation) Backend {
	t.Helper()

	value, ok := operation.Extension(BackendExtensionName)
	if !ok {
		t.Fatal("x-google-backend missing")
	}
	backend, ok := value.(Backend)
	if !ok {
		t.Fatalf("x-google-backend type = %T, want Backend", value)
	}

	return backend
}
