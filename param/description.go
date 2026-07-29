package param

import callerpkg "github.com/zebodotdev/httpapi/caller"

// RequestSpec is the transcribable description of a JSON request parser.
//
// RequestSpec is derived from the same parameter definitions used at runtime,
// so generated documents and request parsing share one contract source.
type RequestSpec struct {
	// Body describes the top-level JSON request body.
	Body ShapeSpec
}

// ShapeSpec is the transcribable description of an accepted JSON wire shape.
type ShapeSpec struct {
	// Type is the JSON type accepted by the shape before custom parsing.
	Type Type

	// Parameters describes object parameters when Type is TypeObject.
	Parameters []ParameterSpec

	// Rules describes object-level presence rules when Type is TypeObject.
	Rules []RuleSpec

	// Item describes array items when Type is TypeArray.
	Item *ShapeSpec
}

// ParameterSpec describes one accepted request parameter.
type ParameterSpec struct {
	// Name is the JSON key accepted for the parameter.
	Name string

	// Required reports whether the parameter must be present for callers that
	// are allowed to supply it.
	Required bool

	// NullPolicy describes how a present JSON null is handled.
	NullPolicy NullPolicy

	// Availability is the caller allow-list for this parameter. The zero value is
	// unrestricted.
	Availability callerpkg.Set

	// Shape describes the JSON value accepted for this parameter.
	Shape ShapeSpec

	// MinSize and MaxSize are inclusive bounds for strings, arrays, objects, and
	// numeric values when configured.
	MinSize *int64
	MaxSize *int64

	// MinItems and MaxItems are inclusive item-count bounds for array parameters
	// when configured.
	MinItems *int64
	MaxItems *int64
}

// RuleSpec describes an object-level presence rule over a parameter set.
type RuleSpec struct {
	// Names is the ordered parameter set participating in the rule.
	Names []string

	// MinPresent is the minimum number of parameters that must be present. Zero
	// means there is no lower bound.
	MinPresent int

	// MaxPresent is the maximum number of parameters that may be present. Zero
	// means there is no upper bound.
	MaxPresent int
}

// Describe returns the transcribable description of request.
func Describe[T any](request *Request[T]) RequestSpec {
	if request == nil || request.def == nil {
		panic("httpapi/param: nil request parser")
	}
	return RequestSpec{Body: request.def.describeShape()}
}

// Describe returns the transcribable description of request.
func (request *Request[T]) Describe() RequestSpec {
	return Describe(request)
}

func (parameter *Parameter[T]) parameterSpec() ParameterSpec {
	return ParameterSpec{
		Name:         parameter.name,
		Required:     parameter.required,
		NullPolicy:   normalizeNullPolicy(parameter.nullPolicy),
		Availability: cloneCallerSet(parameter.availability),
		Shape:        parameter.shape.describeShape(),
		MinSize:      cloneInt64Pointer(parameter.minSize),
		MaxSize:      cloneInt64Pointer(parameter.maxSize),
		MinItems:     cloneInt64Pointer(parameter.minItems),
		MaxItems:     cloneInt64Pointer(parameter.maxItems),
	}
}

func cloneCallerSet(set callerpkg.Set) callerpkg.Set {
	if !set.Restricted() {
		return callerpkg.Set{}
	}
	return callerpkg.SetOf(set.Callers()...)
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
