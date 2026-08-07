package request

import (
	"context"
	"sync"
)

// ContextAnnotator lets services attach service-specific values to the request
// context after httpapi has parsed authentication and request metadata. The hook
// is intentionally generic so the shared httpapi package does not depend on any
// service package.
type ContextAnnotator interface {
	AnnotateRequestContext(context.Context, *Req) context.Context
}

// ContextAnnotatorFunc adapts a function to ContextAnnotator.
type ContextAnnotatorFunc func(context.Context, *Req) context.Context

func (f ContextAnnotatorFunc) AnnotateRequestContext(ctx context.Context, req *Req) context.Context {
	if f == nil {
		return ctx
	}
	return f(ctx, req)
}

type noopContextAnnotator struct{}

func (noopContextAnnotator) AnnotateRequestContext(ctx context.Context, _ *Req) context.Context {
	return ctx
}

var contextAnnotatorState = struct {
	sync.RWMutex
	annotator ContextAnnotator
}{
	annotator: noopContextAnnotator{},
}

// ConfigureContextAnnotator installs the package-level request context
// annotator. It returns a restore function for tests and short-lived overrides.
// Passing nil restores the no-op annotator.
func ConfigureContextAnnotator(annotator ContextAnnotator) func() {
	if annotator == nil {
		annotator = noopContextAnnotator{}
	}

	contextAnnotatorState.Lock()
	prev := contextAnnotatorState.annotator
	contextAnnotatorState.annotator = annotator
	contextAnnotatorState.Unlock()

	return func() {
		contextAnnotatorState.Lock()
		contextAnnotatorState.annotator = prev
		contextAnnotatorState.Unlock()
	}
}

func currentContextAnnotator() ContextAnnotator {
	contextAnnotatorState.RLock()
	defer contextAnnotatorState.RUnlock()
	return contextAnnotatorState.annotator
}

func (r *Req) annotateContext() {
	if r == nil || r.Req == nil {
		return
	}
	ctx := currentContextAnnotator().AnnotateRequestContext(r.Req.Context(), r)
	if ctx == nil {
		return
	}
	r.Req = r.Req.WithContext(ctx)
}
