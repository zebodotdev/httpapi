package endpoint

// WithRouteSpec applies provider-neutral routing/backend metadata to an
// endpoint.
func WithRouteSpec(spec RouteSpec) EndpointOption {
	return func(e *Endpoint) {
		e.route = normalizeRouteSpec(spec)
	}
}

// WithRouteBackend applies backend route metadata to an endpoint.
//
// This is a convenience option for callers that only need to customize the
// backend portion of RouteSpec.
func WithRouteBackend(backend RouteBackend) EndpointOption {
	return func(e *Endpoint) {
		spec := e.routeSpec()
		spec.Backend = normalizeRouteBackend(backend)
		e.route = spec
	}
}

// ConfigureRouteSpec sets default routing/backend metadata for endpoints in the
// group.
func (eg *EndpointGroup) ConfigureRouteSpec(spec RouteSpec) {
	eg.Route = normalizeRouteSpec(spec)
}

// ConfigureRouteBackend sets the default backend route for endpoints in the
// group.
func (eg *EndpointGroup) ConfigureRouteBackend(backend RouteBackend) {
	eg.Route.Backend = normalizeRouteBackend(backend)
	eg.Route = normalizeRouteSpec(eg.Route)
}

func (e Endpoint) routeSpec() RouteSpec {
	return normalizeRouteSpec(e.route)
}

func normalizeRouteSpec(spec RouteSpec) RouteSpec {
	return NormalizeRouteSpec(spec)
}

func normalizeRouteBackend(backend RouteBackend) RouteBackend {
	return NormalizeRouteBackend(backend)
}

func normalizeRoutePathMode(mode RoutePathMode) RoutePathMode {
	return NormalizeRoutePathMode(mode)
}
