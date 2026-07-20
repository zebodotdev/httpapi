package response

import (
	"io"
	"net/http"
	"time"
)

// Content type aliases kept here so response rendering does not depend on the
// root httpapi package.
const (
	ApplicationJson = "application/json"
	TextHTML        = "text/html"
	TextPlain       = "text/plain; charset=utf-8"
)

// Res is the response object populated by handlers and written by endpoint
// runtimes.
type Res struct {
	ContentType string      `json:"content_type"`
	Status      int         `json:"status"`
	SentAt      time.Time   `json:"sent_at"`
	Header      http.Header `json:"header"`
	Body        any         `json:"body"`
	BodyReader  io.Reader   `json:"-"`
}

// Target is the minimum request surface response renderers need.
type Target interface {
	SetResponse(*Res)
}

// RenderJSON sets a JSON response on the request.
func RenderJSON(r Target, status int, body any) {
	if r == nil {
		return
	}
	r.SetResponse(&Res{
		ContentType: ApplicationJson,
		Status:      status,
		SentAt:      time.Now(),
		Header:      http.Header{},
		Body:        body,
	})
}

// RenderStream sets a streaming response without buffering the body in memory.
func RenderStream(r Target, status int, contentType string, header http.Header, body io.Reader) {
	if r == nil {
		return
	}
	r.SetResponse(&Res{
		ContentType: contentType,
		Status:      status,
		SentAt:      time.Now(),
		Header:      header,
		BodyReader:  body,
	})
}
