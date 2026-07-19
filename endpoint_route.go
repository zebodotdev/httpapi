package httpapi

import endpointpkg "github.com/zebodotdev/httpapi/endpoint"

// RoutePathMode describes how a gateway should forward an incoming route path
// to its backend.
type RoutePathMode = endpointpkg.RoutePathMode

const (
	// RoutePathModeAppend forwards the matched request path to the backend.
	RoutePathModeAppend RoutePathMode = endpointpkg.RoutePathModeAppend
	// RoutePathModeConstant forwards to the backend address without appending the
	// matched request path.
	RoutePathModeConstant RoutePathMode = endpointpkg.RoutePathModeConstant
)

// RouteBackend identifies the upstream target for a transcribed endpoint.
type RouteBackend = endpointpkg.RouteBackend

// RouteSpec describes endpoint route metadata used by spec writers.
type RouteSpec = endpointpkg.RouteSpec

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

func normalizeRouteSpec(spec RouteSpec) RouteSpec {
	return endpointpkg.NormalizeRouteSpec(spec)
}

func normalizeRouteBackend(backend RouteBackend) RouteBackend {
	return endpointpkg.NormalizeRouteBackend(backend)
}

func normalizeRoutePathMode(mode RoutePathMode) RoutePathMode {
	return endpointpkg.NormalizeRoutePathMode(mode)
}
