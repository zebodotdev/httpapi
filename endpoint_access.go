package httpapi

import (
	"fmt"

	errresp "github.com/zebodotdev/httpapi/erreur"
)

const (
	endpointAuthenticationRequiredCode         = "endpoint_authentication_required"
	endpointAuthorizationDeniedCode            = "endpoint_authorization_denied"
	internalEndpointAuthenticationRequiredCode = "internal_endpoint_authentication_required"
	internalEndpointAuthorizationDeniedCode    = "internal_endpoint_authorization_denied"
)

func (e Endpoint) accessError(r *Req) *errresp.Error {
	policy := e.accessPolicy()
	if policy.internal && !internalServiceRequest(r) {
		if r == nil || r.Sess == nil || !r.Authorized() {
			err := errresp.Unauthenticated(
				internalEndpointAuthenticationRequiredCode,
				"internal endpoint requires service authentication",
				"this endpoint is internal and can only be called by an authenticated service.",
			)
			recordEndpointAccessFailure(r, err)
			return err
		}

		err := errresp.Forbidden(
			internalEndpointAuthorizationDeniedCode,
			"internal endpoint requires service authentication",
			"this endpoint is internal and cannot be called with regular user credentials.",
		)
		recordEndpointAccessFailure(r, err)
		return err
	}

	if !policy.auth.Required {
		return nil
	}

	if r == nil || r.Sess == nil || !r.Authorized() {
		err := errresp.Unauthenticated(
			endpointAuthenticationRequiredCode,
			"endpoint requires authentication",
			fmt.Sprintf("this endpoint requires %s authorization.", policy.auth.Kind),
		)
		recordEndpointAccessFailure(r, err)
		return err
	}

	if sessionSatisfiesAuthorization(r.Sess, policy.auth.Kind) {
		return nil
	}

	err := errresp.Forbidden(
		endpointAuthorizationDeniedCode,
		"endpoint authorization denied",
		fmt.Sprintf("this endpoint requires %s authorization.", policy.auth.Kind),
	)
	recordEndpointAccessFailure(r, err)
	return err
}

func recordEndpointAccessFailure(r *Req, err *errresp.Error) {
	if r == nil || err == nil {
		return
	}

	r.recordAuthorizationFailure(AuthFailure{
		Type:    authAuditTypeEndpoint,
		Code:    err.Code,
		Message: err.Message,
		Detail:  err.Detail,
	})
}

func internalServiceRequest(r *Req) bool {
	return r != nil &&
		r.Sess != nil &&
		r.Sess.authorized() &&
		r.Sess.serviceSession()
}

func sessionSatisfiesAuthorization(s *Session, kind AuthorizationKind) bool {
	if s == nil || !s.authorized() {
		return false
	}

	switch normalizeAuthorizationKind(kind) {
	case AuthorizationKindAny:
		return true
	case AuthorizationKindBearer:
		return !s.serviceScoped()
	case AuthorizationKindService:
		return s.serviceSession()
	default:
		return false
	}
}
