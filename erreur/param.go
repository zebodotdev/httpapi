package erreur

// ErrInvalidParam is a lightweight parameter validation error used by handlers
// and extension points that want response.RenderParamErr to build the final
// structured error response.
type ErrInvalidParam struct {
	// Param is the request parameter name that failed validation.
	Param string

	// Mesg is the human-readable validation message.
	Mesg string

	// Code is the stable machine-readable validation code.
	Code string

	// Status overrides the default HTTP status when non-zero.
	Status int

	// Cause overrides the default error cause when non-empty.
	Cause string

	// Type overrides the default error type when non-empty.
	Type string

	// FixCode overrides the default remediation code when non-empty.
	FixCode string
}

// InvalidParamErr returns a parameter validation error without a specific code.
func InvalidParamErr(param, mesg string) *ErrInvalidParam {
	return &ErrInvalidParam{
		Param: param,
		Mesg:  mesg,
	}
}

// InvalidParamErrWithCode returns a parameter validation error with a stable
// machine-readable code.
func InvalidParamErrWithCode(param, mesg, code string) *ErrInvalidParam {
	err := InvalidParamErr(param, mesg)
	err.Code = code
	return err
}

// InvalidBodyErr returns an ErrInvalidParam specialized for malformed request
// JSON bodies.
func InvalidBodyErr() *ErrInvalidParam {
	err := InvalidParamErrWithCode(
		"request_body",
		"could not decode request body."+
			" if you're sure that you're sending valid"+
			" json, please reach out to support",
		"invalid_request_body",
	)
	err.Cause = CauseInvalidBody
	err.Type = TypeInvalidRequest
	err.FixCode = FixCodeChangeParams
	return err
}
