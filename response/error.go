package response

import (
	"net/http"
	"time"

	e "github.com/zebodotdev/httpapi/erreur"
)

// ErrRes is the standard JSON response body for structured httpapi errors.
type ErrRes struct {
	// Err is the structured error returned to the client.
	Err *e.Error `json:"error"`
}

// RenderErr sets a structured error response on the target.
//
// The response status is derived from the error. When the error has no URL,
// RenderErr asks the erreur package URL builder for one.
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

// RenderParamErr converts an ErrInvalidParam into the standard structured
// error response format and sets it on the target.
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
