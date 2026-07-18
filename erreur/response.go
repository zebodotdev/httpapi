package erreur

import "net/http"

func ResponseStatus(err *Error) int {
	if err == nil || err.Status == 0 {
		return http.StatusInternalServerError
	}

	return err.Status
}

func New(status int, code, message, cause, typ, fixCode string) *Error {
	return WithDetail(status, code, message, "", cause, typ, fixCode)
}

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
