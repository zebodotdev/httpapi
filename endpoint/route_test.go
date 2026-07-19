package endpoint

import (
	"testing"
	"time"
)

func TestRouteSpecWithDefaults(t *testing.T) {
	defaults := RouteSpec{
		Summary: "default summary",
		Backend: RouteBackend{
			Address:  "https://example.test",
			PathMode: RoutePathModeAppend,
			Timeout:  10 * time.Second,
		},
	}
	spec := RouteSpec{
		OperationID: "operation",
		Backend: RouteBackend{
			PathMode: RoutePathModeConstant,
		},
	}

	got := spec.WithDefaults(defaults)
	if got.OperationID != "operation" {
		t.Fatalf("operation_id = %q, want operation", got.OperationID)
	}
	if got.Summary != defaults.Summary {
		t.Fatalf("summary = %q, want %q", got.Summary, defaults.Summary)
	}
	if got.Backend.Address != defaults.Backend.Address {
		t.Fatalf("backend address = %q, want %q", got.Backend.Address, defaults.Backend.Address)
	}
	if got.Backend.PathMode != RoutePathModeConstant {
		t.Fatalf("path mode = %q, want %q", got.Backend.PathMode, RoutePathModeConstant)
	}
	if got.Backend.Timeout != defaults.Backend.Timeout {
		t.Fatalf("timeout = %s, want %s", got.Backend.Timeout, defaults.Backend.Timeout)
	}
}
