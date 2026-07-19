package endpoint

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

func (spec RouteSpec) WithDefaults(defaults RouteSpec) RouteSpec {
	defaults = NormalizeRouteSpec(defaults)
	spec = NormalizeRouteSpec(spec)

	merged := RouteSpec{OperationID: spec.OperationID}
	if spec.Summary != "" {
		merged.Summary = spec.Summary
	} else {
		merged.Summary = defaults.Summary
	}
	merged.Backend = spec.Backend.WithDefaults(defaults.Backend)

	return NormalizeRouteSpec(merged)
}

func (backend RouteBackend) WithDefaults(defaults RouteBackend) RouteBackend {
	defaults = NormalizeRouteBackend(defaults)
	backend = NormalizeRouteBackend(backend)

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

	return NormalizeRouteBackend(merged)
}

func NormalizeRouteSpec(spec RouteSpec) RouteSpec {
	spec.OperationID = strings.TrimSpace(spec.OperationID)
	spec.Summary = strings.TrimSpace(spec.Summary)
	spec.Backend = NormalizeRouteBackend(spec.Backend)

	return spec
}

func NormalizeRouteBackend(backend RouteBackend) RouteBackend {
	backend.Address = strings.TrimSpace(backend.Address)
	backend.PathMode = NormalizeRoutePathMode(backend.PathMode)
	if backend.Timeout < 0 {
		panic(fmt.Sprintf("httpapi: route backend timeout cannot be negative: %s", backend.Timeout))
	}

	return backend
}

func NormalizeRoutePathMode(mode RoutePathMode) RoutePathMode {
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
