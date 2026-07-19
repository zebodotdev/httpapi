package httpapi

import endpointpkg "github.com/zebodotdev/httpapi/endpoint"

// EndpointPriority captures how important an endpoint is to service operations.
type EndpointPriority = endpointpkg.Priority

const (
	EndpointPriorityCritical EndpointPriority = endpointpkg.PriorityCritical
	EndpointPriorityHigh     EndpointPriority = endpointpkg.PriorityHigh
	EndpointPriorityStandard EndpointPriority = endpointpkg.PriorityStandard
	EndpointPriorityLow      EndpointPriority = endpointpkg.PriorityLow
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
	return endpointpkg.RequiredPriority(priority)
}

func normalizeEndpointPriority(priority EndpointPriority) EndpointPriority {
	return endpointpkg.NormalizePriority(priority)
}

func (e Endpoint) priorityPolicy() endpointPriorityPolicy {
	policy := e.priority
	policy.priority = normalizeEndpointPriority(policy.priority)
	return policy
}

func (e *Endpoint) mutablePriorityPolicy() *endpointPriorityPolicy {
	return &e.priority
}
