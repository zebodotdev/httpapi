package erreur

import "net/http"

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
