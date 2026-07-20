package request

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zebodotdev/httpapi/response"
)

func TestNewReqParsesBodyAndResetsReader(t *testing.T) {
	httpReq := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"ok":true}`))
	httpReq.Header.Set(contentTypeHeaderKey, "application/json")

	req := NewReq(httpReq)

	if req == nil {
		t.Fatal("NewReq returned nil")
	}
	if string(req.Body) != `{"ok":true}` {
		t.Fatalf("body = %q, want buffered request body", req.Body)
	}

	replay := make([]byte, len(req.Body))
	if _, err := req.Req.Body.Read(replay); err != nil {
		t.Fatalf("reading reset body: %v", err)
	}
	if string(replay) != string(req.Body) {
		t.Fatalf("reset body = %q, want %q", replay, req.Body)
	}
}

func TestNewReqWithErrorReturnsBodyReadError(t *testing.T) {
	want := errors.New("read failed")
	httpReq := httptest.NewRequest(http.MethodPost, "/tasks", nil)
	httpReq.Body = errReadCloser{err: want}

	req, err := NewReqWithError(httpReq)

	if req != nil {
		t.Fatalf("req = %#v, want nil", req)
	}
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}

func TestReqImplementsResponseTarget(t *testing.T) {
	req := &Req{}
	res := &response.Res{
		ContentType: response.ApplicationJson,
		Status:      http.StatusAccepted,
		Header:      http.Header{},
		Body:        map[string]bool{"ok": true},
	}

	req.SetResponse(res)

	if req.Response() != res {
		t.Fatal("response target did not retain response")
	}
}

type errReadCloser struct {
	err error
}

func (r errReadCloser) Read([]byte) (int, error) {
	return 0, r.err
}

func (r errReadCloser) Close() error {
	return nil
}

var _ io.ReadCloser = errReadCloser{}
