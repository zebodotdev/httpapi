package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidExtensionName = errors.New(
	"openapi/spec: extension name must begin with x-",
)

var ErrDuplicateExtensionName = errors.New(
	"openapi/spec: duplicate extension name",
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
	Get     *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Put     *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Post    *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Delete  *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options *Operation `json:"options,omitempty" yaml:"options,omitempty"`
	Head    *Operation `json:"head,omitempty" yaml:"head,omitempty"`
	Patch   *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace   *Operation `json:"trace,omitempty" yaml:"trace,omitempty"`
}

// Operation is the minimal operation shape used by endpoint transcription.
type Operation struct {
	OperationID string              `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Summary     string              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Consumes    []string            `json:"consumes,omitempty" yaml:"consumes,omitempty"`
	Produces    []string            `json:"produces,omitempty" yaml:"produces,omitempty"`
	Extensions  Extensions          `json:"-" yaml:"-"`
	Responses   map[string]Response `json:"responses" yaml:"responses"`
}

// Response is the minimal OpenAPI response object.
type Response struct {
	Description string `json:"description" yaml:"description"`
}

// Extensions contains OpenAPI Specification Extensions for an object.
type Extensions map[string]any

// SetExtension sets an OpenAPI Specification Extension on the operation.
func (operation *Operation) SetExtension(name string, value any) error {
	name = normalizeExtensionName(name)
	if err := validateExtensionName(name); err != nil {
		return err
	}

	if operation.Extensions == nil {
		operation.Extensions = Extensions{}
	}
	for existing := range operation.Extensions {
		if normalizeExtensionName(existing) != name {
			continue
		}
		delete(operation.Extensions, existing)
	}
	operation.Extensions[name] = value

	return nil
}

// Extension returns one OpenAPI Specification Extension value.
func (operation Operation) Extension(name string) (any, bool) {
	name = normalizeExtensionName(name)
	if err := validateExtensionName(name); err != nil {
		return nil, false
	}

	for existing, value := range operation.Extensions {
		if normalizeExtensionName(existing) == name {
			return value, true
		}
	}

	return nil, false
}

// MarshalJSON emits operation extensions inline, as required by OpenAPI.
func (operation Operation) MarshalJSON() ([]byte, error) {
	object, err := operation.object()
	if err != nil {
		return nil, err
	}

	return json.Marshal(object)
}

// MarshalYAML emits operation extensions inline, as required by OpenAPI.
func (operation Operation) MarshalYAML() (any, error) {
	return operation.object()
}

// AddOperation adds an operation to the given path and method.
func (paths Paths) AddOperation(path string, method string, operation Operation) error {
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
	if paths == nil {
		return fmt.Errorf("openapi/spec: paths cannot be nil")
	}

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
func (p *PathItem) SetOperation(method string, operation Operation) error {
	slot, err := p.operationSlot(method)
	if err != nil {
		return err
	}
	if *slot != nil {
		return fmt.Errorf("duplicate %s operation", normalizeMethod(method))
	}

	*slot = &operation
	return nil
}

// Merge adds all operations from other into p.
func (p *PathItem) Merge(other PathItem) error {
	operations := []struct {
		method    string
		operation *Operation
	}{
		{method: "GET", operation: other.Get},
		{method: "PUT", operation: other.Put},
		{method: "POST", operation: other.Post},
		{method: "DELETE", operation: other.Delete},
		{method: "OPTIONS", operation: other.Options},
		{method: "HEAD", operation: other.Head},
		{method: "PATCH", operation: other.Patch},
		{method: "TRACE", operation: other.Trace},
	}

	for _, operation := range operations {
		if operation.operation == nil {
			continue
		}
		if err := p.SetOperation(operation.method, *operation.operation); err != nil {
			return err
		}
	}

	return nil
}

func (operation Operation) object() (map[string]any, error) {
	object := map[string]any{
		"responses": operation.responses(),
	}
	if operation.OperationID != "" {
		object["operationId"] = operation.OperationID
	}
	if operation.Summary != "" {
		object["summary"] = operation.Summary
	}
	if len(operation.Consumes) > 0 {
		object["consumes"] = operation.Consumes
	}
	if len(operation.Produces) > 0 {
		object["produces"] = operation.Produces
	}

	extensions, err := normalizeExtensions(operation.Extensions)
	if err != nil {
		return nil, err
	}
	for name, value := range extensions {
		object[name] = value
	}

	return object, nil
}

func (operation Operation) responses() map[string]Response {
	if operation.Responses == nil {
		return map[string]Response{}
	}

	return operation.Responses
}

func (p *PathItem) operationSlot(method string) (**Operation, error) {
	if p == nil {
		return nil, fmt.Errorf("openapi/spec: path item cannot be nil")
	}

	switch normalizeMethod(method) {
	case "GET":
		return &p.Get, nil
	case "PUT":
		return &p.Put, nil
	case "POST":
		return &p.Post, nil
	case "DELETE":
		return &p.Delete, nil
	case "OPTIONS":
		return &p.Options, nil
	case "HEAD":
		return &p.Head, nil
	case "PATCH":
		return &p.Patch, nil
	case "TRACE":
		return &p.Trace, nil
	default:
		return nil, fmt.Errorf("unsupported openapi method %q", method)
	}
}

func normalizeMethod(method string) string {
	return strings.ToUpper(strings.TrimSpace(method))
}

func normalizeExtensions(extensions Extensions) (Extensions, error) {
	if len(extensions) == 0 {
		return nil, nil
	}

	normalized := make(Extensions, len(extensions))
	for rawName, value := range extensions {
		name := normalizeExtensionName(rawName)
		if err := validateExtensionName(name); err != nil {
			return nil, err
		}
		if _, exists := normalized[name]; exists {
			return nil, fmt.Errorf("%w: %q", ErrDuplicateExtensionName, name)
		}
		normalized[name] = value
	}

	return normalized, nil
}

func normalizeExtensionName(name string) string {
	return strings.TrimSpace(name)
}

func validateExtensionName(name string) error {
	if !strings.HasPrefix(name, "x-") || name == "x-" {
		return fmt.Errorf("%w: %q", ErrInvalidExtensionName, name)
	}

	return nil
}
