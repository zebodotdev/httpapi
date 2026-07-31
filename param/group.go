package param

// ParamGroup declares multiple parameters and the presence rule that relates
// them. Add it to JSON or Object builders with Param.
type ParamGroup struct {
	params []AcceptedParam
	rule   presenceRule
}

// AtLeastOneOf declares parameters and requires callers to provide one or more.
func AtLeastOneOf(parameters ...AcceptedParam) *ParamGroup {
	return newParamGroup(1, 0, parameters...)
}

// OneOf declares parameters and requires callers to provide exactly one.
func OneOf(parameters ...AcceptedParam) *ParamGroup {
	return newParamGroup(1, 1, parameters...)
}

// MutuallyExclusive declares parameters that may not be provided together.
//
// The group is optional by default: callers may provide none of the parameters
// or exactly one of them. Call Required when the group must provide exactly one
// value.
func MutuallyExclusive(parameters ...AcceptedParam) *ParamGroup {
	return newParamGroup(0, 1, parameters...)
}

// Required makes a parameter group mandatory.
//
// For a MutuallyExclusive group, Required means callers must provide exactly one
// of the grouped parameters.
func (group *ParamGroup) Required() *ParamGroup {
	if group == nil {
		panic("httpapi/param: parameter group is required")
	}
	group.rule.minPresent = 1
	group.rule = normalizePresenceRule(group.rule)
	return group
}

func newParamGroup(
	minPresent int,
	maxPresent int,
	parameters ...AcceptedParam,
) *ParamGroup {
	collector := &paramGroupCollector{}
	for _, parameter := range parameters {
		if parameter == nil {
			panic("httpapi/param: parameter group parameter is required")
		}
		parameter.addToParamSet(collector)
	}
	if len(collector.rules) > 0 {
		panic("httpapi/param: parameter groups cannot contain parameter groups")
	}

	names := make([]string, 0, len(collector.params))
	params := make([]AcceptedParam, 0, len(collector.params))
	for _, parameter := range collector.params {
		names = append(names, parameter.paramName())
		params = append(params, parameter)
	}

	return &ParamGroup{
		params: params,
		rule: normalizePresenceRule(presenceRule{
			paramNames: names,
			minPresent: minPresent,
			maxPresent: maxPresent,
		}),
	}
}

func (group *ParamGroup) addToParamSet(set paramSet) {
	if group == nil {
		panic("httpapi/param: parameter group is required")
	}
	for _, parameter := range group.params {
		set.addAcceptedParam(parameter)
	}
	set.addRule(group.rule)
}

type paramGroupCollector struct {
	params []AcceptedParam
	rules  []rule
}

func (collector *paramGroupCollector) addAcceptedParam(parameter AcceptedParam) {
	collector.params = append(collector.params, parameter)
}

func (collector *paramGroupCollector) addRule(rule rule) {
	collector.rules = append(collector.rules, rule)
}
