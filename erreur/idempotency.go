package erreur

import "net/http"

// IdempotencyKeyGenerationFailed returns a transient idempotency error for
// services that could not generate a fallback key.
func IdempotencyKeyGenerationFailed() *Error {
	code := "idempotency_key_generation_failed"
	return &Error{
		Message: "could not generate an idempotency key for this request",
		FixCode: FixCodeRepeatSame,
		Detail:  "retry the same request later. the operation has not started.",
		Status:  http.StatusServiceUnavailable,
		Cause:   CauseServiceUnavailable,
		Type:    TypeIdempotency,
		Code:    code,
		URL:     URL(code),
	}
}

// IdempotencyInProgress returns an error for a key currently reserved by an
// incomplete operation.
func IdempotencyInProgress() *Error {
	code := "idempotency_key_in_progress"
	return &Error{
		Message: "idempotency key is already being processed",
		FixCode: FixCodeRepeatSame,
		Detail:  "a request with this idempotency key is already being processed. retry the same request later.",
		Status:  http.StatusConflict,
		Cause:   CauseIdempotencyInProgress,
		Type:    TypeIdempotency,
		Code:    code,
		URL:     URL(code),
	}
}

// IdempotencyConflict returns an error for an idempotency key reused with a
// different operation fingerprint.
func IdempotencyConflict() *Error {
	code := "idempotency_key_conflict"
	return &Error{
		Message: "idempotency key has already been used",
		FixCode: FixCodeUseNewIdempotencyKey,
		Detail:  "this idempotency key has already been used with a different operation payload. use a new idempotency key for this operation.",
		Status:  http.StatusUnprocessableEntity,
		Cause:   CauseIdempotencyConflict,
		Type:    TypeIdempotency,
		Code:    code,
		URL:     URL(code),
	}
}

// IdempotencyStorageUnavailable returns an error for idempotent endpoints whose
// backing store is temporarily unavailable or not configured.
func IdempotencyStorageUnavailable() *Error {
	code := "idempotency_storage_unavailable"
	return &Error{
		Message: "idempotent requests are temporarily unavailable",
		FixCode: FixCodeRepeatSame,
		Detail:  "retry the same request later. the operation has not started.",
		Status:  http.StatusServiceUnavailable,
		Cause:   CauseDependencyUnavailable,
		Type:    TypeIdempotency,
		Code:    code,
		URL:     URL(code),
	}
}
