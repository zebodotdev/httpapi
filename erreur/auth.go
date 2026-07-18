package erreur

import "net/http"

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
