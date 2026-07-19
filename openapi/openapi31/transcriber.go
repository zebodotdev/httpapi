package openapi31

import (
	"errors"
	"fmt"
	"strings"

	httpapi "github.com/zebodotdev/httpapi"
	"github.com/zebodotdev/httpapi/openapi/spec"
)

const (
	DocumentSpecVersion            = "3.1.1"
	DefaultDocumentTitle           = "http api"
	DefaultDocumentDescription     = "machine-readable representation of the http api"
	PlaceholderResponseDescription = "Required placeholder response."
)

var ErrDocumentVersionRequired = errors.New(
	"openapi31: document version is required",
)

// Transcriber writes routes into a public OpenAPI 3.1 shape.
type Transcriber struct {
	PathPrefix  string
	Title       string
	Description string
	Version     string
	ServerURL   string
}

// Transcribe emits public OpenAPI path entries for the supplied routes.
func (t Transcriber) Transcribe(routes httpapi.Routes) (spec.Paths, error) {
	routes, err := t.routes(routes)
	if err != nil {
		return nil, err
	}

	paths := spec.Paths{}
	for _, route := range routes {
		if route.Endpoint.IsInternal() {
			continue
		}
		operation := operationForRoute(route)
		if err := paths.AddOperation(route.Path, route.Method, operation); err != nil {
			return nil, err
		}
	}

	return paths, nil
}

// TranscribeDocument emits a public OpenAPI 3.1 document for the supplied routes.
func (t Transcriber) TranscribeDocument(routes httpapi.Routes) (spec.Document, error) {
	if strings.TrimSpace(t.Version) == "" {
		return spec.Document{}, ErrDocumentVersionRequired
	}

	paths, err := t.Transcribe(routes)
	if err != nil {
		return spec.Document{}, err
	}

	doc := spec.Document{
		OpenAPI: DocumentSpecVersion,
		Info: spec.Info{
			Title:       t.documentTitle(),
			Description: t.documentDescription(),
			Version:     strings.TrimSpace(t.Version),
		},
		Paths: paths,
	}
	if serverURL := strings.TrimSpace(t.ServerURL); serverURL != "" {
		doc.Servers = []spec.Server{{URL: serverURL}}
	}

	return doc, nil
}

func (t Transcriber) routes(routes httpapi.Routes) (httpapi.Routes, error) {
	if strings.TrimSpace(t.PathPrefix) == "" {
		return routes, nil
	}

	return routes.WithPathPrefix(t.PathPrefix)
}

func (t Transcriber) documentTitle() string {
	title := strings.TrimSpace(t.Title)
	if title == "" {
		return DefaultDocumentTitle
	}
	return title
}

func (t Transcriber) documentDescription() string {
	description := strings.TrimSpace(t.Description)
	if description == "" {
		return DefaultDocumentDescription
	}
	return description
}

func operationForRoute(route httpapi.Route) spec.Operation {
	routeSpec := route.Endpoint.RouteSpec()
	operation := spec.Operation{
		OperationID: defaultOperationID(route.Method, route.Path),
		Summary:     fmt.Sprintf("%s %s", route.Method, route.Path),
		Responses:   placeholderResponses(),
	}
	if routeSpec.OperationID != "" {
		operation.OperationID = routeSpec.OperationID
	}
	if routeSpec.Summary != "" {
		operation.Summary = routeSpec.Summary
	}
	if auth := route.Endpoint.Authorization(); auth.Required {
		operation.XHTTPAPIAuthorization = &auth
	}
	if priority := route.Endpoint.Priority(); priority != "" {
		operation.XHTTPAPIPriority = priority
	}

	return operation
}

func placeholderResponses() map[string]spec.Response {
	return map[string]spec.Response{
		"default": {Description: PlaceholderResponseDescription},
	}
}

func defaultOperationID(method httpapi.HttpMethod, path string) string {
	parts := []string{strings.ToLower(method)}
	for _, part := range strings.FieldsFunc(path, operationIDSeparator) {
		part = strings.Trim(part, "{}")
		if part == "" {
			continue
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, "_")
}

func operationIDSeparator(r rune) bool {
	return r == '/' || r == '-' || r == '_' || r == '.' || r == ':'
}
