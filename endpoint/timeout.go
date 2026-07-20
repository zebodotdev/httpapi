package endpoint

import (
	"fmt"
	"time"
)

// TimeoutSpec declares runtime timeout budgets for one endpoint.
//
// These budgets are enforced by the in-process HTTP wrapper. They are separate
// from RouteBackend.Timeout, which only describes generated gateway/backend
// deadlines for transcribed route specs.
type TimeoutSpec struct {
	// ReadBody bounds how long httpapi allows for reading the inbound request
	// body. Zero disables the read-body deadline.
	ReadBody time.Duration

	// Handler bounds how long the application handler may run before its
	// context is cancelled. Zero disables the handler deadline.
	Handler time.Duration

	// Write bounds how long httpapi allows for writing the response. Zero
	// disables the write deadline.
	Write time.Duration
}

// NormalizeTimeoutSpec validates endpoint runtime timeouts and returns the
// normalized value.
func NormalizeTimeoutSpec(spec TimeoutSpec) TimeoutSpec {
	if spec.ReadBody < 0 {
		panic(fmt.Sprintf(
			"httpapi: endpoint read body timeout cannot be negative: %s",
			spec.ReadBody,
		))
	}
	if spec.Handler < 0 {
		panic(fmt.Sprintf(
			"httpapi: endpoint handler timeout cannot be negative: %s",
			spec.Handler,
		))
	}
	if spec.Write < 0 {
		panic(fmt.Sprintf(
			"httpapi: endpoint write timeout cannot be negative: %s",
			spec.Write,
		))
	}

	return spec
}
