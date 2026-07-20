package endpoint

// EndpointSpec is the declarative contract for one HTTP endpoint.
//
// It should be the preferred way to define new endpoints. Endpoint requirements
// live here so Req can remain a safe parse of an incoming request, independent
// of the endpoint that eventually handles it.
type EndpointSpec struct {
	Method      HttpMethod
	Path        string
	Handler     Handler
	Accepts     ContentType
	AcceptsAny  []ContentType
	Access      EndpointAccessSpec
	Idempotency EndpointIdempotencySpec
	Route       RouteSpec
	Priority    EndpointPriority
	Timeout     EndpointTimeoutSpec
	// TimeoutHandler renders the response when the handler context reaches its
	// timeout before the endpoint produces a response. When unset, httpapi
	// renders DefaultEndpointTimeoutHandler.
	TimeoutHandler EndpointTimeoutHandler
	AuthKeys       map[string]bool
}

// EndpointAccessSpec declares endpoint access requirements.
type EndpointAccessSpec struct {
	Internal      bool
	Authorization AuthorizationRequirement
}

// EndpointIdempotencySpec declares endpoint idempotency behavior.
type EndpointIdempotencySpec struct {
	Enabled       bool
	ScopeResolver IdempotencyScopeResolver
}

// DefineEndpoint returns a server endpoint from a declarative endpoint spec.
func DefineEndpoint(spec EndpointSpec) Endpoint {
	return endpointFromSpec(spec).withRebuiltHandler()
}

func endpointFromSpec(spec EndpointSpec) Endpoint {
	if spec.Handler == nil {
		panic("httpapi: endpoint handler is required")
	}

	idempotent := spec.Idempotency.Enabled
	if spec.Idempotency.ScopeResolver != nil {
		idempotent = true
	}

	return Endpoint{
		method:     normalizeEndpointMethod(spec.Method),
		pattern:    spec.Path,
		accepts:    normalizeEndpointContentTypes(spec.Accepts, spec.AcceptsAny...),
		rawHandler: spec.Handler,
		idempotent: idempotent,
		resolver:   spec.Idempotency.ScopeResolver,
		access: endpointAccessPolicy{
			internal: spec.Access.Internal,
			auth:     normalizeAuthorizationRequirement(spec.Access.Authorization),
		},
		route: normalizeRouteSpec(spec.Route),
		priority: endpointPriorityPolicy{
			priority: normalizeEndpointPriority(spec.Priority),
		},
		timeout: endpointTimeoutPolicy{
			timeout: normalizeEndpointTimeoutSpec(spec.Timeout),
			handler: spec.TimeoutHandler,
		},
		authKeys: cloneEndpointAuthKeys(spec.AuthKeys),
	}
}

func defineEndpointWithOptions(spec EndpointSpec, opts ...EndpointOption) Endpoint {
	endpoint := endpointFromSpec(spec)
	for _, opt := range opts {
		opt(&endpoint)
	}

	return endpoint.withRebuiltHandler()
}

func normalizeEndpointMethod(method HttpMethod) HttpMethod {
	return NormalizeMethod(method)
}

func normalizeEndpointContentType(contentType ContentType) ContentType {
	return NormalizeContentType(contentType)
}

func normalizeEndpointContentTypes(primary ContentType, additional ...ContentType) []ContentType {
	return NormalizeContentTypes(primary, additional...)
}

func normalizeEndpointContentTypeSlice(contentTypes []ContentType) []ContentType {
	return NormalizeContentTypeSlice(contentTypes)
}

func primaryContentType(contentTypes []ContentType) ContentType {
	return PrimaryContentType(contentTypes)
}

func cloneContentTypes(contentTypes []ContentType) []ContentType {
	return CloneContentTypes(contentTypes)
}

func joinContentTypes(contentTypes []ContentType) string {
	return JoinContentTypes(contentTypes)
}

// WithAcceptedContentTypes sets every accepted request content type.
func WithAcceptedContentTypes(contentTypes ...ContentType) EndpointOption {
	contentTypes = normalizeEndpointContentTypes("", contentTypes...)
	return func(e *Endpoint) {
		e.accepts = cloneContentTypes(contentTypes)
	}
}

func cloneEndpointAuthKeys(keys map[string]bool) map[string]bool {
	return CloneAuthKeys(keys)
}

func validateEndpointContentType(actual ContentType, expected []ContentType) error {
	return ValidateContentType(actual, expected)
}
