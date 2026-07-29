package response

import (
	"net/http"

	e "github.com/zebodotdev/httpapi/erreur"
)

// ErrorBody is the standard JSON response body for structured httpapi errors.
type ErrorBody struct {
	// Err is the structured error returned to the client.
	Err *e.Error `json:"error"`
}

// ErrRes is kept as a short compatibility alias for ErrorBody.
type ErrRes = ErrorBody

// ErrorShape describes the standard structured httpapi error response body.
var ErrorShape = Object[ErrorBody](
	Required("error", ErrorObjectShape, func(body ErrorBody) *e.Error {
		return body.Err
	}),
)

// ErrorObjectShape describes the structured error object under ErrorBody.Err.
var ErrorObjectShape = Object[*e.Error](
	Optional("message", String(), func(err *e.Error) (string, bool) {
		return errorString(err, func(err *e.Error) string { return err.Message })
	}),
	Optional("fix_code", String(), func(err *e.Error) (string, bool) {
		return errorString(err, func(err *e.Error) string { return err.FixCode })
	}),
	Optional("detail", String(), func(err *e.Error) (string, bool) {
		return errorString(err, func(err *e.Error) string { return err.Detail })
	}),
	Optional("cause", String(), func(err *e.Error) (string, bool) {
		return errorString(err, func(err *e.Error) string { return err.Cause })
	}),
	Required("type", String(), func(err *e.Error) string {
		return requiredErrorString(err, func(err *e.Error) string { return err.Type })
	}),
	Required("code", String(), func(err *e.Error) string {
		return requiredErrorString(err, func(err *e.Error) string { return err.Code })
	}),
	Required("url", String(), func(err *e.Error) string {
		return requiredErrorString(err, func(err *e.Error) string { return err.URL })
	}),
)

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

	Render(r, JSON(status, ErrorBody{Err: &bodyErr}))
}

// RenderParamErr converts an ErrInvalidParam into the standard structured
// error response format and sets it on the target.
func RenderParamErr(r Target, err *e.ErrInvalidParam) {
	if err == nil {
		RenderErr(r, nil)
		return
	}

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

func errorString(err *e.Error, getter func(*e.Error) string) (string, bool) {
	value := requiredErrorString(err, getter)
	return value, value != ""
}

func requiredErrorString(err *e.Error, getter func(*e.Error) string) string {
	if err == nil || getter == nil {
		return ""
	}
	return getter(err)
}
