package response

import (
	"io"
	"net/http"
	"time"
)

// Content type aliases kept here so response rendering does not depend on the
// root httpapi package.
const (
	// ApplicationJson is the JSON response content type used by RenderJSON and
	// RenderErr.
	ApplicationJson = "application/json"

	// TextHTML is the HTML response content type supported by EncodeResponseBody
	// when the response body is a string.
	TextHTML = "text/html"

	// TextPlain is the plain text response content type supported by
	// EncodeResponseBody when the response body is a string.
	TextPlain = "text/plain; charset=utf-8"
)

// Res is the response object populated by handlers and written by endpoint
// runtimes.
type Res struct {
	// ContentType is the HTTP Content-Type header value.
	ContentType string `json:"content_type"`

	// Status is the HTTP response status code.
	Status int `json:"status"`

	// SentAt records when the response object was created.
	SentAt time.Time `json:"sent_at"`

	// Header contains additional response headers. WriteResponse also writes
	// httpapi's standard headers.
	Header http.Header `json:"header"`

	// Body is the JSON, text, or HTML body for non-streaming responses.
	Body any `json:"body"`

	// BodyReader streams response bytes directly to the client. When BodyReader
	// is set, Body is ignored by WriteResponse.
	BodyReader io.Reader `json:"-"`
}

// Target is the minimum request surface response renderers need.
type Target interface {
	// SetResponse stores the response that should be written for the current
	// request.
	SetResponse(*Res)
}

// RenderJSON sets a JSON response on the target.
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
//
// The caller owns the reader until WriteResponse starts writing it; if the
// reader implements io.Closer, WriteResponseStream closes it after copying.
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
