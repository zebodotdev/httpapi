package erreur

import "net/http"

func ResourceLocked(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeRefreshState,
		Detail:  detail,
		Status:  http.StatusConflict,
		Cause:   CauseResourceLocked,
		Type:    TypeResourceLock,
		Code:    code,
		URL:     URL(code),
	}
}

func OperationInProgress(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeBackoff,
		Detail:  detail,
		Status:  http.StatusConflict,
		Cause:   CauseOperationInProgress,
		Type:    TypeResourceLock,
		Code:    code,
		URL:     URL(code),
	}
}
