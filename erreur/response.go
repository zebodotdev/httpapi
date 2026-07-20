package erreur

import "net/http"

// ResponseStatus returns the HTTP status associated with err.
//
// Nil errors and errors without an explicit Status map to 500.
func ResponseStatus(err *Error) int {
	if err == nil || err.Status == 0 {
		return http.StatusInternalServerError
	}

	return err.Status
}

// New constructs an Error without a Detail field.
func New(status int, code, message, cause, typ, fixCode string) *Error {
	return WithDetail(status, code, message, "", cause, typ, fixCode)
}

// WithDetail constructs an Error and populates its documentation URL through
// the configured URL builder.
func WithDetail(status int, code, message, detail, cause, typ, fixCode string) *Error {
	return &Error{
		Message: message,
		FixCode: fixCode,
		Detail:  detail,
		Status:  status,
		Cause:   cause,
		Type:    typ,
		Code:    code,
		URL:     URLFor(code, typ, cause, fixCode),
	}
}

// InvalidParam returns a 400 error for invalid request parameters.
func InvalidParam(code, message, detail string) *Error {
	return WithDetail(
		http.StatusBadRequest,
		code,
		message,
		detail,
		CauseInvalidParam,
		TypeInvalidParam,
		FixCodeChangeParams,
	)
}

// Unauthorized returns a legacy 401 authorization failure.
//
// New authn/authz code should generally prefer Unauthenticated or Forbidden so
// the error Type distinguishes missing credentials from denied credentials.
func Unauthorized(code, message, detail string) *Error {
	return WithDetail(
		http.StatusUnauthorized,
		code,
		message,
		detail,
		CauseAuthzFailed,
		TypeAuthzFailed,
		FixCodeUseValidAPIKey,
	)
}

// NotFound returns a 404 error for missing resources.
func NotFound(code, message, detail string) *Error {
	return WithDetail(
		http.StatusNotFound,
		code,
		message,
		detail,
		CauseNotFound,
		TypeNotFound,
		FixCodeRefreshState,
	)
}

// Precondition returns a 400 error for unmet preconditions.
func Precondition(code, message, detail string) *Error {
	return WithDetail(
		http.StatusBadRequest,
		code,
		message,
		detail,
		CausePreconditionUnmet,
		TypeBadRequest,
		FixCodeSatisfyPrecond,
	)
}

// Conflict returns a 409 state-conflict error.
func Conflict(code, message, detail string) *Error {
	return WithDetail(
		http.StatusConflict,
		code,
		message,
		detail,
		CauseStateConflict,
		TypeStateConflict,
		FixCodeRefreshState,
	)
}

// StateInvalid returns an error for a resource currently in an invalid state
// for the attempted operation.
func StateInvalid(status int, code, message, detail string) *Error {
	return WithDetail(
		status,
		code,
		message,
		detail,
		CauseStateInvalid,
		TypeStateInvalid,
		FixCodeRefreshState,
	)
}

// StateConflict returns a 409 state-conflict error.
func StateConflict(code, message, detail string) *Error {
	return WithDetail(
		http.StatusConflict,
		code,
		message,
		detail,
		CauseStateConflict,
		TypeStateConflict,
		FixCodeRefreshState,
	)
}

// ServiceUnavailable returns a 503 service-unavailable error.
func ServiceUnavailable(code, message, detail string) *Error {
	return WithDetail(
		http.StatusServiceUnavailable,
		code,
		message,
		detail,
		CauseServiceUnavailable,
		TypeServiceUnavailable,
		FixCodeContactSupport,
	)
}

// Transient returns a 503 transient error suitable for retrying the same
// request later.
func Transient(code, message, detail string) *Error {
	return WithDetail(
		http.StatusServiceUnavailable,
		code,
		message,
		detail,
		CauseServiceUnavailable,
		TypeTransient,
		FixCodeRepeatSame,
	)
}

// Unexpected returns the standard opaque 500 response for unclassified
// failures.
func Unexpected() *Error {
	return WithDetail(
		http.StatusInternalServerError,
		"request_failed",
		"request failed",
		"we could not process this request. contact support with the request id if the problem persists.",
		CauseUnknown,
		TypeUnknown,
		FixCodeContactSupport,
	)
}
