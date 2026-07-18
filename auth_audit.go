package httpapi

import (
	"errors"
	"strings"
	"time"
)

const (
	authTypeUnknown        = "unknown"
	authAuditTypeNone      = "none"
	authAuditTypeBearerKey = "bearer_token"
	authAuditTypeEndpoint  = "endpoint"
)

var ErrUnsupportedAuthorizationScheme = errors.New("httpapi: unsupported_authorization_scheme")

type AuthAudit struct {
	Type                  string       `json:"type"`
	Authenticated         bool         `json:"authenticated"`
	SessionAuthorized     bool         `json:"session_authorized"`
	Authorized            bool         `json:"authorized"`
	Failure               *AuthFailure `json:"failure,omitempty"`
	AuthenticationFailure *AuthFailure `json:"authentication_failure,omitempty"`
	AuthorizationFailure  *AuthFailure `json:"authorization_failure,omitempty"`
}

type AuthFailure struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

type SessionAudit struct {
	ID                           string    `json:"id,omitempty"`
	App                          App       `json:"app,omitzero"`
	Token                        string    `json:"token,omitempty"`
	TokenPresent                 bool      `json:"token_present,omitempty"`
	SecretKeyID                  string    `json:"secret_key_id,omitempty"`
	InitiatedAt                  time.Time `json:"initiated_at,omitzero"`
	ExpiresAt                    time.Time `json:"expires_at,omitzero"`
	MultiUse                     bool      `json:"multi_use,omitempty"`
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
	IdempotencyKeyPresent        bool      `json:"idempotency_key_present,omitempty"`
}

func (r *Req) recordAuthFailure(authType string, err error) {
	if r == nil || err == nil {
		return
	}

	failure := authFailureFor(authType, err)
	r.AuthFailure = &failure
}

func (r *Req) recordAuthorizationFailure(failure AuthFailure) {
	if r == nil {
		return
	}

	r.AuthorizationFailure = &failure
}

func (r Req) authAudit() AuthAudit {
	authenticated := r.Authenticated()
	sessionAuthorized := r.Authorized()
	authorizationFailure := r.AuthorizationFailure
	return AuthAudit{
		Type:                  authAuditType(r),
		Authenticated:         authenticated,
		SessionAuthorized:     sessionAuthorized,
		Authorized:            sessionAuthorized && authorizationFailure == nil,
		Failure:               firstAuthFailure(r.AuthFailure, authorizationFailure),
		AuthenticationFailure: r.AuthFailure,
		AuthorizationFailure:  authorizationFailure,
	}
}

func authAuditType(r Req) string {
	if r.Sess != nil && strings.TrimSpace(r.Sess.AuthMode) == SessionAuthModeService {
		return SessionAuthModeService
	}
	if r.Req == nil {
		return authAuditTypeNone
	}

	auth := strings.TrimSpace(r.Authorization())
	if auth == "" {
		return authAuditTypeNone
	}

	prefix, _, _ := strings.Cut(auth, " ")
	schemes := currentAuthorizationSchemes()
	switch strings.ToLower(strings.TrimSpace(prefix)) {
	case strings.ToLower(schemes.Bearer):
		return authAuditTypeBearerKey
	case strings.ToLower(schemes.Service):
		return SessionAuthModeService
	default:
		return authTypeUnknown
	}
}

func authFailureFor(authType string, err error) AuthFailure {
	failure := AuthFailure{
		Type:    authAuditTypeForAuthType(authType),
		Code:    "authentication_failed",
		Message: "authentication failed",
	}

	switch authType {
	case authTypeBearer:
		if errors.Is(err, ErrNoKeyToken) {
			failure.Code = "bearer_token_invalid"
			failure.Message = "Bearer authorization header is invalid"
			failure.Detail = "the authorization header did not contain a valid bearer token"
			return failure
		}
		if errors.Is(err, ErrAuthenticatorNotConfigured) {
			failure.Code = "authenticator_not_configured"
			failure.Message = "request authenticator is not configured"
			failure.Detail = "the service did not configure an authenticator for bearer authorization"
			return failure
		}

		failure.Code = "bearer_session_failed"
		failure.Message = "Bearer token could not be exchanged for a session"
		failure.Detail = "the authenticator did not return an authorized bearer session"
		return failure
	case authTypeService:
		if errors.Is(err, ErrNoKeyToken) {
			failure.Code = "service_authorization_invalid"
			failure.Message = "Service authorization header is invalid"
			failure.Detail = "the authorization header did not contain a valid service assertion"
			return failure
		}
		if errors.Is(err, ErrAuthenticatorNotConfigured) {
			failure.Code = "authenticator_not_configured"
			failure.Message = "request authenticator is not configured"
			failure.Detail = "the service did not configure an authenticator for service authorization"
			return failure
		}

		failure.Code = "service_session_failed"
		failure.Message = "Service assertion could not be exchanged for a session"
		failure.Detail = "the authenticator did not return an authorized service session"
		return failure
	default:
		if errors.Is(err, ErrUnsupportedAuthorizationScheme) {
			failure.Code = "authorization_scheme_unsupported"
			failure.Message = "authorization scheme is not supported"
			failure.Detail = "the authorization header uses a scheme this service does not accept"
			return failure
		}

		return failure
	}
}

func firstAuthFailure(failures ...*AuthFailure) *AuthFailure {
	for _, failure := range failures {
		if failure != nil {
			return failure
		}
	}
	return nil
}

func authAuditTypeForAuthType(authType string) string {
	switch authType {
	case authTypeBearer:
		return authAuditTypeBearerKey
	case authTypeService:
		return SessionAuthModeService
	default:
		return authTypeUnknown
	}
}

func sessionAudit(s *Session) *SessionAudit {
	if s == nil {
		return nil
	}

	audit := &SessionAudit{
		ID:                           s.ID,
		App:                          s.App,
		SecretKeyID:                  s.SecretKeyID,
		InitiatedAt:                  s.InitiatedAt,
		ExpiresAt:                    s.ExpiresAt,
		MultiUse:                     s.MultiUse,
		AuthMode:                     s.AuthMode,
		ActorService:                 s.ActorService,
		ActorServiceIdentity:         s.ActorServiceIdentity,
		ActorCertFingerprint:         s.ActorCertFingerprint,
		IssuerCertificateFingerprint: s.IssuerCertificateFingerprint,
		ServiceAuthKeyID:             s.ServiceAuthKeyID,
		ServiceAuthSignatureHash:     s.ServiceAuthSignatureHash,
		ServiceAuthDecisionID:        s.ServiceAuthDecisionID,
		ServiceAuthRoute:             s.ServiceAuthRoute,
		ServiceAuthAction:            s.ServiceAuthAction,
		SubjectUserID:                s.SubjectUserID,
		SubjectMemberID:              s.SubjectMemberID,
		OrganizationID:               s.OrganizationID,
		Audience:                     s.Audience,
		Scopes:                       append([]string(nil), s.Scopes...),
		Intent:                       s.Intent,
		LifetimeClass:                s.LifetimeClass,
		RequestID:                    s.RequestID,
		TraceID:                      s.TraceID,
		TokenID:                      s.TokenID,
	}
	if strings.TrimSpace(s.Token) != "" {
		audit.TokenPresent = true
		audit.Token = "REDACTED"
	}
	if strings.TrimSpace(s.IdempotencyKey) != "" {
		audit.IdempotencyKeyPresent = true
		audit.IdempotencyKey = "REDACTED"
	}

	return audit
}
