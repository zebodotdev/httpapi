package endpoint

import (
	"fmt"
	"strings"
)

// Priority captures how important an endpoint is to service operations.
type Priority string

const (
	// PriorityCritical marks endpoints that are essential to core service
	// operation or revenue-critical request paths.
	PriorityCritical Priority = "critical"

	// PriorityHigh marks important endpoints that should be favored in
	// operational reviews and gateway configuration.
	PriorityHigh Priority = "high"

	// PriorityStandard marks normal production endpoints.
	PriorityStandard Priority = "standard"

	// PriorityLow marks endpoints that are less operationally sensitive.
	PriorityLow Priority = "low"
)

// RequiredPriority returns a normalized priority and panics when priority is
// empty.
func RequiredPriority(priority Priority) Priority {
	priority = NormalizePriority(priority)
	if priority == "" {
		panic("httpapi: endpoint priority is required")
	}

	return priority
}

// NormalizePriority trims and canonicalizes an endpoint priority.
//
// Unknown non-empty priorities panic so endpoint definitions fail at startup
// instead of producing ambiguous generated specs.
func NormalizePriority(priority Priority) Priority {
	switch strings.TrimSpace(strings.ToLower(string(priority))) {
	case "":
		return ""
	case string(PriorityCritical):
		return PriorityCritical
	case string(PriorityHigh):
		return PriorityHigh
	case string(PriorityStandard):
		return PriorityStandard
	case string(PriorityLow):
		return PriorityLow
	default:
		panic(fmt.Sprintf("httpapi: unsupported endpoint priority %q", priority))
	}
}
