package erreur

import (
	"fmt"
	"net/http"
)

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

// RequestTooLarge returns a 413 error when the full request exceeds the
// endpoint's accepted size.
func RequestTooLarge(maxBytes int64) *Error {
	code := "request_too_large"
	detail := "the full request, including the request line, headers, and body, exceeds this endpoint's configured size limit."
	if maxBytes > 0 {
		detail = fmt.Sprintf(
			"the full request, including the request line, headers, and body, exceeds this endpoint's configured %d byte size limit.",
			maxBytes,
		)
	}

	return PayloadTooLarge(code, "request is too large", detail)
}
