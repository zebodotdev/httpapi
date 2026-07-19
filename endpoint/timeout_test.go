package endpoint

import (
	"testing"
	"time"
)

func TestNormalizeTimeoutSpec(t *testing.T) {
	spec := TimeoutSpec{
		ReadBody: time.Second,
		Handler:  2 * time.Second,
		Write:    3 * time.Second,
	}
	if got := NormalizeTimeoutSpec(spec); got != spec {
		t.Fatalf("timeout spec = %#v, want %#v", got, spec)
	}
}
