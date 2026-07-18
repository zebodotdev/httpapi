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
	Bearer         string
	Service        string
	ServiceAliases []string
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
	schemes.ServiceAliases = normalizeAuthorizationSchemeAliases(schemes.ServiceAliases)
	return schemes
}

func normalizeAuthorizationSchemeAliases(aliases []string) []string {
	if len(aliases) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(aliases))
	seen := map[string]bool{}
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			continue
		}
		key := strings.ToLower(alias)
		if seen[key] {
			continue
		}
		seen[key] = true
		normalized = append(normalized, alias)
	}

	return normalized
}

func (s AuthorizationSchemes) ServiceAuthorizationSchemes() []string {
	s = normalizeAuthorizationSchemes(s)
	schemes := make([]string, 0, 1+len(s.ServiceAliases))
	schemes = append(schemes, s.Service)
	for _, alias := range s.ServiceAliases {
		if strings.EqualFold(alias, s.Service) {
			continue
		}
		schemes = append(schemes, alias)
	}
	return schemes
}
