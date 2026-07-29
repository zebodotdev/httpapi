package response

import (
	"net/http"
	"testing"

	callerpkg "github.com/zebodotdev/httpapi/caller"
	e "github.com/zebodotdev/httpapi/erreur"
	"github.com/zebodotdev/httpapi/param"
)

func TestRenderParamErrUsesCustomErrorFields(t *testing.T) {
	target := &testTarget{}
	RenderParamErr(target, &e.ErrInvalidParam{
		Param:   "media.hero_image",
		Mesg:    "wrong purpose",
		Code:    "file_reference_purpose_mismatch",
		Status:  http.StatusServiceUnavailable,
		Cause:   "service_unavailable",
		Type:    e.TypeTransient,
		FixCode: e.FixCodeRepeatSame,
	})

	if target.res.Status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", target.res.Status, http.StatusServiceUnavailable)
	}

	body, ok := target.res.Body.(ErrRes)
	if !ok {
		t.Fatalf("body = %T, want ErrRes", target.res.Body)
	}
	if body.Err.Code != "file_reference_purpose_mismatch" {
		t.Fatalf("code = %q", body.Err.Code)
	}
	if body.Err.Cause != "service_unavailable" || body.Err.Type != e.TypeTransient || body.Err.FixCode != e.FixCodeRepeatSame {
		t.Fatalf("unexpected error fields: %#v", body.Err)
	}
}

func TestRenderParamErrorUsesStandardErrorFields(t *testing.T) {
	target := &testTarget{}
	RenderParamError(target, param.NewError(
		"customer.email",
		param.CodeMissing,
		"`customer.email` is required",
		nil,
	))

	if target.res.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", target.res.Status, http.StatusBadRequest)
	}

	body, ok := target.res.Body.(ErrRes)
	if !ok {
		t.Fatalf("body = %T, want ErrRes", target.res.Body)
	}
	if body.Err.Code != string(param.CodeMissing) {
		t.Fatalf("code = %q", body.Err.Code)
	}
	if body.Err.Message != "`customer.email` is required" {
		t.Fatalf("message = %q", body.Err.Message)
	}
	if body.Err.Detail != "customer.email" {
		t.Fatalf("detail = %q", body.Err.Detail)
	}
	if body.Err.Cause != e.CauseMissingParam || body.Err.Type != e.TypeInvalidParam {
		t.Fatalf("unexpected error fields: %#v", body.Err)
	}
}

func TestRenderErrUsesErreurResponseStatus(t *testing.T) {
	target := &testTarget{}
	RenderErr(target, e.MethodNotAllowed(http.MethodPost, http.MethodGet))

	if target.res.Status != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", target.res.Status, http.StatusMethodNotAllowed)
	}
	if target.res.ContentType != ApplicationJson {
		t.Fatalf("content_type = %q, want %q", target.res.ContentType, ApplicationJson)
	}

	body, ok := target.res.Body.(ErrRes)
	if !ok {
		t.Fatalf("body = %T, want ErrRes", target.res.Body)
	}
	if body.Err.Code != "method_not_allowed" {
		t.Fatalf("code = %q, want method_not_allowed", body.Err.Code)
	}
	if body.Err.Type != e.TypeInvalidRequest || body.Err.Cause != e.CauseMethodNotAllowed {
		t.Fatalf("unexpected error fields: %#v", body.Err)
	}
}

func TestRenderErrDefaultsNilToUnexpected(t *testing.T) {
	target := &testTarget{}
	RenderErr(target, nil)

	if target.res.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", target.res.Status, http.StatusInternalServerError)
	}

	body, ok := target.res.Body.(ErrRes)
	if !ok {
		t.Fatalf("body = %T, want ErrRes", target.res.Body)
	}
	if body.Err.Code != "request_failed" {
		t.Fatalf("code = %q, want request_failed", body.Err.Code)
	}
}

func TestErrorShapeDescribesAndProjectsStandardErrorBody(t *testing.T) {
	spec := Describe(ErrorShape)
	if spec.Type != TypeObject || len(spec.Attributes) != 1 {
		t.Fatalf("error shape spec = %#v", spec)
	}
	if spec.Attributes[0].Name != "error" || !spec.Attributes[0].Required {
		t.Fatalf("error attribute = %#v", spec.Attributes[0])
	}

	body := ErrorShape.ProjectForCaller(ErrorBody{Err: e.InvalidRequestBody()}, callerpkg.Caller{})
	errObject, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("projected error = %#v", body["error"])
	}
	if errObject["code"] != "invalid_request_body" {
		t.Fatalf("code = %#v", errObject["code"])
	}
	if errObject["message"] == nil {
		t.Fatalf("message missing from projected error: %#v", errObject)
	}
}
