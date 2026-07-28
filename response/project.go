package response

import callerpkg "github.com/zebodotdev/httpapi/caller"

// Project adapts a response shape written for TPrepared so it can render from
// TSource.
//
// Use Project when a domain value needs a small response-facing view before its
// attributes are extracted. The prepare function runs once per projected source
// value, then the wrapped shape renders that prepared value for the active
// caller.
func Project[TSource any, TPrepared any](
	shape Shape[TPrepared],
	prepare func(TSource) TPrepared,
) *ProjectedShape[TSource, TPrepared] {
	if shape == nil {
		panic("httpapi/response: projected shape is required")
	}
	if prepare == nil {
		panic("httpapi/response: projection prepare function is required")
	}
	return &ProjectedShape[TSource, TPrepared]{
		shape:   shape,
		prepare: prepare,
	}
}

// ProjectedShape is a response shape that prepares source values before
// projecting them through another shape.
type ProjectedShape[TSource any, TPrepared any] struct {
	shape   Shape[TPrepared]
	prepare func(TSource) TPrepared
}

// ProjectForCaller returns the JSON value visible to caller.
func (shape *ProjectedShape[TSource, TPrepared]) ProjectForCaller(
	value TSource,
	caller callerpkg.Caller,
) any {
	if shape == nil {
		panic("httpapi/response: nil projected shape")
	}
	return shape.projectShape(caller, value)
}

// Body returns a caller-aware body that RenderJSON can project from its target.
func (shape *ProjectedShape[TSource, TPrepared]) Body(value TSource) ShapedBody[TSource] {
	return Body[TSource](shape, value)
}

func (shape *ProjectedShape[TSource, TPrepared]) describeShape() ShapeSpec {
	if shape == nil || shape.shape == nil {
		panic("httpapi/response: nil projected shape")
	}
	return shape.shape.describeShape()
}

func (shape *ProjectedShape[TSource, TPrepared]) projectShape(caller callerpkg.Caller, value TSource) any {
	if shape == nil || shape.shape == nil || shape.prepare == nil {
		panic("httpapi/response: invalid projected shape")
	}
	return shape.shape.projectShape(caller, shape.prepare(value))
}
