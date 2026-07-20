package auth

import (
	"errors"
	"time"
)

const (
	// TypeUnknown is used when an Authorization header is present but its
	// scheme does not match any configured httpapi authentication mode.
	TypeUnknown = "unknown"

	// TypeBearer identifies a request authenticated with a regular bearer
	// credential, usually representing a direct API client or user-facing app.
	TypeBearer = "bearer"

	// TypeService identifies a request authenticated with a service assertion
	// or other trusted service-to-service credential.
	TypeService = "service"

	// SessionModeBearer is the session AuthMode value for bearer sessions.
	SessionModeBearer = TypeBearer

	// SessionModeService is the session AuthMode value for service sessions.
	SessionModeService = TypeService
)

var (
	// ErrNoKeyToken reports that a recognized Authorization scheme did not
	// contain the credential material needed by the configured authenticator.
	ErrNoKeyToken = errors.New("httpapi: no_key_token")

	// ErrAuthenticatorNotConfigured reports that request authentication was
	// attempted before a concrete Authenticator was installed.
	ErrAuthenticatorNotConfigured = errors.New("httpapi: authenticator_not_configured")

	// ErrAuthzFailed reports that credentials were parsed but could not produce
	// an authorized session.
	ErrAuthzFailed = errors.New("httpapi: api_auth_failed")

	// ErrUnsupportedAuthorizationScheme reports that a request used an
	// Authorization scheme that httpapi is not configured to understand.
	ErrUnsupportedAuthorizationScheme = errors.New("httpapi: unsupported_authorization_scheme")
)

// App identifies the application a session belongs to.
type App struct {
	// ID is the stable application identifier used as the request partition key
	// in audit records and downstream authorization decisions.
	ID string `json:"id"`

	// Description is optional human-readable context for operators inspecting a
	// session.
	Description string `json:"description,omitempty"`

	// Alias is an optional short, stable name for the application.
	Alias string `json:"alias,omitempty"`

	// Name is the display name for the application.
	Name string `json:"name,omitempty"`
}

