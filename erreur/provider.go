package erreur

import "net/http"

func DependencyUnavailable(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeBackoff,
		Detail:  detail,
		Status:  http.StatusServiceUnavailable,
		Cause:   CauseDependencyUnavailable,
		Type:    TypeTransient,
		Code:    code,
		URL:     URL(code),
	}
}

func ProviderUnavailable(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeBackoff,
		Detail:  detail,
		Status:  http.StatusServiceUnavailable,
		Cause:   CauseProviderUnavailable,
		Type:    TypeProvider,
		Code:    code,
		URL:     URL(code),
	}
}

func ProviderDeclined(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeUseDifferentProviderMethod,
		Detail:  detail,
		Status:  http.StatusBadRequest,
		Cause:   CauseProviderDeclined,
		Type:    TypeProvider,
		Code:    code,
		URL:     URL(code),
	}
}
