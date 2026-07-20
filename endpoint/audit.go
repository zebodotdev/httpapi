package endpoint

import (
	"context"
	"sync"
)

// AuditSink records completed request audits.
type AuditSink interface {
	RecordRequest(context.Context, *Req) error
}

// AuditSinkFunc adapts a function to AuditSink.
type AuditSinkFunc func(context.Context, *Req) error

func (f AuditSinkFunc) RecordRequest(ctx context.Context, req *Req) error {
	return f(ctx, req)
}

type noopAuditSink struct{}

func (noopAuditSink) RecordRequest(context.Context, *Req) error {
	return nil
}

var auditSinkState = struct {
	sync.RWMutex
	sink AuditSink
}{
	sink: noopAuditSink{},
}

// ConfigureAuditSink installs the package-level request audit sink.
// It returns a restore function for tests and short-lived overrides.
func ConfigureAuditSink(sink AuditSink) func() {
	if sink == nil {
		sink = noopAuditSink{}
	}

	auditSinkState.Lock()
	prev := auditSinkState.sink
	auditSinkState.sink = sink
	auditSinkState.Unlock()

	return func() {
		auditSinkState.Lock()
		auditSinkState.sink = prev
		auditSinkState.Unlock()
	}
}

func currentAuditSink() AuditSink {
	auditSinkState.RLock()
	defer auditSinkState.RUnlock()
	return auditSinkState.sink
}
