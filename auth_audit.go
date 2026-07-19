package httpapi

import (
	authpkg "github.com/zebodotdev/httpapi/auth"
	requestpkg "github.com/zebodotdev/httpapi/request"
)

const (
	authTypeUnknown        = authpkg.TypeUnknown
	authAuditTypeNone      = authpkg.AuditTypeNone
	authAuditTypeBearerKey = authpkg.AuditTypeBearerKey
	authAuditTypeEndpoint  = authpkg.AuditTypeEndpoint
)

var ErrUnsupportedAuthorizationScheme = requestpkg.ErrUnsupportedAuthorizationScheme

type AuthAudit = requestpkg.AuthAudit
type AuthFailure = requestpkg.AuthFailure
type SessionAudit = requestpkg.SessionAudit
