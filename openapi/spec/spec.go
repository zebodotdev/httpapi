package spec

import (
	"fmt"

	authpkg "github.com/zebodotdev/httpapi/auth"
	endpointpkg "github.com/zebodotdev/httpapi/endpoint"
)

// Document is the document-level schema produced from endpoints.
type Document struct {
	OpenAPI  string   `json:"openapi,omitempty" yaml:"openapi,omitempty"`
	Swagger  string   `json:"swagger,omitempty" yaml:"swagger,omitempty"`
	Info     Info     `json:"info" yaml:"info"`
	Servers  []Server `json:"servers,omitempty" yaml:"servers,omitempty"`
	Host     string   `json:"host,omitempty" yaml:"host,omitempty"`
	Schemes  []string `json:"schemes,omitempty" yaml:"schemes,omitempty"`
	Produces []string `json:"produces,omitempty" yaml:"produces,omitempty"`
	Paths    Paths    `json:"paths" yaml:"paths"`
}

// Info is the generated document info block.
type Info struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string `json:"version" yaml:"version"`
}

// Server is an OpenAPI 3 server entry.
type Server struct {
	URL string `json:"url" yaml:"url"`
}

// Paths maps OpenAPI path strings to method entries.
type Paths map[string]PathItem

// PathItem is the per-path OpenAPI method container.
type PathItem struct {
	Get  *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Post *Operation `json:"post,omitempty" yaml:"post,omitempty"`
}

// Operation is the minimal operation shape used by endpoint transcription.
type Operation struct {
	OperationID           string                            `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Summary               string                            `json:"summary,omitempty" yaml:"summary,omitempty"`
	Consumes              []string                          `json:"consumes,omitempty" yaml:"consumes,omitempty"`
	Produces              []string                          `json:"produces,omitempty" yaml:"produces,omitempty"`
	XGoogleBackend        *GCPGatewayBackend                `json:"x-google-backend,omitempty" yaml:"x-google-backend,omitempty"`
	XHTTPAPIInternal      bool                              `json:"x-httpapi-internal,omitempty" yaml:"x-httpapi-internal,omitempty"`
	XHTTPAPIAuthorization *authpkg.AuthorizationRequirement `json:"x-httpapi-authorization,omitempty" yaml:"x-httpapi-authorization,omitempty"`
	XHTTPAPIPriority      endpointpkg.Priority              `json:"x-httpapi-priority,omitempty" yaml:"x-httpapi-priority,omitempty"`
	Responses             map[string]Response               `json:"responses" yaml:"responses"`
}

// GCPGatewayBackend is the GCP API Gateway x-google-backend extension.
type GCPGatewayBackend struct {
	Address         string   `json:"address,omitempty" yaml:"address,omitempty"`
	PathTranslation string   `json:"path_translation,omitempty" yaml:"path_translation,omitempty"`
	Deadline        *float64 `json:"deadline,omitempty" yaml:"deadline,omitempty"`
}

// Response is the minimal OpenAPI response object.
type Response struct {
	Description string `json:"description" yaml:"description"`
}

// AddOperation adds an operation to the given path and method.
func (paths Paths) AddOperation(path string, method endpointpkg.Method, operation Operation) error {
	if paths == nil {
		return fmt.Errorf("openapi/spec: paths cannot be nil")
	}

	item := paths[path]
	if err := item.SetOperation(method, operation); err != nil {
		return fmt.Errorf("openapi/spec: add operation for %q: %w", path, err)
	}
	paths[path] = item

	return nil
}

// Merge copies path operations from other into paths.
func (paths Paths) Merge(other Paths) error {
	for path, item := range other {
		existing := paths[path]
		if err := existing.Merge(item); err != nil {
			return fmt.Errorf("openapi/spec: merge path %q: %w", path, err)
		}
		paths[path] = existing
	}

	return nil
}

// SetOperation sets one method operation on a path item.
func (p *PathItem) SetOperation(method endpointpkg.Method, operation Operation) error {
	switch method {
	case endpointpkg.GET:
		if p.Get != nil {
			return fmt.Errorf("duplicate %s operation", method)
		}
		p.Get = &operation
	case endpointpkg.POST:
		if p.Post != nil {
			return fmt.Errorf("duplicate %s operation", method)
		}
		p.Post = &operation
	default:
		return fmt.Errorf("unsupported openapi method %q", method)
	}

	return nil
}

// Merge adds all operations from other into p.
func (p *PathItem) Merge(other PathItem) error {
	if other.Get != nil {
		if p.Get != nil {
			return fmt.Errorf("duplicate %s operation", endpointpkg.GET)
		}
		p.Get = other.Get
	}
	if other.Post != nil {
		if p.Post != nil {
			return fmt.Errorf("duplicate %s operation", endpointpkg.POST)
		}
		p.Post = other.Post
	}

	return nil
}
