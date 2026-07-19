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
	ReadBody time.Duration
	Handler  time.Duration
	Write    time.Duration
}

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
