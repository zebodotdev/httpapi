package httpapi

import (
	"fmt"
	"strings"
)

// EndpointPriority captures how important an endpoint is to service operations.
type EndpointPriority string

const (
	EndpointPriorityCritical EndpointPriority = "critical"
	EndpointPriorityHigh     EndpointPriority = "high"
	EndpointPriorityStandard EndpointPriority = "standard"
	EndpointPriorityLow      EndpointPriority = "low"
)

type endpointPriorityPolicy struct {
	priority  EndpointPriority
	inherited bool
}

// WithPriority marks the endpoint with an operational priority.
func WithPriority(priority EndpointPriority) EndpointOption {
	priority = requiredEndpointPriority(priority)
	return func(e *Endpoint) {
		e.priority.priority = priority
		e.priority.inherited = false
	}
}

// SetPriority sets the default operational priority for endpoints in the group.
// Endpoint-level priorities override the group default.
func (eg *EndpointGroup) SetPriority(priority EndpointPriority) {
	eg.Priority = requiredEndpointPriority(priority)
	for i := range eg.Endpoints {
		policy := eg.Endpoints[i].mutablePriorityPolicy()
		if policy.priority != "" && !policy.inherited {
			continue
		}
		policy.priority = eg.Priority
		policy.inherited = true
	}
}

func requiredEndpointPriority(priority EndpointPriority) EndpointPriority {
	priority = normalizeEndpointPriority(priority)
	if priority == "" {
		panic("httpapi: endpoint priority is required")
	}

	return priority
}

func normalizeEndpointPriority(priority EndpointPriority) EndpointPriority {
	switch strings.TrimSpace(strings.ToLower(string(priority))) {
	case "":
		return ""
	case string(EndpointPriorityCritical):
		return EndpointPriorityCritical
	case string(EndpointPriorityHigh):
		return EndpointPriorityHigh
	case string(EndpointPriorityStandard):
		return EndpointPriorityStandard
	case string(EndpointPriorityLow):
		return EndpointPriorityLow
	default:
		panic(fmt.Sprintf("httpapi: unsupported endpoint priority %q", priority))
	}
}

func (e Endpoint) priorityPolicy() endpointPriorityPolicy {
	policy := e.priority
	policy.priority = normalizeEndpointPriority(policy.priority)
	return policy
}

func (e *Endpoint) mutablePriorityPolicy() *endpointPriorityPolicy {
	return &e.priority
}
