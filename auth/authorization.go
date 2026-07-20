package auth

import (
	"fmt"
	"strings"
)

// AuthorizationKind identifies the authorization scheme an endpoint requires.
type AuthorizationKind string

const (
	// AuthorizationKindAny accepts any authenticated session.
	AuthorizationKindAny AuthorizationKind = "any"
	// AuthorizationKindBearer requires a bearer session.
	AuthorizationKindBearer AuthorizationKind = TypeBearer
	// AuthorizationKindService requires service authorization.
	AuthorizationKindService AuthorizationKind = TypeService
)

// AuthorizationRequirement describes whether an endpoint requires auth and
// which authorization kind satisfies that requirement.
type AuthorizationRequirement struct {
	// Required determines whether the endpoint must have an authenticated
	// session before its handler runs.
	Required bool `json:"required" yaml:"required"`

	// Kind is the session mode that satisfies this requirement. It is ignored
	// when Required is false.
	Kind AuthorizationKind `json:"kind,omitempty" yaml:"kind,omitempty"`
}

// RequiredAuthorization returns an authorization requirement for an endpoint spec.
func RequiredAuthorization(kind AuthorizationKind) AuthorizationRequirement {
	return NormalizeAuthorizationRequirement(AuthorizationRequirement{
		Required: true,
		Kind:     kind,
	})
}

// NormalizeAuthorizationRequirement canonicalizes an endpoint authorization
// requirement and panics when a required endpoint omits or uses an unsupported
// authorization kind.
func NormalizeAuthorizationRequirement(auth AuthorizationRequirement) AuthorizationRequirement {
	if !auth.Required {
		return AuthorizationRequirement{}
	}

	auth.Kind = NormalizeAuthorizationKind(auth.Kind)
	if auth.Kind == "" {
		panic("httpapi: authorization kind is required")
	}

	return auth
}

// NormalizeAuthorizationKind canonicalizes a configured authorization kind.
//
// It accepts the built-in kind names case-insensitively and panics on unknown
// non-empty values so invalid endpoint definitions fail during construction.
func NormalizeAuthorizationKind(kind AuthorizationKind) AuthorizationKind {
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

// SessionSatisfiesAuthorization reports whether a session meets a normalized
// endpoint authorization requirement.
func SessionSatisfiesAuthorization(s *Session, kind AuthorizationKind) bool {
	if s == nil || !s.Authorized() {
		return false
	}

	switch NormalizeAuthorizationKind(kind) {
	case AuthorizationKindAny:
		return true
	case AuthorizationKindBearer:
		return !s.ServiceScoped()
	case AuthorizationKindService:
		return s.ServiceSession()
	default:
		return false
	}
}
