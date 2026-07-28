package response

import callerpkg "github.com/zebodotdev/httpapi/caller"

// CallerTarget is a response target that can expose the caller attached to the
// request being rendered.
//
// request.Req satisfies this interface. Render uses it to project shaped bodies
// without asking endpoint handlers to pass caller information manually.
type CallerTarget interface {
	Target

	// RequestCaller returns the application-defined request caller.
	RequestCaller() callerpkg.Caller
}

// CallerProjector is a response body that can project itself for a caller.
//
// ShapedBody implements CallerProjector. Custom body types may implement it
// when they need caller-aware projection but do not fit ObjectShape.
type CallerProjector interface {
	// ProjectResponse returns the body visible to caller.
	ProjectResponse(caller callerpkg.Caller) any
}

func responseForTarget(target Target, res *Res) *Res {
	if res == nil {
		return nil
	}

	body, ok := projectResponseBody(res.Body, callerFromTarget(target))
	if !ok {
		return res
	}

	projected := *res
	projected.Header = cloneHeader(res.Header)
	projected.Body = body
	return &projected
}

func callerFromTarget(target Target) callerpkg.Caller {
	callerTarget, ok := target.(CallerTarget)
	if !ok {
		return callerpkg.Caller{}
	}
	return callerTarget.RequestCaller()
}

func projectResponseBody(body any, caller callerpkg.Caller) (any, bool) {
	projector, ok := body.(CallerProjector)
	if !ok {
		return body, false
	}
	return projector.ProjectResponse(caller), true
}