// Session is an authn/authz session attached by the configured authenticator.
type Session struct {
	// ID is the session identifier returned by the authenticator.
	ID string `json:"id"`

	// App is the application the request is authorized to act against.
	App App `json:"app"`

	// Token is optional bearer credential material that may be forwarded to a
	// downstream service. It is redacted when a Session is converted to audit
	// form.
	Token string `json:"token,omitempty"`

	// SecretKeyID identifies the API key, bearer key, or equivalent credential
	// record that produced this session.
	SecretKeyID string `json:"secret_key_id,omitempty"`

	// InitiatedAt is when the session was created by the authority.
	InitiatedAt time.Time `json:"initiated_at"`

	// ExpiresAt is the time after which the session must not authorize work.
	ExpiresAt time.Time `json:"expires_at"`

	// MultiUse indicates whether the credential may be used for more than one
	// request.
	MultiUse bool `json:"multi_use"`

	// AuthMode identifies the kind of credential that produced the session. Use
	// SessionModeBearer and SessionModeService for standard httpapi modes.
	AuthMode string `json:"auth_mode,omitempty"`

	// ActorService is the service name that authenticated to obtain or assert
	// this session.
	ActorService string `json:"actor_service,omitempty"`

	// ActorServiceIdentity is the canonical service identity from the service
	// registry, certificate subject, or equivalent trust source.
	ActorServiceIdentity string `json:"actor_service_identity,omitempty"`

	// ActorCertFingerprint is the fingerprint of the service certificate or
	// signing certificate used by the actor.
	ActorCertFingerprint string `json:"actor_cert_fingerprint,omitempty"`

	// IssuerCertificateFingerprint is the fingerprint of the CA or issuing
	// certificate that chained the actor identity into the trusted root.
	IssuerCertificateFingerprint string `json:"issuer_certificate_fingerprint,omitempty"`

	// ServiceAuthKeyID identifies the service signing key used by a service
	// assertion.
	ServiceAuthKeyID string `json:"service_auth_key_id,omitempty"`

	// ServiceAuthSignatureHash is an audit-safe hash of the service assertion
	// signature or equivalent proof.
	ServiceAuthSignatureHash string `json:"service_auth_signature_hash,omitempty"`

	// ServiceAuthDecisionID links the session to a decision or verification
	// record held by the authenticator.
	ServiceAuthDecisionID string `json:"service_auth_decision_id,omitempty"`

	// ServiceAuthRoute is the route or endpoint name the service assertion was
	// scoped to, when the authenticator provides that detail.
	ServiceAuthRoute string `json:"service_auth_route,omitempty"`

	// ServiceAuthAction is the action name or scope action that authorized the
	// service request.
	ServiceAuthAction string `json:"service_auth_action,omitempty"`

	// SubjectUserID is the user on whose behalf the service is acting, when a
	// delegated user context exists.
	SubjectUserID string `json:"subject_user_id,omitempty"`

	// SubjectMemberID is the organization membership for the delegated user.
	SubjectMemberID string `json:"subject_member_id,omitempty"`

	// OrganizationID is the organization boundary for the session.
	OrganizationID string `json:"organization_id,omitempty"`

	// Audience is the intended audience for the credential or session.
	Audience string `json:"audience,omitempty"`

	// Scopes are the normalized permissions granted to the session.
	Scopes []string `json:"scopes,omitempty"`

	// Intent is optional caller-supplied context describing why the credential
	// was issued or how the service intends to use it.
	Intent string `json:"intent,omitempty"`

	// LifetimeClass labels the durability of the session, such as short-lived
	// or long-lived, without requiring callers to interpret exact timestamps.
	LifetimeClass string `json:"lifetime_class,omitempty"`

	// RequestID is the request identifier associated with credential issuance or
	// verification.
	RequestID string `json:"request_id,omitempty"`

	// TraceID carries distributed tracing context associated with the
	// authentication event.
	TraceID string `json:"trace_id,omitempty"`

	// TokenID is the JWT ID, JTI, or equivalent unique token identifier.
	TokenID string `json:"jti,omitempty"`

	// IdempotencyKey is the request idempotency key associated with the session
	// or assertion. It is redacted in audit output.
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// AuthenticationRequest describes the credential envelope parsed from a request.
type AuthenticationRequest struct {
	// Type is the normalized authentication type selected by httpapi, such as
	// TypeBearer or TypeService.
	Type string

	// Scheme is the literal Authorization header scheme found on the request.
	Scheme string

	// Authorization is the full Authorization header value. Authenticators may
	// need the whole value for service assertions; avoid logging it directly.
	Authorization string

	// Token is the parsed credential token for bearer-style schemes.
	Token string

	// RequestID is the httpapi request identifier assigned before
	// authentication.
	RequestID string

	// Method is the incoming HTTP method.
	Method string

	// RequestTarget is the request-target used for signature verification. It
	// includes the path and query string when present.
	RequestTarget string

	// BodySHA256 is the lowercase hex SHA-256 digest of the request body bytes.
	BodySHA256 string

	// UserAgent is the request User-Agent value after httpapi normalization.
	UserAgent string

	// RemoteAddr is the remote network address observed by the Go HTTP server.
	RemoteAddr string

	// TraceID carries inbound tracing context when the request provided it.
	TraceID string
}

// Valid reports whether the session is currently within its allowed lifetime.
func (s *Session) Valid() bool { return !s.Expired() }

// Authorized reports whether the session can authorize work. It currently
// matches Valid so services can evolve authorization semantics without changing
// handler code.
func (s *Session) Authorized() bool { return s.Valid() }

// Expired reports whether the session is nil or expired. A small clock-skew
// allowance treats sessions that expired more than four minutes ago as expired.
func (s *Session) Expired() bool {
	return s == nil || s.ExpiresAt.Before(time.Now().Add(-4*time.Minute))
}

// ServiceSession reports whether this session came from service
// authentication.
func (s *Session) ServiceSession() bool {
	return s != nil && s.AuthMode == SessionModeService
}

// ServiceScoped reports whether the session should be treated as scoped to a
// service actor. It currently aliases ServiceSession for compatibility.
func (s *Session) ServiceScoped() bool {
	return s.ServiceSession()
}
