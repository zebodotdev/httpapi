package auth

import (
	"strings"
	"time"
)

const (
	AuditTypeNone      = "none"
	AuditTypeBearerKey = "bearer_token"
	AuditTypeEndpoint  = "endpoint"
)

type Audit struct {
	Type                  string   `json:"type"`
	Authenticated         bool     `json:"authenticated"`
	SessionAuthorized     bool     `json:"session_authorized"`
	Authorized            bool     `json:"authorized"`
	Failure               *Failure `json:"failure,omitempty"`
	AuthenticationFailure *Failure `json:"authentication_failure,omitempty"`
	AuthorizationFailure  *Failure `json:"authorization_failure,omitempty"`
}

type Failure struct {
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

func SessionAuditFor(s *Session) *SessionAudit {
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
