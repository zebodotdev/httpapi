package endpoint

import "testing"

func TestNormalizePriority(t *testing.T) {
	if got := NormalizePriority(" HIGH "); got != PriorityHigh {
		t.Fatalf("priority = %q, want %q", got, PriorityHigh)
	}
}
