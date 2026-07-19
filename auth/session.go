package auth

import (
	"errors"
	"time"
)

const (
	TypeUnknown = "unknown"
	TypeBearer  = "bearer"
	TypeService = "service"

	SessionModeBearer  = TypeBearer
	SessionModeService = TypeService
)

var (
	ErrNoKeyToken                     = errors.New("httpapi: no_key_token")
	ErrAuthenticatorNotConfigured     = errors.New("httpapi: authenticator_not_configured")
	ErrAuthzFailed                    = errors.New("httpapi: api_auth_failed")
	ErrUnsupportedAuthorizationScheme = errors.New("httpapi: unsupported_authorization_scheme")
)

// App identifies the application a session belongs to.
type App struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	Alias       string `json:"alias,omitempty"`
	Name        string `json:"name,omitempty"`
}

// Session is an authn/authz session attached by the configured authenticator.
type Session struct {
	ID                           string    `json:"id"`
	App                          App       `json:"app"`
	Token                        string    `json:"token,omitempty"`
	SecretKeyID                  string    `json:"secret_key_id,omitempty"`
	InitiatedAt                  time.Time `json:"initiated_at"`
	ExpiresAt                    time.Time `json:"expires_at"`
	MultiUse                     bool      `json:"multi_use"`
	AuthMode                     string    `json:"auth_mode,omitempty"`
	ActorService                 string    `json:"actor_service,omitempty"`
	ActorServiceIdentity         string    `json:"actor_service_identity,omitempty"`
	ActorCertFingerprint         string    `json:"actor_cert_fingerprint,omitempty"`
	IssuerCertificateFingerprint string    `json:"issuer_certificate_fingerprint,omitempty"`
	ServiceAuthKeyID             string    `json:"service_auth_key_id,omitempty"`
	ServiceAuthSignatureHash     string    `json:"service_auth_signature_hash,omitempty"`
	ServiceAuthDecisionID        string    `json:"service_auth_decision_id,omitempty"`
	ServiceAuthRoute             string    `json:"service_auth_route,omitempty"`
	ServiceAuthAction            string    `json:"service_auth_action,omitempty"`
	SubjectUserID                string    `json:"subject_user_id,omitempty"`
	SubjectMemberID              string    `json:"subject_member_id,omitempty"`
	OrganizationID               string    `json:"organization_id,omitempty"`
	Audience                     string    `json:"audience,omitempty"`
	Scopes                       []string  `json:"scopes,omitempty"`
	Intent                       string    `json:"intent,omitempty"`
	LifetimeClass                string    `json:"lifetime_class,omitempty"`
	RequestID                    string    `json:"request_id,omitempty"`
	TraceID                      string    `json:"trace_id,omitempty"`
	TokenID                      string    `json:"jti,omitempty"`
	IdempotencyKey               string    `json:"idempotency_key,omitempty"`
}

// AuthenticationRequest describes the credential envelope parsed from a request.
type AuthenticationRequest struct {
	Type          string
	Scheme        string
	Authorization string
	Token         string
	RequestID     string
	Method        string
	RequestTarget string
	BodySHA256    string
	UserAgent     string
	RemoteAddr    string
	TraceID       string
}

func (s *Session) Valid() bool { return !s.Expired() }

func (s *Session) Authorized() bool { return s.Valid() }

func (s *Session) Expired() bool {
	return s == nil || s.ExpiresAt.Before(time.Now().Add(-4*time.Minute))
}

func (s *Session) ServiceSession() bool {
	return s != nil && s.AuthMode == SessionModeService
}

func (s *Session) ServiceScoped() bool {
	return s.ServiceSession()
}
