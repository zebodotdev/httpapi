package httpapi

import (
	"context"

	authpkg "github.com/zebodotdev/httpapi/auth"
	requestpkg "github.com/zebodotdev/httpapi/request"
)

const (
	SessionAuthModeBearer  = requestpkg.SessionAuthModeBearer
	SessionAuthModeService = requestpkg.SessionAuthModeService
)

var (
	ErrNoKeyToken                 = requestpkg.ErrNoKeyToken
	ErrAuthenticatorNotConfigured = requestpkg.ErrAuthenticatorNotConfigured
	ErrAuthzFailed                = requestpkg.ErrAuthzFailed
)

type App = authpkg.App
type Session = authpkg.Session
type AuthenticationRequest = authpkg.AuthenticationRequest
type Authenticator = requestpkg.Authenticator
type AuthenticatorFunc = requestpkg.AuthenticatorFunc

func ConfigureAuthenticator(auth Authenticator) func() {
	return requestpkg.ConfigureAuthenticator(auth)
}

func ContextWithSession(ctx context.Context, session *Session) context.Context {
	return requestpkg.ContextWithSession(ctx, session)
}

func SessionFromContext(ctx context.Context) *Session {
	return requestpkg.SessionFromContext(ctx)
}

func ContextWithAuthenticatedApp(ctx context.Context, appID string) context.Context {
	return requestpkg.ContextWithAuthenticatedApp(ctx, appID)
}
