package endpoint

// EndpointPriority captures how important an endpoint is to service operations.
type EndpointPriority = Priority

const (
	// EndpointPriorityCritical marks endpoints that are essential to core
	// service operation or revenue-critical request paths.
	EndpointPriorityCritical EndpointPriority = PriorityCritical

	// EndpointPriorityHigh marks important endpoints that should be favored in
	// operational reviews and gateway configuration.
	EndpointPriorityHigh EndpointPriority = PriorityHigh

	// EndpointPriorityStandard marks normal production endpoints.
	EndpointPriorityStandard EndpointPriority = PriorityStandard

	// EndpointPriorityLow marks endpoints that are less operationally sensitive.
	EndpointPriorityLow EndpointPriority = PriorityLow
)

type endpointPriorityPolicy struct {
	priority  EndpointPriority
	inherited bool
}

// WithPriority marks the endpoint with an operational priority.
//
// Endpoint-level priorities override any priority inherited from an
// EndpointGroup.
func WithPriority(priority EndpointPriority) EndpointOption {
	priority = requiredEndpointPriority(priority)
	return func(e *Endpoint) {
		e.priority.priority = priority
		e.priority.inherited = false
	}
}

// SetPriority sets the default operational priority for endpoints in the group.
//
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
	return RequiredPriority(priority)
}

func normalizeEndpointPriority(priority EndpointPriority) EndpointPriority {
	return NormalizePriority(priority)
}

func (e Endpoint) priorityPolicy() endpointPriorityPolicy {
	policy := e.priority
	policy.priority = normalizeEndpointPriority(policy.priority)
	return policy
}

func (e *Endpoint) mutablePriorityPolicy() *endpointPriorityPolicy {
	return &e.priority
}
