package erreur

import "net/http"

// Unauthenticated returns a 401 error for missing, invalid, or expired
// credentials.
func Unauthenticated(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeAuthenticate,
		Detail:  detail,
		Status:  http.StatusUnauthorized,
		Cause:   CauseAuthnFailed,
		Type:    TypeAuthentication,
		Code:    code,
		URL:     URL(code),
	}
}

// Forbidden returns a 403 error for authenticated callers that are not allowed
// to perform the operation.
func Forbidden(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeUseAuthorizedCredentials,
		Detail:  detail,
		Status:  http.StatusForbidden,
		Cause:   CauseAuthzFailed,
		Type:    TypeAuthorization,
		Code:    code,
		URL:     URL(code),
	}
}
