package httpapi

import (
	"fmt"
	"strings"
	"time"
)

// RoutePathMode describes how a gateway should forward an incoming route path
// to its backend.
type RoutePathMode string

const (
	// RoutePathModeAppend forwards the matched request path to the backend.
	RoutePathModeAppend RoutePathMode = "append"
	// RoutePathModeConstant forwards to the backend address without appending the
	// matched request path.
	RoutePathModeConstant RoutePathMode = "constant"
)

// RouteBackend identifies the upstream target for a transcribed endpoint.
type RouteBackend struct {
	Address  string        `json:"address,omitempty" yaml:"address,omitempty"`
	PathMode RoutePathMode `json:"path_mode,omitempty" yaml:"path_mode,omitempty"`
	// Timeout is intentionally a Go duration. Spec writers translate it into
	// their target-specific backend timeout fields.
	Timeout time.Duration `json:"-" yaml:"-"`
}

// RouteSpec describes endpoint route metadata used by spec writers.
type RouteSpec struct {
	// OperationID identifies one endpoint operation. Group-level defaults do not
	// inherit this value because operation IDs must remain unique.
	OperationID string       `json:"operation_id,omitempty" yaml:"operation_id,omitempty"`
	Summary     string       `json:"summary,omitempty" yaml:"summary,omitempty"`
	Backend     RouteBackend `json:"backend" yaml:"backend,omitempty"`
}

// WithRouteSpec applies route metadata to an endpoint.
func WithRouteSpec(spec RouteSpec) EndpointOption {
	return func(e *Endpoint) {
		e.route = normalizeRouteSpec(spec)
	}
}

// WithRouteBackend applies backend route metadata to an endpoint.
func WithRouteBackend(backend RouteBackend) EndpointOption {
	return func(e *Endpoint) {
		spec := e.routeSpec()
		spec.Backend = normalizeRouteBackend(backend)
		e.route = spec
	}
}

// ConfigureRouteSpec sets default route metadata for endpoints in the group.
// Endpoint-level metadata overrides these defaults.
func (eg *EndpointGroup) ConfigureRouteSpec(spec RouteSpec) {
	eg.Route = normalizeRouteSpec(spec)
}

// ConfigureRouteBackend sets the default backend route for endpoints in the group.
func (eg *EndpointGroup) ConfigureRouteBackend(backend RouteBackend) {
	eg.Route.Backend = normalizeRouteBackend(backend)
	eg.Route = normalizeRouteSpec(eg.Route)
}

func (e Endpoint) routeSpec() RouteSpec {
	return normalizeRouteSpec(e.route)
}

func (spec RouteSpec) withDefaults(defaults RouteSpec) RouteSpec {
	defaults = normalizeRouteSpec(defaults)
	spec = normalizeRouteSpec(spec)

	merged := RouteSpec{OperationID: spec.OperationID}
	if spec.Summary != "" {
		merged.Summary = spec.Summary
	} else {
		merged.Summary = defaults.Summary
	}
	merged.Backend = spec.Backend.withDefaults(defaults.Backend)

	return normalizeRouteSpec(merged)
}

func (backend RouteBackend) withDefaults(defaults RouteBackend) RouteBackend {
	defaults = normalizeRouteBackend(defaults)
	backend = normalizeRouteBackend(backend)

	merged := defaults
	if backend.Address != "" {
		merged.Address = backend.Address
	}
	if backend.PathMode != "" {
		merged.PathMode = backend.PathMode
	}
	if backend.Timeout != 0 {
		merged.Timeout = backend.Timeout
	}

	return normalizeRouteBackend(merged)
}

func normalizeRouteSpec(spec RouteSpec) RouteSpec {
	spec.OperationID = strings.TrimSpace(spec.OperationID)
	spec.Summary = strings.TrimSpace(spec.Summary)
	spec.Backend = normalizeRouteBackend(spec.Backend)

	return spec
}

func normalizeRouteBackend(backend RouteBackend) RouteBackend {
	backend.Address = strings.TrimSpace(backend.Address)
	backend.PathMode = normalizeRoutePathMode(backend.PathMode)
	if backend.Timeout < 0 {
		panic(fmt.Sprintf("httpapi: route backend timeout cannot be negative: %s", backend.Timeout))
	}

	return backend
}

func normalizeRoutePathMode(mode RoutePathMode) RoutePathMode {
	switch strings.TrimSpace(strings.ToLower(string(mode))) {
	case "":
		return ""
	case string(RoutePathModeAppend):
		return RoutePathModeAppend
	case string(RoutePathModeConstant):
		return RoutePathModeConstant
	default:
		panic(fmt.Sprintf("httpapi: unsupported route path mode %q", mode))
	}
}
