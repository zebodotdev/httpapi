package response

import (
	"time"

	callerpkg "github.com/zebodotdev/httpapi/caller"
)

// Type identifies the JSON type emitted by a response shape.
type Type string

const (
	// TypeAny describes a response attribute with no narrower declared type.
	TypeAny Type = "any"

	// TypeArray describes a JSON array response attribute.
	TypeArray Type = "array"

	// TypeBool describes a JSON boolean response attribute.
	TypeBool Type = "boolean"

	// TypeFloat64 describes a JSON number response attribute.
	TypeFloat64 Type = "number"

	// TypeInt describes a JSON integer response attribute.
	TypeInt Type = "integer"

	// TypeInt64 describes a JSON integer response attribute.
	TypeInt64 Type = "integer"

	// TypeObject describes a JSON object response attribute.
	TypeObject Type = "object"

	// TypeString describes a JSON string response attribute.
	TypeString Type = "string"

	// TypeTime describes a JSON string timestamp response attribute.
	TypeTime Type = "string"
)

// ShapeSpec is the transcribable description of a response shape.
//
// ShapeSpec is deliberately provider-neutral. OpenAPI and gateway writers can
// translate it into their own schema objects without importing runtime response
// internals or re-reading application serializers.
type ShapeSpec struct {
	// Type is the JSON type emitted by the shape.
	Type Type

	// Attributes describes object attributes when Type is TypeObject.
	Attributes []AttributeSpec

	// Item describes array items when Type is TypeArray.
	Item *ShapeSpec
}

// Shape describes the JSON value emitted by a response attribute.
//
// Shapes are used at render time to project caller-aware response bodies and at
// transcription time to expose the same response contract to OpenAPI writers.
type Shape[T any] interface {
	describeShape() ShapeSpec
	projectShape(callerpkg.Caller, T) any
}

type scalarShape[T any] struct {
	typ Type
}

func (shape scalarShape[T]) describeShape() ShapeSpec {
	return ShapeSpec{Type: shape.typ}
}

func (shape scalarShape[T]) projectShape(_ callerpkg.Caller, value T) any {
	return value
}

// String emits a JSON string attribute.
func String() Shape[string] { return scalarShape[string]{typ: TypeString} }

// Int emits a JSON integer attribute.
func Int() Shape[int] { return scalarShape[int]{typ: TypeInt} }

// Int64 emits a JSON integer attribute from an int64 value.
func Int64() Shape[int64] { return scalarShape[int64]{typ: TypeInt64} }

// Float64 emits a JSON number attribute.
func Float64() Shape[float64] { return scalarShape[float64]{typ: TypeFloat64} }

// Bool emits a JSON boolean attribute.
func Bool() Shape[bool] { return scalarShape[bool]{typ: TypeBool} }

// Time emits a JSON timestamp attribute.
//
// The standard JSON encoder formats the value with time.Time's
// RFC3339-compatible JSON representation. Spec writers can treat this as a
// string and apply their target's timestamp format convention.
func Time() Shape[time.Time] { return scalarShape[time.Time]{typ: TypeTime} }

// Any emits a JSON attribute without additional shaping.
//
// Prefer narrower shapes when possible. Any is best for intentionally flexible
// blobs such as custom metadata where the service does not promise a fixed
// object layout.
func Any[T any]() Shape[T] { return scalarShape[T]{typ: TypeAny} }

// Array emits a JSON array whose items do not need caller-aware projection.
//
// Use Array for primitive or already-public item values. Use ArrayOf when each
// item has its own response shape or caller-sensitive attributes.
func Array[T any]() Shape[[]T] { return arrayShape[T]{} }

type arrayShape[T any] struct{}

func (shape arrayShape[T]) describeShape() ShapeSpec {
	return ShapeSpec{Type: TypeArray, Item: &ShapeSpec{Type: TypeAny}}
}

func (shape arrayShape[T]) projectShape(_ callerpkg.Caller, values []T) any {
	return values
}

// ArrayOf emits a JSON array and projects each item through itemShape.
//
// ArrayOf is the collection equivalent of ObjectShape.Body: each item is
// projected for the active caller before the final JSON response is encoded.
func ArrayOf[T any](itemShape Shape[T]) Shape[[]T] {
	if itemShape == nil {
		panic("httpapi/response: array item shape is required")
	}
	return arrayOfShape[T]{itemShape: itemShape}
}

type arrayOfShape[T any] struct {
	itemShape Shape[T]
}

func (shape arrayOfShape[T]) describeShape() ShapeSpec {
	item := shape.itemShape.describeShape()
	return ShapeSpec{Type: TypeArray, Item: &item}
}

func (shape arrayOfShape[T]) projectShape(caller callerpkg.Caller, values []T) any {
	projected := make([]any, 0, len(values))
	for _, value := range values {
		projected = append(projected, shape.itemShape.projectShape(caller, value))
	}
	return projected
}

// Describe returns the transcribable description of shape.
//
// Runtime handlers usually do not need Describe. It exists for tests, docs, and
// transcribers that need to turn response contracts into schema documents.
func Describe[T any](shape Shape[T]) ShapeSpec {
	if shape == nil {
		panic("httpapi/response: shape is required")
	}
	return shape.describeShape()
}
