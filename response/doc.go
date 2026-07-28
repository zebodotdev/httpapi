// Package response owns response objects, render helpers, caller-aware response
// shapes, body encoding, and HTTP response writing for httpapi endpoint
// runtimes.
//
// Handlers can either construct a Res with JSON, Text, HTML, Bytes, Empty,
// NoContent, Redirect, or Stream and pass it to Render, or call convenience
// helpers such as RenderJSON, RenderErr, RenderNoContent, RenderRedirect, and
// RenderStream. Endpoint runtimes call WriteResponse after the handler returns
// so standard headers, timing, request IDs, streaming, raw bytes, and write
// deadlines are applied consistently. CORS belongs at the server middleware
// boundary, not in response writing.
//
// JSON object responses can also declare a response shape. A response shape is
// a projection contract: handlers still return normal Go values, while the
// shape explicitly lists which JSON attributes are emitted, their value types,
// and whether particular attributes are visible only to selected caller.Caller
// values. When a shaped body is rendered through a request target, response
// derives the caller from the target and omits attributes unavailable to that
// caller.
//
// The same caller model is shared by endpoints, request parameters, and response
// attributes. This keeps authorization-sensitive request and response contracts
// aligned without introducing response-specific service labels.
//
// The most direct response shape is Object:
//
//	type taskView struct {
//		ID           string
//		Status       string
//		InternalNote string
//	}
//
//	var taskResponse = response.Object[taskView](
//		response.Required("id", response.String(), func(task taskView) string { return task.ID }),
//		response.Required("status", response.String(), func(task taskView) string { return task.Status }),
//		response.Optional("internal_note", response.String(), func(task taskView) (string, bool) {
//			return task.InternalNote, task.InternalNote != ""
//		}).AvailableTo(Worker),
//	)
//
// Use Project when the handler has a domain value that should be prepared once
// before attributes are extracted. This keeps nil checks, defaults, and
// redaction in a small view adapter while the object shape stays explicit:
//
//	type taskView struct{ task Task }
//
//	func taskValue(task *Task) taskView {
//		if task == nil {
//			return taskView{}
//		}
//		return taskView{task: *task}
//	}
//
//	func (task taskView) ID() string     { return task.task.ID }
//	func (task taskView) Status() string { return task.task.Status }
//
//	var taskResponse = response.Project(
//		response.Object[taskView](
//			response.Required("id", response.String(), taskView.ID),
//			response.Required("status", response.String(), taskView.Status),
//		),
//		taskValue,
//	)
//
// Handlers should render shaped responses with shape.Body(value). RenderJSON and
// WriteResponse project the body for the caller attached to the request target.
package response
