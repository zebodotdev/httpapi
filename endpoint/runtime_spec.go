package endpoint

// EndpointSpec is the declarative contract for one HTTP endpoint.
//
// It should be the preferred way to define new endpoints. Endpoint requirements
// live here so Req can remain a safe parse of an incoming request, independent
// of the endpoint that eventually handles it.
type EndpointSpec struct {
	// Method is the HTTP method accepted by the endpoint. Empty and unsupported
	// values panic during endpoint construction.
	Method HttpMethod

	// Path is the route pattern registered on the Go ServeMux and later exposed
	// to transcribers.
	Path string

	// Handler is the application function invoked after httpapi has parsed the
	// request, authenticated it, applied access policy, and enforced idempotency.
	Handler Handler

	// Accepts is the primary request content type. It defaults to
	// ApplicationJson.
	Accepts ContentType

	// AcceptsAny declares additional accepted request content types. Duplicate
	// entries are removed after normalization.
	AcceptsAny []ContentType

	// Access declares whether the endpoint is internal and which session kind it
	// requires.
	Access EndpointAccessSpec

	// Idempotency declares whether successful responses should be reserved and
	// replayed for repeated idempotency keys.
	Idempotency EndpointIdempotencySpec

	// Route carries provider-neutral metadata for generated route documents.
	Route RouteSpec

	// Priority captures the operational importance of the endpoint.
	Priority EndpointPriority

	// Timeout declares in-process read, handler, and write deadlines for this
	// endpoint.
	Timeout EndpointTimeoutSpec

	// TimeoutHandler renders the response when the handler context reaches its
	// timeout before the endpoint produces a response. When unset, httpapi
	// renders DefaultEndpointTimeoutHandler.
	TimeoutHandler EndpointTimeoutHandler

	// AuthKeys is an optional arbitrary key set for service-specific
	// authorization metadata. httpapi clones the map before storing it.
	AuthKeys map[string]bool
}

// EndpointAccessSpec declares endpoint access requirements.
type EndpointAccessSpec struct {
	// Internal marks the endpoint as callable only by service sessions.
	Internal bool

	// Authorization declares whether a session is required and which session
	// kind satisfies the endpoint.
	Authorization AuthorizationRequirement
}

// EndpointIdempotencySpec declares endpoint idempotency behavior.
type EndpointIdempotencySpec struct {
	// Enabled requires callers to provide idempotency keys and stores successful
	// responses for replay.
	Enabled bool

	// ScopeResolver computes a custom idempotency scope for a request. Providing
	// a resolver implies Enabled even when Enabled is false.
	ScopeResolver IdempotencyScopeResolver
}

// DefineEndpoint returns a server endpoint from a declarative endpoint spec.
//
// Prefer DefineEndpoint for new code because all runtime, access,
// idempotency, priority, timeout, and route metadata lives in one named struct.
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
//
// It replaces the endpoint's existing content-type list rather than appending
// to it.
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
