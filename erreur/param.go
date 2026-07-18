package erreur

type ErrInvalidParam struct {
	Param string
	Mesg  string
	Code  string

	Status  int
	Cause   string
	Type    string
	FixCode string
}

func InvalidParamErr(param, mesg string) *ErrInvalidParam {
	return &ErrInvalidParam{
		Param: param,
		Mesg:  mesg,
	}
}

func InvalidParamErrWithCode(param, mesg, code string) *ErrInvalidParam {
	err := InvalidParamErr(param, mesg)
	err.Code = code
	return err
}

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
