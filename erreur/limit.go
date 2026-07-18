package erreur

import "net/http"

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
