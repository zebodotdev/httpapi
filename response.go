package httpapi

import (
	"io"
	"net/http"

	e "github.com/zebodotdev/httpapi/erreur"
	responsepkg "github.com/zebodotdev/httpapi/response"
)

type Res = responsepkg.Res
type ErrRes = responsepkg.ErrRes
type ResponseWriteOptions = responsepkg.WriteOptions
type ResponseWriteResult = responsepkg.WriteResult

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
	return responsepkg.WriteResponse(w, res, opts)
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
	return responsepkg.WriteResponseBody(w, res, body, opts)
}

func WriteResponseStream(
	w http.ResponseWriter,
	res *Res,
	opts ResponseWriteOptions,
) (ResponseWriteResult, error) {
	return responsepkg.WriteResponseStream(w, res, opts)
}
