package httpapi

import (
	"fmt"
	"strings"
)

// AuthorizationKind identifies the authorization scheme an endpoint requires.
type AuthorizationKind string

const (
	// AuthorizationKindAny accepts any authenticated session.
	AuthorizationKindAny AuthorizationKind = "any"
	// AuthorizationKindBearer requires a bearer secret-key session.
	AuthorizationKindBearer AuthorizationKind = authTypeBearer
	// AuthorizationKindService requires service authorization.
	AuthorizationKindService AuthorizationKind = authTypeService
)

// AuthorizationRequirement describes whether an endpoint requires auth and
// which authorization kind satisfies that requirement.
type AuthorizationRequirement struct {
	Required bool              `json:"required" yaml:"required"`
	Kind     AuthorizationKind `json:"kind,omitempty" yaml:"kind,omitempty"`
}

type endpointAccessPolicy struct {
	internal      bool
	auth          AuthorizationRequirement
	authInherited bool
}

// EndpointOption applies endpoint metadata at construction time.
type EndpointOption func(*Endpoint)

// WithInternal marks an endpoint as internal-only.
func WithInternal() EndpointOption {
	return func(e *Endpoint) {
		e.mutableAccessPolicy().internal = true
	}
}

// WithRequiredAuthorization declares the authorization kind required by an endpoint.
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

// RequiredAuthorization returns an authorization requirement for an endpoint spec.
func RequiredAuthorization(kind AuthorizationKind) AuthorizationRequirement {
	return requiredAuthorization(kind)
}

// MarkInternal marks all endpoints in the group as internal-only.
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
	return normalizeAuthorizationRequirement(AuthorizationRequirement{
		Required: true,
		Kind:     kind,
	})
}

func normalizeAuthorizationRequirement(auth AuthorizationRequirement) AuthorizationRequirement {
	if !auth.Required {
		return AuthorizationRequirement{}
	}

	auth.Kind = normalizeAuthorizationKind(auth.Kind)
	if auth.Kind == "" {
		panic("httpapi: authorization kind is required")
	}

	return auth
}

func normalizeAuthorizationKind(kind AuthorizationKind) AuthorizationKind {
	switch strings.TrimSpace(strings.ToLower(string(kind))) {
	case string(AuthorizationKindAny):
		return AuthorizationKindAny
	case string(AuthorizationKindBearer):
		return AuthorizationKindBearer
	case string(AuthorizationKindService):
		return AuthorizationKindService
	case "":
		return ""
	default:
		panic(fmt.Sprintf("httpapi: unsupported authorization kind %q", kind))
	}
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

	return endpoint.withRebuiltHandler()
}

func (e Endpoint) accessPolicy() endpointAccessPolicy {
	return e.access
}

func (e *Endpoint) mutableAccessPolicy() *endpointAccessPolicy {
	return &e.access
}
