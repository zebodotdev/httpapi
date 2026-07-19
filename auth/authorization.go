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
	Required bool              `json:"required" yaml:"required"`
	Kind     AuthorizationKind `json:"kind,omitempty" yaml:"kind,omitempty"`
}

// RequiredAuthorization returns an authorization requirement for an endpoint spec.
func RequiredAuthorization(kind AuthorizationKind) AuthorizationRequirement {
	return NormalizeAuthorizationRequirement(AuthorizationRequirement{
		Required: true,
		Kind:     kind,
	})
}

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
