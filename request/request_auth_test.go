package request

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewReqAuthenticatesServiceAlias(t *testing.T) {
	restoreSchemes := ConfigureAuthorizationSchemes(AuthorizationSchemes{
		Service:        "Service",
		ServiceAliases: []string{"Commerce-Service", "System-Internal"},
	})
	t.Cleanup(restoreSchemes)

	restoreAuthenticator := ConfigureAuthenticator(AuthenticatorFunc(
		func(_ context.Context, _ *Req, auth AuthenticationRequest) (*Session, error) {
			if auth.Type != SessionAuthModeService {
				t.Fatalf("auth type = %q, want %q", auth.Type, SessionAuthModeService)
			}
			if auth.Scheme != "System-Internal" {
				t.Fatalf("auth scheme = %q, want System-Internal", auth.Scheme)
			}
			now := time.Now().UTC()
			return &Session{
				ID:          "sess_service_alias",
				App:         App{ID: "app_alias"},
				InitiatedAt: now,
				ExpiresAt:   now.Add(time.Hour),
				AuthMode:    SessionAuthModeService,
			}, nil
		},
	))
	t.Cleanup(restoreAuthenticator)

	req := httptest.NewRequest(http.MethodPost, "/alias", nil)
	req.Header.Set("authorization", "System-Internal token")

	wrapped := NewReq(req)
	if !wrapped.Authorized() {
		t.Fatal("request was not authorized")
	}
	if wrapped.AppID != "app_alias" {
		t.Fatalf("app id = %q, want app_alias", wrapped.AppID)
	}
}

func TestNewReqUsesContextSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/context", nil)
	req = req.WithContext(ContextWithAuthenticatedApp(req.Context(), "app_context"))

	wrapped := NewReq(req)
	if !wrapped.Authorized() {
		t.Fatal("request was not authorized")
	}
	if wrapped.AppID != "app_context" {
		t.Fatalf("app id = %q, want app_context", wrapped.AppID)
	}
}
