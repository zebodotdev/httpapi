package erreur

import "net/http"

// ResourceLocked returns a 409 error when the target resource is locked by
// another write or workflow.
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

// OperationInProgress returns a 409 error when the requested operation is
// already running and should not be started again.
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
