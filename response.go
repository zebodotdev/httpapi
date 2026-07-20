package httpapi

import (
	"io"
	"net/http"
	"time"

	e "github.com/zebodotdev/httpapi/erreur"
	responsepkg "github.com/zebodotdev/httpapi/response"
)

type Res = responsepkg.Res
type ErrRes = responsepkg.ErrRes
type ResponseWriteResult = responsepkg.WriteResult

// ResponseWriteOptions preserves the root httpapi write options shape while the
// response package owns the concrete writer implementation.
type ResponseWriteOptions struct {
	RequestID string
	Duration  time.Duration
	Timeout   EndpointTimeoutSpec
}

func RenderJSON(r *Req, status int, body any) {
	responsepkg.RenderJSON(r, status, body)
}

func RenderStream(r *Req, status int, contentType string, header http.Header, body io.Reader) {
	responsepkg.RenderStream(r, status, contentType, header, body)
}

func RenderErr(r *Req, err *e.Error) {
	responsepkg.RenderErr(r, err)
}

func RenderParamErr(r *Req, err *e.ErrInvalidParam) {
	responsepkg.RenderParamErr(r, err)
}

func WriteResponse(
	w http.ResponseWriter,
	res *Res,
	opts ResponseWriteOptions,
) (ResponseWriteResult, error) {
	return responsepkg.WriteResponse(w, res, responseWriteOptions(opts))
}

func EncodeResponseBody(res *Res) ([]byte, error) {
	return responsepkg.EncodeResponseBody(res)
}

func WriteResponseBody(
	w http.ResponseWriter,
	res *Res,
	body []byte,
	opts ResponseWriteOptions,
) (ResponseWriteResult, error) {
	return responsepkg.WriteResponseBody(w, res, body, responseWriteOptions(opts))
}

func WriteResponseStream(
	w http.ResponseWriter,
	res *Res,
	opts ResponseWriteOptions,
) (ResponseWriteResult, error) {
	return responsepkg.WriteResponseStream(w, res, responseWriteOptions(opts))
}

func responseWriteOptions(opts ResponseWriteOptions) responsepkg.WriteOptions {
	return responsepkg.WriteOptions{
		RequestID:    opts.RequestID,
		Duration:     opts.Duration,
		WriteTimeout: normalizeEndpointTimeoutSpec(opts.Timeout).Write,
	}
}
