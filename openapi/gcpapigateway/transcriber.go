package gcpapigateway

import (
	"errors"
	"fmt"
	"strings"
	"time"

	httpapi "github.com/zebodotdev/httpapi"
	endpointpkg "github.com/zebodotdev/httpapi/endpoint"
	"github.com/zebodotdev/httpapi/openapi/spec"
)

const (
	DocumentSpecVersion            = "2.0"
	DefaultDocumentTitle           = "http api"
	DefaultDocumentDescription     = "machine-readable representation of the http api"
	DefaultScheme                  = "https"
	PathTranslationAppend          = "APPEND_PATH_TO_ADDRESS"
	PathTranslationConstant        = "CONSTANT_ADDRESS"
	BackendDeadlineMax             = 600 * time.Second
	PlaceholderResponseDescription = "Required placeholder response."
)

var (
	ErrDocumentVersionRequired = errors.New(
		"gcpapigateway: document version is required",
	)
	ErrBackendAddressRequired = errors.New(
		"gcpapigateway: backend address is required",
	)
	ErrBackendDeadlineExceeded = errors.New(
		"gcpapigateway: backend deadline exceeds maximum",
	)
)

// Transcriber writes routes into a GCP API Gateway Swagger 2.0 shape.
type Transcriber struct {
	PathPrefix     string
	Title          string
	Description    string
	Version        string
	Host           string
	BackendAddress string
}

// Transcribe emits GCP API Gateway path entries for the supplied routes.
func (t Transcriber) Transcribe(routes httpapi.Routes) (spec.Paths, error) {
	routes, err := t.routes(routes)
	if err != nil {
		return nil, err
	}

	paths := spec.Paths{}
	for _, route := range routes {
		operation, err := t.operationForRoute(route)
		if err != nil {
			return nil, err
		}
		if err := paths.AddOperation(route.Path, route.Method, operation); err != nil {
			return nil, err
		}
	}

	return paths, nil
}

// TranscribeDocument emits a GCP API Gateway document for the supplied routes.
func (t Transcriber) TranscribeDocument(routes httpapi.Routes) (spec.Document, error) {
	if strings.TrimSpace(t.Version) == "" {
		return spec.Document{}, ErrDocumentVersionRequired
	}

	paths, err := t.Transcribe(routes)
	if err != nil {
		return spec.Document{}, err
	}

	return spec.Document{
		Swagger: DocumentSpecVersion,
		Info: spec.Info{
			Title:       t.documentTitle(),
			Description: t.documentDescription(),
			Version:     strings.TrimSpace(t.Version),
		},
		Host:     strings.TrimSpace(t.Host),
		Schemes:  []string{DefaultScheme},
		Produces: []string{endpointpkg.ApplicationJson},
		Paths:    paths,
	}, nil
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

func (t Transcriber) operationForRoute(route httpapi.Route) (spec.Operation, error) {
	routeSpec := route.Endpoint.RouteSpec()
	backend, err := t.gatewayBackend(routeSpec.Backend)
	if err != nil {
		return spec.Operation{}, err
	}

	operation := spec.Operation{
		OperationID:    defaultOperationID(route.Method, route.Path),
		Summary:        fmt.Sprintf("%s %s", route.Method, route.Path),
		Consumes:       contentTypesForOpenAPI(route.Endpoint.AcceptedContentTypes()),
		Produces:       []string{endpointpkg.ApplicationJson},
		XGoogleBackend: &backend,
		Responses:      placeholderResponses(),
	}
	if routeSpec.OperationID != "" {
		operation.OperationID = routeSpec.OperationID
	}
	if routeSpec.Summary != "" {
		operation.Summary = routeSpec.Summary
	}
	if route.Endpoint.IsInternal() {
		operation.XHTTPAPIInternal = true
	}
	if auth := route.Endpoint.Authorization(); auth.Required {
		operation.XHTTPAPIAuthorization = &auth
	}
	if priority := route.Endpoint.Priority(); priority != "" {
		operation.XHTTPAPIPriority = priority
	}

	return operation, nil
}

func (t Transcriber) gatewayBackend(backend endpointpkg.RouteBackend) (spec.GCPGatewayBackend, error) {
	backend = backend.WithDefaults(endpointpkg.RouteBackend{
		Address:  strings.TrimSpace(t.BackendAddress),
		PathMode: endpointpkg.RoutePathModeAppend,
	})
	backend = endpointpkg.NormalizeRouteBackend(backend)
	if backend.Address == "" {
		return spec.GCPGatewayBackend{}, ErrBackendAddressRequired
	}

	pathTranslation := PathTranslationAppend
	switch backend.PathMode {
	case "", endpointpkg.RoutePathModeAppend:
		pathTranslation = PathTranslationAppend
	case endpointpkg.RoutePathModeConstant:
		pathTranslation = PathTranslationConstant
	default:
		return spec.GCPGatewayBackend{}, fmt.Errorf(
			"gcpapigateway: unsupported route path mode %q",
			backend.PathMode,
		)
	}

	deadline, err := gatewayBackendDeadline(backend)
	if err != nil {
		return spec.GCPGatewayBackend{}, err
	}

	return spec.GCPGatewayBackend{
		Address:         backend.Address,
		PathTranslation: pathTranslation,
		Deadline:        deadline,
	}, nil
}

func gatewayBackendDeadline(backend endpointpkg.RouteBackend) (*float64, error) {
	if backend.Timeout == 0 {
		return nil, nil
	}
	if backend.Timeout > BackendDeadlineMax {
		return nil, fmt.Errorf(
			"%w: timeout=%s max=%s",
			ErrBackendDeadlineExceeded,
			backend.Timeout,
			BackendDeadlineMax,
		)
	}

	seconds := backend.Timeout.Seconds()
	return &seconds, nil
}

func contentTypesForOpenAPI(contentTypes []endpointpkg.ContentType) []string {
	contentTypes = endpointpkg.NormalizeContentTypeSlice(contentTypes)
	values := make([]string, 0, len(contentTypes))
	for _, contentType := range contentTypes {
		values = append(values, string(contentType))
	}
	return values
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
