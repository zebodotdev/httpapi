package httpapi

import endpointpkg "github.com/zebodotdev/httpapi/endpoint"

type HttpMethod = endpointpkg.HttpMethod
type ContentType = endpointpkg.ContentType
type Handler = endpointpkg.Handler
type Endpoint = endpointpkg.Endpoint
type EndpointGroup = endpointpkg.EndpointGroup
type EndpointSpec = endpointpkg.EndpointSpec
type EndpointAccessSpec = endpointpkg.EndpointAccessSpec
type EndpointIdempotencySpec = endpointpkg.EndpointIdempotencySpec
type EndpointOption = endpointpkg.EndpointOption
type EndpointPriority = endpointpkg.EndpointPriority
type EndpointTimeoutSpec = endpointpkg.EndpointTimeoutSpec
type EndpointTimeoutHandler = endpointpkg.EndpointTimeoutHandler
type RoutePathMode = endpointpkg.RoutePathMode
type RouteBackend = endpointpkg.RouteBackend
type RouteSpec = endpointpkg.RouteSpec
type AuthorizationKind = endpointpkg.AuthorizationKind
type AuthorizationRequirement = endpointpkg.AuthorizationRequirement
type RequestMeta = endpointpkg.RequestMeta
type IdempotencyScopeResolver = endpointpkg.IdempotencyScopeResolver
type IdempotencyRecord = endpointpkg.IdempotencyRecord
type IdempotencyStore = endpointpkg.IdempotencyStore
type AuditSink = endpointpkg.AuditSink
type AuditSinkFunc = endpointpkg.AuditSinkFunc

const (
	APIRequestsTable = endpointpkg.APIRequestsTable

	OPTIONS HttpMethod = endpointpkg.OPTIONS
	POST    HttpMethod = endpointpkg.POST
	GET     HttpMethod = endpointpkg.GET

	ApplicationJson           ContentType = endpointpkg.ApplicationJson
	ApplicationFormURLEncoded ContentType = endpointpkg.ApplicationFormURLEncoded
	MultipartFormData         ContentType = endpointpkg.MultipartFormData
	TextHTML                  ContentType = endpointpkg.TextHTML
	TextPlain                 ContentType = endpointpkg.TextPlain

	AuthorizationKindAny     AuthorizationKind = endpointpkg.AuthorizationKindAny
	AuthorizationKindBearer  AuthorizationKind = endpointpkg.AuthorizationKindBearer
	AuthorizationKindService AuthorizationKind = endpointpkg.AuthorizationKindService

	EndpointPriorityCritical EndpointPriority = endpointpkg.EndpointPriorityCritical
	EndpointPriorityHigh     EndpointPriority = endpointpkg.EndpointPriorityHigh
	EndpointPriorityStandard EndpointPriority = endpointpkg.EndpointPriorityStandard
	EndpointPriorityLow      EndpointPriority = endpointpkg.EndpointPriorityLow

	RoutePathModeAppend   RoutePathMode = endpointpkg.RoutePathModeAppend
	RoutePathModeConstant RoutePathMode = endpointpkg.RoutePathModeConstant
)

var ErrIdempotencyStoreNotConfigured = endpointpkg.ErrIdempotencyStoreNotConfigured

func DefineEndpoint(spec EndpointSpec) Endpoint {
	return endpointpkg.DefineEndpoint(spec)
}

func NewEndpoint(
	meth HttpMethod,
	pattern string,
	handler Handler,
	opts ...EndpointOption,
) Endpoint {
	return endpointpkg.NewEndpoint(meth, pattern, handler, opts...)
}

func NewIdempotentEndpoint(
	meth HttpMethod,
	pattern string,
	handler Handler,
	opts ...EndpointOption,
) Endpoint {
	return endpointpkg.NewIdempotentEndpoint(meth, pattern, handler, opts...)
}

func NewIdempotentEndpointWithScopeResolver(
	meth HttpMethod,
	pattern string,
	resolver IdempotencyScopeResolver,
	handler Handler,
	opts ...EndpointOption,
) Endpoint {
	return endpointpkg.NewIdempotentEndpointWithScopeResolver(meth, pattern, resolver, handler, opts...)
}

func WithAcceptedContentTypes(contentTypes ...ContentType) EndpointOption {
	return endpointpkg.WithAcceptedContentTypes(contentTypes...)
}

func WithInternal() EndpointOption {
	return endpointpkg.WithInternal()
}

func WithRequiredAuthorization(kind AuthorizationKind) EndpointOption {
	return endpointpkg.WithRequiredAuthorization(kind)
}

func WithAuthorization(kind AuthorizationKind) EndpointOption {
	return endpointpkg.WithAuthorization(kind)
}

func RequiredAuthorization(kind AuthorizationKind) AuthorizationRequirement {
	return endpointpkg.RequiredAuthorization(kind)
}

func WithRouteSpec(spec RouteSpec) EndpointOption {
	return endpointpkg.WithRouteSpec(spec)
}

func WithRouteBackend(backend RouteBackend) EndpointOption {
	return endpointpkg.WithRouteBackend(backend)
}

func WithPriority(priority EndpointPriority) EndpointOption {
	return endpointpkg.WithPriority(priority)
}

func WithTimeoutSpec(spec EndpointTimeoutSpec) EndpointOption {
	return endpointpkg.WithTimeoutSpec(spec)
}

func WithTimeoutHandler(handler EndpointTimeoutHandler) EndpointOption {
	return endpointpkg.WithTimeoutHandler(handler)
}

func DefaultEndpointTimeoutHandler(req *Req) {
	endpointpkg.DefaultEndpointTimeoutHandler(req)
}

func ConfigureAuditSink(sink AuditSink) func() {
	return endpointpkg.ConfigureAuditSink(sink)
}

func ConfigureIdempotencyStore(store IdempotencyStore) func() {
	return endpointpkg.ConfigureIdempotencyStore(store)
}

func ConfigureIdempotencyScopeNamespace(namespace string) func() {
	return endpointpkg.ConfigureIdempotencyScopeNamespace(namespace)
}

func normalizeEndpointTimeoutSpec(spec EndpointTimeoutSpec) EndpointTimeoutSpec {
	return endpointpkg.NormalizeTimeoutSpec(spec)
}
