package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

const (
	SessionAuthModeBearer  = authTypeBearer
	SessionAuthModeService = authTypeService
)

var (
	ErrNoKeyToken                 = errors.New("httpapi: no_key_token")
	ErrAuthenticatorNotConfigured = errors.New("httpapi: authenticator_not_configured")
	ErrAuthzFailed                = errors.New("httpapi: api_auth_failed")
)

type sessionContextKey struct{}

// App identifies the application a session belongs to.
type App struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	Alias       string `json:"alias,omitempty"`
	Name        string `json:"name,omitempty"`
}

// Session is an authn/authz session attached by the configured Authenticator.
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

// AuthenticationRequest describes the credential envelope parsed from a Req.
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

// Authenticator turns a supported Authorization header into a Session.
type Authenticator interface {
	Authenticate(context.Context, *Req, AuthenticationRequest) (*Session, error)
}

// AuthenticatorFunc adapts a function to Authenticator.
type AuthenticatorFunc func(context.Context, *Req, AuthenticationRequest) (*Session, error)

func (f AuthenticatorFunc) Authenticate(
	ctx context.Context,
	req *Req,
	auth AuthenticationRequest,
) (*Session, error) {
	return f(ctx, req, auth)
}

type noAuthenticator struct{}

func (noAuthenticator) Authenticate(
	context.Context,
	*Req,
	AuthenticationRequest,
) (*Session, error) {
	return nil, ErrAuthenticatorNotConfigured
}

var authenticatorState = struct {
	sync.RWMutex
	authenticator Authenticator
}{
	authenticator: noAuthenticator{},
}

// ConfigureAuthenticator installs the package-level authenticator.
// It returns a restore function for tests and short-lived overrides.
func ConfigureAuthenticator(auth Authenticator) func() {
	if auth == nil {
		auth = noAuthenticator{}
	}

	authenticatorState.Lock()
	prev := authenticatorState.authenticator
	authenticatorState.authenticator = auth
	authenticatorState.Unlock()

	return func() {
		authenticatorState.Lock()
		authenticatorState.authenticator = prev
		authenticatorState.Unlock()
	}
}

func currentAuthenticator() Authenticator {
	authenticatorState.RLock()
	defer authenticatorState.RUnlock()
	return authenticatorState.authenticator
}

// ContextWithSession returns a child context carrying a pre-authenticated
// session. Trusted middleware and tests can use it when authentication has
// already happened before the request reaches an Endpoint.
func ContextWithSession(ctx context.Context, session *Session) context.Context {
	if session == nil {
		return ctx
	}

	return context.WithValue(ctx, sessionContextKey{}, session)
}

// SessionFromContext returns a session previously installed by ContextWithSession.
func SessionFromContext(ctx context.Context) *Session {
	session, _ := ctx.Value(sessionContextKey{}).(*Session)
	return session
}

// ContextWithAuthenticatedApp returns a context carrying a generic bearer
// session for an application. It is primarily useful in tests.
func ContextWithAuthenticatedApp(ctx context.Context, appID string) context.Context {
	now := time.Now().UTC()
	return ContextWithSession(ctx, &Session{
		ID:          "sess_context",
		App:         App{ID: appID},
		InitiatedAt: now,
		ExpiresAt:   now.Add(30 * time.Minute),
		MultiUse:    true,
		AuthMode:    SessionAuthModeBearer,
	})
}

func (s *Session) valid() bool      { return !s.expired() }
func (s *Session) authorized() bool { return s.valid() }
func (s *Session) expired() bool {
	return s == nil || s.ExpiresAt.Before(time.Now().Add(-4*time.Minute))
}

func (s *Session) serviceSession() bool {
	return s != nil && s.AuthMode == SessionAuthModeService
}

func (s *Session) serviceScoped() bool {
	return s.serviceSession()
}

func (r *Req) Authenticated() bool { return r != nil && r.Sess.authorized() }
func (r *Req) Authorized() bool    { return r.Authenticated() }

func (r *Req) AttachSession(sess *Session) {
	if r == nil || sess == nil {
		return
	}

	r.Sess = sess
	r.SessID = sess.ID
	if strings.TrimSpace(sess.App.ID) != "" {
		r.AppID = strings.TrimSpace(sess.App.ID)
	}
}

func (r *Req) authenticate(authType string) error {
	auth := strings.TrimSpace(r.Authorization())
	var token string
	var ok bool

	switch authType {
	case authTypeBearer:
		token, ok = authorizationToken(auth, currentAuthorizationSchemes().Bearer)
	case authTypeService:
		ok = serviceAuthorizationValue(auth, currentAuthorizationSchemes())
	default:
		return nil
	}
	if !ok {
		return ErrNoKeyToken
	}

	sess, err := currentAuthenticator().Authenticate(r.Context(), r, AuthenticationRequest{
		Type:          authType,
		Scheme:        authorizationScheme(auth),
		Authorization: auth,
		Token:         token,
		RequestID:     r.ID,
		Method:        r.Method(),
		RequestTarget: requestTarget(r),
		BodySHA256:    requestBodySHA256(r.Body),
		UserAgent:     r.UserAgent(),
		RemoteAddr:    r.RemoteAddr(),
		TraceID:       strings.TrimSpace(r.Req.Header.Get(traceParentHeaderKey)),
	})
	if err != nil {
		return err
	}
	r.AttachSession(sess)
	return nil
}

func authorizationToken(auth, prefix string) (string, bool) {
	parts := strings.SplitN(strings.TrimSpace(auth), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], prefix) {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	return token, token != ""
}

func authorizationValue(auth, prefix string) bool {
	parts := strings.SplitN(strings.TrimSpace(auth), " ", 2)
	return len(parts) == 2 &&
		strings.EqualFold(parts[0], prefix) &&
		strings.TrimSpace(parts[1]) != ""
}

func serviceAuthorizationValue(auth string, schemes AuthorizationSchemes) bool {
	for _, scheme := range schemes.ServiceAuthorizationSchemes() {
		if authorizationValue(auth, scheme) {
			return true
		}
	}
	return false
}

func authorizationScheme(auth string) string {
	scheme, _, _ := strings.Cut(strings.TrimSpace(auth), " ")
	return strings.TrimSpace(scheme)
}

func requestTarget(r *Req) string {
	if r == nil || r.Req == nil || r.Req.URL == nil {
		return ""
	}
	return r.Req.URL.RequestURI()
}

func requestBodySHA256(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
