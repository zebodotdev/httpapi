package cost

import (
	"context"
	"sync"
	"time"
)

type recorderContextKey struct{}

// Recorder collects usage observations for one request, job, workflow, or
// other operation.
//
// Recorder is safe for concurrent use. Application code can attach a recorder
// to context once and let repository, provider, workflow, activity, or test
// helpers report usage with Record(ctx, usage).
type Recorder struct {
	mu        sync.Mutex
	operation Operation
	units     []UsageUnit
}

// NewRecorder returns a recorder for operation.
//
// When operation.ID is empty, httpapi generates one. When RootID is empty, it
// defaults to ID for a root operation. Service-owned background work should
// prefer durable ids such as job ids, workflow ids, activity ids, or request ids
// when those are available.
func NewRecorder(operation Operation) *Recorder {
	return &Recorder{operation: normalizeOperation(operation)}
}

// NewChildRecorder returns a recorder linked to parent.
//
// Empty ParentID, RootID, TraceID, and CausationRequestID values are inherited
// from parent. If parent is nil, NewChildRecorder behaves like NewRecorder.
func NewChildRecorder(parent *Recorder, operation Operation) *Recorder {
	if parent == nil {
		return NewRecorder(operation)
	}
	return &Recorder{
		operation: childOperation(parent.Operation(), operation),
	}
}

// ContextWithRecorder attaches recorder to ctx.
func ContextWithRecorder(ctx context.Context, recorder *Recorder) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if recorder == nil {
		return ctx
	}
	return context.WithValue(ctx, recorderContextKey{}, recorder)
}

// RecorderFromContext returns the recorder attached to ctx, if any.
func RecorderFromContext(ctx context.Context) *Recorder {
	if ctx == nil {
		return nil
	}
	recorder, _ := ctx.Value(recorderContextKey{}).(*Recorder)
	return recorder
}

// StartChild creates a child recorder from the recorder already attached to ctx
// and returns a context containing the child recorder.
//
// When ctx has no recorder, StartChild starts a new root recorder. This keeps
// background jobs and tests easy to wire while preserving parent linkage when a
// request or workflow recorder is already present.
func StartChild(ctx context.Context, operation Operation) (context.Context, *Recorder) {
	parent := RecorderFromContext(ctx)
	recorder := NewChildRecorder(parent, operation)
	return ContextWithRecorder(ctx, recorder), recorder
}

// Record records usage on the recorder attached to ctx.
//
// Record is a no-op when ctx has no recorder. That lets lower-level helpers
// report usage without forcing every caller or test to install cost accounting.
func Record(ctx context.Context, usage UsageUnit) error {
	recorder := RecorderFromContext(ctx)
	if recorder == nil {
		return nil
	}
	return recorder.Record(usage)
}

// RecordUnit records one usage unit on the recorder attached to ctx.
func RecordUnit(
	ctx context.Context,
	provider Provider,
	service Service,
	sku SKU,
	unit Unit,
	quantity Quantity,
) error {
	return Record(ctx, NewUsageUnit(provider, service, sku, unit, quantity))
}

// Operation returns this recorder's operation metadata.
func (recorder *Recorder) Operation() Operation {
	if recorder == nil {
		return Operation{}
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return recorder.operation
}

// Complete marks the operation completed at completedAt.
//
// Passing a zero time uses time.Now. Complete does not write to a sink; call
// Flush or pass Snapshot/Event to a service-owned sink when a durable estimate
// should be persisted.
func (recorder *Recorder) Complete(completedAt time.Time) {
	if recorder == nil {
		return
	}
	if completedAt.IsZero() {
		completedAt = time.Now()
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	recorder.operation.CompletedAt = completedAt
}

// Record records one usage unit.
//
// Record validates and normalizes the unit before storing it. If
// usage.ObservedAt is empty, Record fills it with the current time.
func (recorder *Recorder) Record(usage UsageUnit) error {
	if recorder == nil {
		return nil
	}

	usage, err := NormalizeUsageUnit(usage)
	if err != nil {
		return err
	}
	if usage.ObservedAt.IsZero() {
		usage.ObservedAt = time.Now()
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	recorder.units = append(recorder.units, usage)
	return nil
}

// AddUnit records one usage unit from its required fields.
func (recorder *Recorder) AddUnit(
	provider Provider,
	service Service,
	sku SKU,
	unit Unit,
	quantity Quantity,
) error {
	return recorder.Record(NewUsageUnit(provider, service, sku, unit, quantity))
}

// Usage returns a copy of every usage unit recorded so far.
func (recorder *Recorder) Usage() []UsageUnit {
	if recorder == nil {
		return nil
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return cloneUsageUnits(recorder.units)
}

// Empty reports whether recorder has recorded no usage units.
func (recorder *Recorder) Empty() bool {
	if recorder == nil {
		return true
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return len(recorder.units) == 0
}

// Snapshot returns an operation cost event containing operation metadata and a
// snapshot of the currently recorded usage units.
func (recorder *Recorder) Snapshot() OperationEvent {
	if recorder == nil {
		return OperationEvent{}
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return OperationEvent{
		Operation:  recorder.operation,
		Usage:      cloneUsageUnits(recorder.units),
		RecordedAt: time.Now(),
	}
}

// Event returns an operation cost event containing operation, request,
// endpoint, and usage metadata.
func (recorder *Recorder) Event(
	request RequestMetadata,
	endpoint EndpointMetadata,
) OperationEvent {
	event := recorder.Snapshot()
	event.Request = request
	event.Endpoint = endpoint
	return event
}

// Flush records this recorder's current operation event in sink.
//
// Flush is the generic durable estimate boundary for jobs, workflows,
// activities, and tests that do not pass through endpoint completion. The sink
// receives a cloned OperationEvent and owns estimate production, pricing,
// storage, invoice matching, reconciliation, and retry policy. Passing a nil
// sink or using a nil recorder is a no-op.
func (recorder *Recorder) Flush(ctx context.Context, sink OperationSink) error {
	if recorder == nil || sink == nil {
		return nil
	}

	recorder.Complete(time.Now())
	return sink.RecordOperation(ctx, recorder.Snapshot())
}

// Flush records the context recorder's current operation event in sink.
func Flush(ctx context.Context, sink OperationSink) error {
	recorder := RecorderFromContext(ctx)
	if recorder == nil {
		return nil
	}
	return recorder.Flush(ctx, sink)
}
