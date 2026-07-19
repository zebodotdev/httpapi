package response

import (
	"net/http"
	"time"

	e "github.com/zebodotdev/httpapi/erreur"
)

type ErrRes struct {
	Err *e.Error `json:"error"`
}

func RenderErr(r Target, err *e.Error) {
	if r == nil {
		return
	}
	if err == nil {
		err = e.Unexpected()
	}

	status := e.ResponseStatus(err)
	bodyErr := *err
	bodyErr.Status = status
	if bodyErr.URL == "" && bodyErr.Code != "" {
		bodyErr.URL = e.URLFor(
			bodyErr.Code,
			bodyErr.Type,
			bodyErr.Cause,
			bodyErr.FixCode,
		)
	}

	r.SetResponse(&Res{
		ContentType: ApplicationJson,
		Status:      status,
		SentAt:      time.Now(),
		Body: ErrRes{
			Err: &bodyErr,
		},
	})
}

func RenderParamErr(r Target, err *e.ErrInvalidParam) {
	status := err.Status
	if status == 0 {
		status = http.StatusBadRequest
	}
	code := err.Code
	if code == "" {
		code = "invalid_request_parameter"
	}
	cause := err.Cause
	if cause == "" {
		cause = e.CauseInvalidParam
	}
	typ := err.Type
	if typ == "" {
		typ = e.TypeInvalidParam
	}
	fixCode := err.FixCode
	if fixCode == "" {
		fixCode = e.FixCodeChangeParams
	}

	RenderErr(r, &e.Error{
		Message: err.Mesg,
		FixCode: fixCode,
		Status:  status,
		Cause:   cause,
		Type:    typ,
		Code:    code,
		URL:     e.URLFor(code, typ, cause, fixCode),
	})
}
