package endpoint

import (
	"testing"
	"time"
)

func TestRouteSpecWithDefaults(t *testing.T) {
	defaults := RouteSpec{
		Backend: RouteBackend{
			Address:  "https://example.test",
			PathMode: RoutePathModeAppend,
			Timeout:  10 * time.Second,
		},
	}
	spec := RouteSpec{
		Backend: RouteBackend{
			PathMode: RoutePathModeConstant,
		},
	}

	got := spec.WithDefaults(defaults)
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
