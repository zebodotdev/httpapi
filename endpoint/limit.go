package endpoint

import "fmt"

// LimitsSpec declares runtime request limits for one endpoint.
//
// A zero value disables endpoint-level limits. Services should still configure
// a server-level header cap so very large headers are rejected before routing.
type LimitsSpec struct {
	// MaxRequestBytes caps the full parsed HTTP request envelope accepted by an
	// endpoint. The limit includes the request line, headers, the blank line
	// after headers, and body bytes.
	MaxRequestBytes int64
}

// NormalizeLimitsSpec validates endpoint runtime limits and returns the
// normalized value.
func NormalizeLimitsSpec(spec LimitsSpec) LimitsSpec {
	if spec.MaxRequestBytes < 0 {
		panic(fmt.Sprintf(
			"httpapi: endpoint max request bytes cannot be negative: %d",
			spec.MaxRequestBytes,
		))
	}

	return spec
}
