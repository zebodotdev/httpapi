package endpoint

import (
	"fmt"
	"strings"
)

// Priority captures how important an endpoint is to service operations.
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityStandard Priority = "standard"
	PriorityLow      Priority = "low"
)

func RequiredPriority(priority Priority) Priority {
	priority = NormalizePriority(priority)
	if priority == "" {
		panic("httpapi: endpoint priority is required")
	}

	return priority
}

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
