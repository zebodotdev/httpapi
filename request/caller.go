package request

import (
	"context"

	callerpkg "github.com/zebodotdev/httpapi/caller"
)

type callerContextKey struct{}

// Caller is the provider-neutral request caller label used by endpoint and
// parameter availability rules.
type Caller = callerpkg.Caller

// ContextWithCaller returns a child context carrying caller.
func ContextWithCaller(ctx context.Context, caller Caller) context.Context {
	if !caller.Defined() {
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, callerContextKey{}, caller)
}

// CallerFromContext returns a caller previously installed by ContextWithCaller.
func CallerFromContext(ctx context.Context) Caller {
	if ctx == nil {
		return Caller{}
	}
	caller, _ := ctx.Value(callerContextKey{}).(Caller)
	return caller
}

// AttachCaller attaches caller to the parsed request.
func (r *Req) AttachCaller(caller Caller) {
	if r == nil || !caller.Defined() {
		return
	}
	r.Caller = caller
}
