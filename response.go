package httpapi

import (
	"io"
	"net/http"
	"time"
)

// RenderJSON sets a JSON response on the request.
func RenderJSON(r *Req, status int, body any) {
	r.Res = &Res{
		ContentType: ApplicationJson,
		Status:      status,
		SentAt:      time.Now(),
		Header:      http.Header{},
		Body:        body,
	}
}

// RenderStream sets a streaming response without buffering the body in memory.
func RenderStream(r *Req, status int, contentType string, header http.Header, body io.Reader) {
	r.Res = &Res{
		ContentType: contentType,
		Status:      status,
		SentAt:      time.Now(),
		Header:      header,
		BodyReader:  body,
	}
}
