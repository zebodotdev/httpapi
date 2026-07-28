package response

import (
	"fmt"

	callerpkg "github.com/zebodotdev/httpapi/caller"
)

// ObjectShape defines a JSON object response shape for values of type T.
//
// Object shapes are useful when a response is part of a documented endpoint
// contract or when individual attributes have caller availability rules. The
// shape does not mutate the source value; it projects the value into a
// map[string]any just before JSON encoding.
type ObjectShape[T any] struct {
	attributes []AcceptedAttribute[T]
}

// Object starts a JSON object response shape definition.
//
// Define shapes once, usually at package level beside the endpoint response
// view type, then reuse them from handlers and spec transcribers.
func Object[T any](attributes ...AcceptedAttribute[T]) *ObjectShape[T] {
	shape := &ObjectShape[T]{}
	for _, attribute := range attributes {
		shape.Attribute(attribute)
	}
	return shape
}

// Attribute adds a response attribute to this object shape.
//
// Attribute order is preserved in the shape metadata for transcribers, although
// JSON object encoding itself does not promise object key order.
func (shape *ObjectShape[T]) Attribute(attribute AcceptedAttribute[T]) *ObjectShape[T] {
	if shape == nil {
		panic("httpapi/response: nil object shape")
	}
	if attribute == nil {
		panic("httpapi/response: attribute is required")
	}
	shape.attributes = append(shape.attributes, attribute)
	return shape
}

// ProjectForCaller returns the JSON object visible to caller.
//
// It is primarily useful for tests, manual projection, and spec tooling. Normal
// endpoint handlers should prefer response.Body or ObjectShape.Body and let
// RenderJSON derive the caller from the request target.
func (shape *ObjectShape[T]) ProjectForCaller(value T, caller callerpkg.Caller) map[string]any {
	if shape == nil {
		panic("httpapi/response: nil object shape")
	}
	shape.validateAttributes()

	object := make(map[string]any, len(shape.attributes))
	for _, attribute := range shape.attributes {
		name, projected, ok := attribute.projectAttribute(caller, value)
		if ok {
			object[name] = projected
		}
	}
	return object
}

// Attributes returns this object shape's transcribable attribute descriptions.
//
// The returned slice is a snapshot. Callers can inspect each AttributeSpec to
// determine whether a field belongs in a public, internal, or caller-specific
// response document.
func (shape *ObjectShape[T]) Attributes() []AttributeSpec {
	if shape == nil {
		panic("httpapi/response: nil object shape")
	}
	shape.validateAttributes()

	attributes := make([]AttributeSpec, 0, len(shape.attributes))
	for _, attribute := range shape.attributes {
		attributes = append(attributes, attribute.attributeSpec())
	}
	return attributes
}

// Body returns a caller-aware body that RenderJSON can project from its target.
//
// This method is equivalent to response.Body(shape, value). It is convenient
// when reading from left to right: response.JSON(status, taskShape.Body(task)).
func (shape *ObjectShape[T]) Body(value T) ShapedBody[T] {
	return Body[T](shape, value)
}

func (shape *ObjectShape[T]) describeShape() ShapeSpec {
	return ShapeSpec{Type: TypeObject, Attributes: shape.Attributes()}
}

func (shape *ObjectShape[T]) projectShape(caller callerpkg.Caller, value T) any {
	return shape.ProjectForCaller(value, caller)
}

func (shape *ObjectShape[T]) validateAttributes() {
	names := make(map[string]struct{}, len(shape.attributes))
	for _, attribute := range shape.attributes {
		name := attribute.attributeName()
		if _, ok := names[name]; ok {
			panic(fmt.Sprintf("httpapi/response: duplicate attribute %q", name))
		}
		names[name] = struct{}{}
	}
}

// Body returns a caller-aware body that render and write helpers can project
// before JSON encoding.
//
// Use Body when the response value should be filtered through a Shape. If the
// render target exposes RequestCaller, the active caller is used automatically;
// otherwise the body is projected as though the caller is undefined.
func Body[T any](shape Shape[T], value T) ShapedBody[T] {
	if shape == nil {
		panic("httpapi/response: body shape is required")
	}
	return ShapedBody[T]{
		shape: shape,
		value: value,
	}
}

// ShapedBody is a caller-aware response body returned by Body.
//
// ShapedBody is safe to store in Res.Body. It delays projection until render or
// encode time, which lets the same response value be written differently for
// different callers.
type ShapedBody[T any] struct {
	shape Shape[T]
	value T
}

// ProjectResponse returns this body projected for caller.
//
// ProjectResponse is the low-level escape hatch used by Render and
// EncodeResponseBodyForCaller. Endpoint handlers normally do not call it
// directly.
func (body ShapedBody[T]) ProjectResponse(caller callerpkg.Caller) any {
	if body.shape == nil {
		panic("httpapi/response: nil body shape")
	}
	return body.shape.projectShape(caller, body.value)
}
