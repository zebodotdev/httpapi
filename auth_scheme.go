package httpapi

import authpkg "github.com/zebodotdev/httpapi/auth"

const (
	defaultBearerAuthorizationScheme  = authpkg.DefaultBearerAuthorizationScheme
	defaultServiceAuthorizationScheme = authpkg.DefaultServiceAuthorizationScheme
)

type AuthorizationSchemes = authpkg.AuthorizationSchemes

// ConfigureAuthorizationSchemes sets the Authorization header schemes used by
// NewReq when deciding which authenticator mode to invoke.
func ConfigureAuthorizationSchemes(schemes AuthorizationSchemes) func() {
	return authpkg.ConfigureAuthorizationSchemes(schemes)
}

func currentAuthorizationSchemes() AuthorizationSchemes {
	return authpkg.CurrentAuthorizationSchemes()
}

func normalizeAuthorizationSchemes(schemes AuthorizationSchemes) AuthorizationSchemes {
	return authpkg.NormalizeAuthorizationSchemes(schemes)
}
