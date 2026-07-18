package httpapi

import (
	"net/http"
	"testing"

	e "github.com/zebodotdev/httpapi/erreur"
)

func TestRenderParamErrUsesCustomErrorFields(t *testing.T) {
	req := &Req{}
	RenderParamErr(req, &e.ErrInvalidParam{
		Param:   "media.hero_image",
		Mesg:    "wrong purpose",
		Code:    "file_reference_purpose_mismatch",
		Status:  http.StatusServiceUnavailable,
		Cause:   "service_unavailable",
		Type:    e.TypeTransient,
		FixCode: e.FixCodeRepeatSame,
	})

	if req.Res.Status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", req.Res.Status, http.StatusServiceUnavailable)
	}

	body, ok := req.Res.Body.(ErrRes)
	if !ok {
		t.Fatalf("body = %T, want ErrRes", req.Res.Body)
	}
	if body.Err.Code != "file_reference_purpose_mismatch" {
		t.Fatalf("code = %q", body.Err.Code)
	}
	if body.Err.Cause != "service_unavailable" || body.Err.Type != e.TypeTransient || body.Err.FixCode != e.FixCodeRepeatSame {
		t.Fatalf("unexpected error fields: %#v", body.Err)
	}
}

func TestRenderErrUsesErreurResponseStatus(t *testing.T) {
	req := &Req{}
	RenderErr(req, e.MethodNotAllowed(POST, GET))

	if req.Res.Status != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", req.Res.Status, http.StatusMethodNotAllowed)
	}
	if req.Res.ContentType != ApplicationJson {
		t.Fatalf("content_type = %q, want %q", req.Res.ContentType, ApplicationJson)
	}

	body, ok := req.Res.Body.(ErrRes)
	if !ok {
		t.Fatalf("body = %T, want ErrRes", req.Res.Body)
	}
	if body.Err.Code != "method_not_allowed" {
		t.Fatalf("code = %q, want method_not_allowed", body.Err.Code)
	}
	if body.Err.Type != e.TypeInvalidRequest || body.Err.Cause != e.CauseMethodNotAllowed {
		t.Fatalf("unexpected error fields: %#v", body.Err)
	}
}

func TestRenderErrDefaultsNilToUnexpected(t *testing.T) {
	req := &Req{}
	RenderErr(req, nil)

	if req.Res.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", req.Res.Status, http.StatusInternalServerError)
	}

	body, ok := req.Res.Body.(ErrRes)
	if !ok {
		t.Fatalf("body = %T, want ErrRes", req.Res.Body)
	}
	if body.Err.Code != "request_failed" {
		t.Fatalf("code = %q, want request_failed", body.Err.Code)
	}
}
