package endpoint

import authpkg "github.com/zebodotdev/httpapi/auth"

// AuthorizationKind identifies the authorization scheme an endpoint requires.
type AuthorizationKind = authpkg.AuthorizationKind

const (
	// AuthorizationKindAny accepts any authenticated session.
	AuthorizationKindAny AuthorizationKind = authpkg.AuthorizationKindAny
	// AuthorizationKindBearer requires a bearer secret-key session.
	AuthorizationKindBearer AuthorizationKind = authpkg.AuthorizationKindBearer
	// AuthorizationKindService requires service authorization.
	AuthorizationKindService AuthorizationKind = authpkg.AuthorizationKindService
)

// AuthorizationRequirement describes whether an endpoint requires auth and
// which authorization kind satisfies that requirement.
type AuthorizationRequirement = authpkg.AuthorizationRequirement

type endpointAccessPolicy struct {
	internal      bool
	auth          AuthorizationRequirement
	authInherited bool
}

// EndpointOption applies endpoint metadata at construction time.
type EndpointOption func(*Endpoint)

// WithInternal marks an endpoint as internal-only.
//
// Internal endpoints can only be called by service sessions. They may still
// also declare an explicit authorization requirement.
func WithInternal() EndpointOption {
	return func(e *Endpoint) {
		e.mutableAccessPolicy().internal = true
	}
}

// WithRequiredAuthorization declares the authorization kind required by an
// endpoint.
func WithRequiredAuthorization(kind AuthorizationKind) EndpointOption {
	auth := requiredAuthorization(kind)
	return func(e *Endpoint) {
		policy := e.mutableAccessPolicy()
		policy.auth = auth
		policy.authInherited = false
	}
}

// WithAuthorization is an alias for WithRequiredAuthorization.
func WithAuthorization(kind AuthorizationKind) EndpointOption {
	return WithRequiredAuthorization(kind)
}

// RequiredAuthorization returns an authorization requirement for an endpoint
// spec.
func RequiredAuthorization(kind AuthorizationKind) AuthorizationRequirement {
	return requiredAuthorization(kind)
}

// MarkInternal marks all endpoints in the group as internal-only and rebuilds
// their runtime wrappers.
func (eg *EndpointGroup) MarkInternal() {
	eg.Internal = true
	for i := range eg.Endpoints {
		eg.Endpoints[i].mutableAccessPolicy().internal = true
		eg.Endpoints[i] = eg.Endpoints[i].withRebuiltHandler()
	}
}

// RequireAuthorization declares the default authorization required by endpoints
// in the group. Endpoint-level requirements override the group default.
func (eg *EndpointGroup) RequireAuthorization(kind AuthorizationKind) {
	eg.Auth = requiredAuthorization(kind)
	for i := range eg.Endpoints {
		policy := eg.Endpoints[i].mutableAccessPolicy()
		if policy.auth.Required && !policy.authInherited {
			continue
		}
		policy.auth = eg.Auth
		policy.authInherited = true
		eg.Endpoints[i] = eg.Endpoints[i].withRebuiltHandler()
	}
}

func requiredAuthorization(kind AuthorizationKind) AuthorizationRequirement {
	return authpkg.RequiredAuthorization(kind)
}

func normalizeAuthorizationRequirement(auth AuthorizationRequirement) AuthorizationRequirement {
	return authpkg.NormalizeAuthorizationRequirement(auth)
}

func normalizeAuthorizationKind(kind AuthorizationKind) AuthorizationKind {
	return authpkg.NormalizeAuthorizationKind(kind)
}

func (eg EndpointGroup) endpointWithGroupMetadata(endpoint Endpoint) Endpoint {
	// force the endpoint to be internal if the group
	// is internal. this is because it doesn't make sense
	// to have an endpoint in an internal group that isn't
	// internal itself. if the group is internal, then all
	// of its endpoints should be internal as well.
	policy := endpoint.mutableAccessPolicy()
	if eg.Internal {
		policy.internal = true
	}

	auth := normalizeAuthorizationRequirement(eg.Auth)
	if auth.Required && (!policy.auth.Required || policy.authInherited) {
		policy.auth = auth
		policy.authInherited = true
	}

	priority := normalizeEndpointPriority(eg.Priority)
	priorityPolicy := endpoint.mutablePriorityPolicy()
	if priority != "" && (priorityPolicy.priority == "" || priorityPolicy.inherited) {
		priorityPolicy.priority = priority
		priorityPolicy.inherited = true
	}

	endpoint.mutableTimeoutPolicy().inheritDefaults(eg.Timeout)
	endpoint.mutableTimeoutPolicy().inheritHandler(eg.TimeoutHandler)
	endpoint.route = endpoint.routeSpec().WithDefaults(eg.Route)

	return endpoint.withRebuiltHandler()
}

func (e Endpoint) accessPolicy() endpointAccessPolicy {
	return e.access
}

func (e *Endpoint) mutableAccessPolicy() *endpointAccessPolicy {
	return &e.access
}
