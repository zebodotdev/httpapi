package httpapi

import (
	"strings"
	"sync"
)

const (
	defaultBearerAuthorizationScheme  = "Bearer"
	defaultServiceAuthorizationScheme = "Service"
)

type AuthorizationSchemes struct {
	Bearer  string
	Service string
}

var authorizationSchemesState = struct {
	sync.RWMutex
	schemes AuthorizationSchemes
}{
	schemes: AuthorizationSchemes{
		Bearer:  defaultBearerAuthorizationScheme,
		Service: defaultServiceAuthorizationScheme,
	},
}

// ConfigureAuthorizationSchemes sets the Authorization header schemes used by
// NewReq when deciding which authenticator mode to invoke.
func ConfigureAuthorizationSchemes(schemes AuthorizationSchemes) func() {
	schemes = normalizeAuthorizationSchemes(schemes)

	authorizationSchemesState.Lock()
	prev := authorizationSchemesState.schemes
	authorizationSchemesState.schemes = schemes
	authorizationSchemesState.Unlock()

	return func() {
		authorizationSchemesState.Lock()
		authorizationSchemesState.schemes = prev
		authorizationSchemesState.Unlock()
	}
}

func currentAuthorizationSchemes() AuthorizationSchemes {
	authorizationSchemesState.RLock()
	defer authorizationSchemesState.RUnlock()
	return authorizationSchemesState.schemes
}

func normalizeAuthorizationSchemes(schemes AuthorizationSchemes) AuthorizationSchemes {
	schemes.Bearer = strings.TrimSpace(schemes.Bearer)
	if schemes.Bearer == "" {
		schemes.Bearer = defaultBearerAuthorizationScheme
	}
	schemes.Service = strings.TrimSpace(schemes.Service)
	if schemes.Service == "" {
		schemes.Service = defaultServiceAuthorizationScheme
	}
	return schemes
}
