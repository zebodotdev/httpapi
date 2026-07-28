package request

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	callerpkg "github.com/zebodotdev/httpapi/caller"
)

func TestNewReqAttachesCallerFromContext(t *testing.T) {
	worker := callerpkg.Define("worker")
	req := httptest.NewRequest("POST", "/records/new", strings.NewReader(`{}`)).
		WithContext(ContextWithCaller(context.Background(), worker))

	parsed := NewReq(req)
	if parsed == nil {
		t.Fatal("NewReq returned nil")
	}
	if parsed.Caller != worker {
		t.Fatalf("caller = %q, want %q", parsed.Caller, worker)
	}
	if parsed.RequestCaller() != worker {
		t.Fatalf("request caller = %q, want %q", parsed.RequestCaller(), worker)
	}
}
