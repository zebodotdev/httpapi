package erreur

import "net/http"

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

func MissingParamErr(param, mesg string) *ErrInvalidParam {
	err := InvalidParamErrWithCode(param, mesg, "missing_request_parameter")
	err.Cause = CauseMissingParam
	return err
}
