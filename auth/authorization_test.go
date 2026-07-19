package auth

import (
	"testing"
	"time"
)

func TestRequiredAuthorizationNormalizesKind(t *testing.T) {
	requirement := RequiredAuthorization(" SERVICE ")
	if !requirement.Required {
		t.Fatalf("required = false, want true")
	}
	if requirement.Kind != AuthorizationKindService {
		t.Fatalf("kind = %q, want %q", requirement.Kind, AuthorizationKindService)
	}
}

func TestSessionSatisfiesAuthorization(t *testing.T) {
	now := time.Now().UTC()
	bearer := &Session{
		AuthMode:  SessionModeBearer,
		ExpiresAt: now.Add(time.Hour),
	}
	service := &Session{
		AuthMode:  SessionModeService,
		ExpiresAt: now.Add(time.Hour),
	}

	if !SessionSatisfiesAuthorization(bearer, AuthorizationKindBearer) {
		t.Fatalf("bearer session did not satisfy bearer authorization")
	}
	if SessionSatisfiesAuthorization(service, AuthorizationKindBearer) {
		t.Fatalf("service session satisfied bearer authorization")
	}
	if !SessionSatisfiesAuthorization(service, AuthorizationKindService) {
		t.Fatalf("service session did not satisfy service authorization")
	}
	if !SessionSatisfiesAuthorization(bearer, AuthorizationKindAny) {
		t.Fatalf("bearer session did not satisfy any authorization")
	}
}
