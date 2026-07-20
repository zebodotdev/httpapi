package erreur

import "net/http"

// MissingParam returns a 400 error for a missing required request parameter.
func MissingParam(code, message, detail string) *Error {
	return &Error{
		Message: message,
		FixCode: FixCodeChangeParams,
		Detail:  detail,
		Status:  http.StatusBadRequest,
		Cause:   CauseMissingParam,
		Type:    TypeInvalidParam,
		Code:    code,
		URL:     URL(code),
	}
}

// MissingParamErr returns an ErrInvalidParam specialized for missing required
// parameters.
func MissingParamErr(param, mesg string) *ErrInvalidParam {
	err := InvalidParamErrWithCode(param, mesg, "missing_request_parameter")
	err.Cause = CauseMissingParam
	return err
}
