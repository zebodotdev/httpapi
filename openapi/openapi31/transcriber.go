package openapi31

import (
	"errors"
	"fmt"
	"strings"

	httpapi "github.com/zebodotdev/httpapi"
	internalroute "github.com/zebodotdev/httpapi/openapi/internal/route"
	"github.com/zebodotdev/httpapi/openapi/spec"
)

const (
	DocumentSpecVersion               = "3.1.1"
	DefaultDocumentTitle              = "http api"
	DefaultDocumentDescription        = "machine-readable representation of the http api"
	PlaceholderResponseDescription    = "Required placeholder response."
	HTTPAPIAuthorizationExtensionName = "x-httpapi-authorization"
	HTTPAPIPriorityExtensionName      = "x-httpapi-priority"
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

// TranscribeEndpoint emits public OpenAPI path entries for one endpoint.
func (t Transcriber) TranscribeEndpoint(endpoint httpapi.Endpoint) (spec.Paths, error) {
	routes, err := internalroute.FromEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	return t.transcribe(routes)
}

// TranscribeGroup emits public OpenAPI path entries for one endpoint group.
func (t Transcriber) TranscribeGroup(group httpapi.EndpointGroup) (spec.Paths, error) {
	routes, err := internalroute.FromGroup(group)
	if err != nil {
		return nil, err
	}

	return t.transcribe(routes)
}

// TranscribeGroups emits public OpenAPI path entries for several endpoint
// groups.
func (t Transcriber) TranscribeGroups(groups ...httpapi.EndpointGroup) (spec.Paths, error) {
	routes, err := internalroute.FromGroups(groups...)
	if err != nil {
		return nil, err
	}

	return t.transcribe(routes)
}

// TranscribeEndpointDocument emits a public OpenAPI 3.1 document for one endpoint.
func (t Transcriber) TranscribeEndpointDocument(endpoint httpapi.Endpoint) (spec.Document, error) {
	routes, err := internalroute.FromEndpoint(endpoint)
	if err != nil {
		return spec.Document{}, err
	}

	return t.transcribeDocument(routes)
}

// TranscribeGroupDocument emits a public OpenAPI 3.1 document for one endpoint group.
func (t Transcriber) TranscribeGroupDocument(group httpapi.EndpointGroup) (spec.Document, error) {
	routes, err := internalroute.FromGroup(group)
	if err != nil {
		return spec.Document{}, err
	}

	return t.transcribeDocument(routes)
}

// TranscribeGroupsDocument emits a public OpenAPI 3.1 document for several
// endpoint groups.
func (t Transcriber) TranscribeGroupsDocument(groups ...httpapi.EndpointGroup) (spec.Document, error) {
	routes, err := internalroute.FromGroups(groups...)
	if err != nil {
		return spec.Document{}, err
	}

	return t.transcribeDocument(routes)
}

func (t Transcriber) transcribe(routes internalroute.Routes) (spec.Paths, error) {
	routes, err := t.routes(routes)
	if err != nil {
		return nil, err
	}

	paths := spec.Paths{}
	for _, route := range routes {
		if route.Endpoint.IsInternal() {
			continue
		}
		operation, err := operationForRoute(route)
		if err != nil {
			return nil, err
		}
		if err := paths.AddOperation(route.Path, route.Method, operation); err != nil {
			return nil, err
		}
	}

	return paths, nil
}

func (t Transcriber) transcribeDocument(routes internalroute.Routes) (spec.Document, error) {
	if strings.TrimSpace(t.Version) == "" {
		return spec.Document{}, ErrDocumentVersionRequired
	}

	paths, err := t.transcribe(routes)
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

func (t Transcriber) routes(routes internalroute.Routes) (internalroute.Routes, error) {
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

func operationForRoute(route internalroute.Route) (spec.Operation, error) {
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
		if err := operation.SetExtension(HTTPAPIAuthorizationExtensionName, auth); err != nil {
			return spec.Operation{}, err
		}
	}
	if priority := route.Endpoint.Priority(); priority != "" {
		if err := operation.SetExtension(HTTPAPIPriorityExtensionName, priority); err != nil {
			return spec.Operation{}, err
		}
	}

	return operation, nil
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
