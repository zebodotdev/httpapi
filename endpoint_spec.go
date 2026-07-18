package httpapi

import (
	"fmt"
	"maps"
	"strings"
)

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
	AuthKeys    map[string]bool
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
	normalized := HttpMethod(strings.ToUpper(strings.TrimSpace(string(method))))
	switch normalized {
	case POST, GET:
		return normalized
	default:
		panic(
			"invalid method for this http endpoint." +
				" supported methods are `GET` and `POST`",
		)
	}
}

func normalizeEndpointContentType(contentType ContentType) ContentType {
	contentType = ContentType(strings.TrimSpace(string(contentType)))
	if contentType == "" {
		return ApplicationJson
	}

	return contentType
}

func normalizeEndpointContentTypes(primary ContentType, additional ...ContentType) []ContentType {
	var contentTypes []ContentType
	seen := map[ContentType]bool{}
	if strings.TrimSpace(string(primary)) != "" || len(additional) == 0 {
		contentType := normalizeEndpointContentType(primary)
		contentTypes = append(contentTypes, contentType)
		seen[contentType] = true
	}
	for _, contentType := range additional {
		contentType = normalizeEndpointContentType(contentType)
		if seen[contentType] {
			continue
		}
		seen[contentType] = true
		contentTypes = append(contentTypes, contentType)
	}

	return contentTypes
}

func normalizeEndpointContentTypeSlice(contentTypes []ContentType) []ContentType {
	if len(contentTypes) == 0 {
		return []ContentType{ApplicationJson}
	}

	normalized := make([]ContentType, 0, len(contentTypes))
	seen := map[ContentType]bool{}
	for _, contentType := range contentTypes {
		contentType = normalizeEndpointContentType(contentType)
		if seen[contentType] {
			continue
		}
		seen[contentType] = true
		normalized = append(normalized, contentType)
	}
	return normalized
}

func primaryContentType(contentTypes []ContentType) ContentType {
	if len(contentTypes) == 0 {
		return ApplicationJson
	}
	return normalizeEndpointContentType(contentTypes[0])
}

func cloneContentTypes(contentTypes []ContentType) []ContentType {
	if len(contentTypes) == 0 {
		return nil
	}
	return append([]ContentType(nil), contentTypes...)
}

func joinContentTypes(contentTypes []ContentType) string {
	if len(contentTypes) == 0 {
		return string(ApplicationJson)
	}

	parts := make([]string, 0, len(contentTypes))
	for _, contentType := range contentTypes {
		parts = append(parts, string(normalizeEndpointContentType(contentType)))
	}
	return strings.Join(parts, ", ")
}

// WithAcceptedContentTypes sets every accepted request content type.
func WithAcceptedContentTypes(contentTypes ...ContentType) EndpointOption {
	contentTypes = normalizeEndpointContentTypes("", contentTypes...)
	return func(e *Endpoint) {
		e.accepts = cloneContentTypes(contentTypes)
	}
}

func cloneEndpointAuthKeys(keys map[string]bool) map[string]bool {
	if len(keys) == 0 {
		return nil
	}

	cloned := make(map[string]bool, len(keys))
	maps.Copy(cloned, keys)

	return cloned
}

func validateEndpointContentType(actual ContentType, expected []ContentType) error {
	actual = normalizeEndpointContentType(actual)
	expected = normalizeEndpointContentTypeSlice(expected)
	for _, candidate := range expected {
		if strings.Contains(string(actual), string(candidate)) {
			return nil
		}
	}

	return fmt.Errorf("httpapi: content type %q does not match any of %q", actual, joinContentTypes(expected))
}
