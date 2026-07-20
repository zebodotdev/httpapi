package erreur

import "net/http"

// DependencyUnavailable returns a 503 error for a temporarily unavailable
// internal dependency.
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

// ProviderUnavailable returns a 503 error for a temporarily unavailable
// external provider or processor.
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

// ProviderDeclined returns a provider error for an operation actively rejected
// by an external provider.
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
