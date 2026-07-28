package request

import (
	"context"

	callerpkg "github.com/zebodotdev/httpapi/caller"
)

type callerContextKey struct{}

// Caller is the provider-neutral request caller label used by endpoint,
// parameter, and response-attribute availability rules.
type Caller = callerpkg.Caller

// ContextWithCaller returns a child context carrying caller.
//
// Trusted middleware should attach the caller before endpoint handling begins.
// NewReq reads the value back from the request context and stores it on Req so
// endpoint access checks, parameter parsing, and response projection all use the
// same caller value.
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

// RequestCaller returns the application-defined caller attached to the request.
//
// RequestCaller is the small interface shared with param.WithRequestCaller and
// response rendering. Prefer it over reading Req.Caller directly when another
// package only needs caller identity.
func (r *Req) RequestCaller() Caller {
	if r == nil {
		return Caller{}
	}
	return r.Caller
}
