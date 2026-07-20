package erreur

import "net/http"

// ConfigurationMissing returns a 400 error for missing account, application, or
// service configuration required by the operation.
func ConfigurationMissing(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeResolveAccountConfiguration,
		Detail:  detail,
		Status:  http.StatusBadRequest,
		Cause:   CauseConfigurationMissing,
		Type:    TypeConfiguration,
		Code:    code,
		URL:     URL(code),
	}
}

// CapabilityUnsupported returns a 403 error when the requested capability is
// not enabled or not supported for the caller's account or application.
func CapabilityUnsupported(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeSatisfyPrecond,
		Detail:  detail,
		Status:  http.StatusForbidden,
		Cause:   CauseCapabilityUnsupported,
		Type:    TypeConfiguration,
		Code:    code,
		URL:     URL(code),
	}
}
