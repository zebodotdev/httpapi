package auth

import (
	"strings"
	"sync"
)

const (
	// DefaultBearerAuthorizationScheme is the Authorization header scheme used
	// for regular bearer credentials.
	DefaultBearerAuthorizationScheme = "Bearer"

	// DefaultServiceAuthorizationScheme is the Authorization header scheme used
	// for service-to-service credentials.
	DefaultServiceAuthorizationScheme = "Service"
)

// AuthorizationSchemes configures how httpapi interprets Authorization header
// scheme prefixes.
type AuthorizationSchemes struct {
	// Bearer is the primary scheme for bearer-token authentication.
	Bearer string

	// Service is the primary scheme for service-to-service authentication.
	Service string

	// ServiceAliases are additional schemes accepted as service authentication.
	// They are useful during migrations away from older service scheme names.
	ServiceAliases []string
}

var authorizationSchemesState = struct {
	sync.RWMutex
	schemes AuthorizationSchemes
}{
	schemes: AuthorizationSchemes{
		Bearer:  DefaultBearerAuthorizationScheme,
		Service: DefaultServiceAuthorizationScheme,
	},
}

// ConfigureAuthorizationSchemes sets the Authorization header schemes used when
// deciding which authenticator mode to invoke.
func ConfigureAuthorizationSchemes(schemes AuthorizationSchemes) func() {
	schemes = cloneAuthorizationSchemes(NormalizeAuthorizationSchemes(schemes))

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

// CurrentAuthorizationSchemes returns a copy of the globally configured
// Authorization header schemes.
func CurrentAuthorizationSchemes() AuthorizationSchemes {
	authorizationSchemesState.RLock()
	defer authorizationSchemesState.RUnlock()
	return cloneAuthorizationSchemes(authorizationSchemesState.schemes)
}

// NormalizeAuthorizationSchemes trims configured schemes, applies defaults,
// removes duplicate aliases, and returns a copy safe for storage.
func NormalizeAuthorizationSchemes(schemes AuthorizationSchemes) AuthorizationSchemes {
	schemes.Bearer = strings.TrimSpace(schemes.Bearer)
	if schemes.Bearer == "" {
		schemes.Bearer = DefaultBearerAuthorizationScheme
	}
	schemes.Service = strings.TrimSpace(schemes.Service)
	if schemes.Service == "" {
		schemes.Service = DefaultServiceAuthorizationScheme
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

func cloneAuthorizationSchemes(schemes AuthorizationSchemes) AuthorizationSchemes {
	schemes.ServiceAliases = append([]string(nil), schemes.ServiceAliases...)
	return schemes
}

// ServiceAuthorizationSchemes returns the primary service scheme followed by
// non-duplicate aliases in their configured order.
func (s AuthorizationSchemes) ServiceAuthorizationSchemes() []string {
	s = NormalizeAuthorizationSchemes(s)
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
