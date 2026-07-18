package erreur

import "net/http"

func ResponseStatus(err *Error) int {
	if err == nil || err.Status == 0 {
		return http.StatusInternalServerError
	}

	return err.Status
}

func InvalidParam(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeChangeParams,
		Detail:  detail,
		Status:  http.StatusBadRequest,
		Cause:   CauseInvalidParam,
		Type:    TypeInvalidParam,
		Code:    code,
		URL:     URL(code),
	}
}

func NotFound(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeChangeParams,
		Detail:  detail,
		Status:  http.StatusNotFound,
		Cause:   CauseInvalidParam,
		Type:    TypeInvalidParam,
		Code:    code,
		URL:     URL(code),
	}
}

func Precondition(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeSatisfyPrecond,
		Detail:  detail,
		Status:  http.StatusBadRequest,
		Cause:   CausePreconditionUnmet,
		Type:    TypeBadRequest,
		Code:    code,
		URL:     URL(code),
	}
}

func Conflict(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeRefreshState,
		Detail:  detail,
		Status:  http.StatusConflict,
		Cause:   CauseStateConflict,
		Type:    TypeStateConflict,
		Code:    code,
		URL:     URL(code),
	}
}

func Transient(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeRepeatSame,
		Detail:  detail,
		Status:  http.StatusServiceUnavailable,
		Cause:   CauseServiceUnavailable,
		Type:    TypeTransient,
		Code:    code,
		URL:     URL(code),
	}
}

func Unexpected() *Error {
	return &Error{
		Message: "request failed",
		FixCode: FixCodeContactSupport,
		Detail:  "we could not process this request. contact support with the request id if the problem persists.",
		Status:  http.StatusInternalServerError,
		Cause:   CauseUnknown,
		Type:    TypeUnknown,
		Code:    "request_failed",
		URL:     URL("request_failed"),
	}
}
