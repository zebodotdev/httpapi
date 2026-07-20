package erreur

import "net/http"

// RateLimited returns a 429 error when the client is sending requests too
// quickly.
func RateLimited(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeBackoff,
		Detail:  detail,
		Status:  http.StatusTooManyRequests,
		Cause:   CauseRateLimited,
		Type:    TypeLimit,
		Code:    code,
		URL:     URL(code),
	}
}

// QuotaExceeded returns a 429 error when an account or application quota has
// been exhausted.
func QuotaExceeded(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeResolveAccountConfiguration,
		Detail:  detail,
		Status:  http.StatusTooManyRequests,
		Cause:   CauseQuotaExceeded,
		Type:    TypeLimit,
		Code:    code,
		URL:     URL(code),
	}
}

// LimitExceeded returns a 400 error when a request or resource exceeds a
// non-rate limit.
func LimitExceeded(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeReduceRequest,
		Detail:  detail,
		Status:  http.StatusBadRequest,
		Cause:   CauseLimitExceeded,
		Type:    TypeLimit,
		Code:    code,
		URL:     URL(code),
	}
}

// PayloadTooLarge returns a 413 error when the request payload exceeds the
// accepted size.
func PayloadTooLarge(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeReduceRequest,
		Detail:  detail,
		Status:  http.StatusRequestEntityTooLarge,
		Cause:   CausePayloadTooLarge,
		Type:    TypeLimit,
		Code:    code,
		URL:     URL(code),
	}
}
