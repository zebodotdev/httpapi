package response

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestWriteResponseWritesJSONResponse(t *testing.T) {
	res := &Res{
		ContentType: ApplicationJson,
		Status:      http.StatusCreated,
		Header:      http.Header{"X-Custom": []string{"yes"}},
		Body:        map[string]string{"ok": "true"},
	}
	rec := httptest.NewRecorder()

	result, err := WriteResponse(rec, res, WriteOptions{
		RequestID: "req_123",
		Duration:  12 * time.Millisecond,
	})

	if err != nil {
		t.Fatalf("WriteResponse error = %v", err)
	}
	if result.Status != http.StatusCreated {
		t.Fatalf("status = %d, want %d", result.Status, http.StatusCreated)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("response code = %d, want %d", rec.Code, http.StatusCreated)
	}
	if got := rec.Header().Get(contentTypeHeaderKey); got != ApplicationJson {
		t.Fatalf("content-type = %q, want %q", got, ApplicationJson)
	}
	if got := rec.Header().Get(xReqIDHeaderKey); got != "req_123" {
		t.Fatalf("request id header = %q, want req_123", got)
	}
	if got := rec.Header().Get("X-Custom"); got != "yes" {
		t.Fatalf("custom header = %q, want yes", got)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"ok":"true"`) {
		t.Fatalf("body = %s, want JSON body", body)
	}
	if result.BytesWritten <= 0 {
		t.Fatalf("bytes written = %d, want positive", result.BytesWritten)
	}
}

func TestWriteResponseDoesNotWriteCORSDefaults(t *testing.T) {
	res := JSON(http.StatusOK, map[string]bool{"ok": true})
	rec := httptest.NewRecorder()

	_, err := WriteResponse(rec, res, WriteOptions{RequestID: "req_123"})

	if err != nil {
		t.Fatalf("WriteResponse error = %v", err)
	}
	for _, name := range []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
	} {
		if got := rec.Header().Get(name); got != "" {
			t.Fatalf("%s = %q, want empty", name, got)
		}
	}
}

func TestWriteResponseWritesPlainTextString(t *testing.T) {
	res := &Res{
		ContentType: TextPlain,
		Status:      http.StatusAccepted,
		Header:      http.Header{},
		Body:        "accepted",
	}
	rec := httptest.NewRecorder()

	result, err := WriteResponse(rec, res, WriteOptions{RequestID: "req_123"})

	if err != nil {
		t.Fatalf("WriteResponse error = %v", err)
	}
	if result.Status != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", result.Status, http.StatusAccepted)
	}
	if rec.Body.String() != "accepted" {
		t.Fatalf("body = %q, want accepted", rec.Body.String())
	}
}

func TestWriteResponseStreamsAndClosesBody(t *testing.T) {
	body := &trackedReadCloser{Reader: strings.NewReader("streamed")}
	res := &Res{
		ContentType: TextPlain,
		Status:      http.StatusOK,
		Header:      http.Header{},
		BodyReader:  body,
	}
	rec := httptest.NewRecorder()

	result, err := WriteResponse(rec, res, WriteOptions{RequestID: "req_123"})

	if err != nil {
		t.Fatalf("WriteResponse error = %v", err)
	}
	if !result.Streamed {
		t.Fatal("streamed = false, want true")
	}
	if result.BytesWritten != len("streamed") {
		t.Fatalf("bytes written = %d, want %d", result.BytesWritten, len("streamed"))
	}
	if rec.Body.String() != "streamed" {
		t.Fatalf("body = %q, want streamed", rec.Body.String())
	}
	if !body.closed {
		t.Fatal("stream body was not closed")
	}
}

func TestWriteResponseWritesNoContentWithoutContentType(t *testing.T) {
	res := NoContent()
	rec := httptest.NewRecorder()

	result, err := WriteResponse(rec, res, WriteOptions{RequestID: "req_123"})

	if err != nil {
		t.Fatalf("WriteResponse error = %v", err)
	}
	if result.Status != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", result.Status, http.StatusNoContent)
	}
	if result.BytesWritten != 0 {
		t.Fatalf("bytes written = %d, want 0", result.BytesWritten)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body length = %d, want 0", rec.Body.Len())
	}
	if got := rec.Header().Get(contentTypeHeaderKey); got != "" {
		t.Fatalf("content-type = %q, want empty", got)
	}
}

func TestWriteResponseStoresWrittenHeaderSnapshot(t *testing.T) {
	res := JSON(http.StatusOK, map[string]string{"ok": "true"}, WithHeader("X-Custom", "yes"))
	rec := httptest.NewRecorder()

	_, err := WriteResponse(rec, res, WriteOptions{RequestID: "req_123"})

	if err != nil {
		t.Fatalf("WriteResponse error = %v", err)
	}
	rec.Header().Set("X-Custom", "changed")
	if !reflect.DeepEqual(res.Header.Values("X-Custom"), []string{"yes"}) {
		t.Fatalf("response header snapshot = %#v", res.Header)
	}
	if res.Header.Get(xReqIDHeaderKey) != "req_123" {
		t.Fatalf("request id header = %q", res.Header.Get(xReqIDHeaderKey))
	}
}

func TestWriteResponseReturnsBeforeHeadersOnEncodeError(t *testing.T) {
	res := &Res{
		ContentType: ApplicationJson,
		Status:      http.StatusOK,
		Header:      http.Header{},
		Body:        make(chan int),
	}
	rec := httptest.NewRecorder()

	result, err := WriteResponse(rec, res, WriteOptions{RequestID: "req_123"})

	if err == nil {
		t.Fatal("WriteResponse error = nil, want encode error")
	}
	if result.Status != 0 {
		t.Fatalf("result status = %d, want 0 before headers", result.Status)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body length = %d, want 0", rec.Body.Len())
	}
}

type trackedReadCloser struct {
	*strings.Reader
	closed bool
}

func (r *trackedReadCloser) Close() error {
	r.closed = true
	return nil
}
