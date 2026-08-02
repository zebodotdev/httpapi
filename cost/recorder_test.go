package cost

import (
	"context"
	"errors"
	"testing"
)

func TestRecorderContextPropagation(t *testing.T) {
	recorder := NewRecorder(Operation{
		ID:      "op_root",
		TraceID: "trace_root",
	})
	ctx := ContextWithRecorder(context.Background(), recorder)
	usage := NewUsageUnit(
		"aws",
		"dynamodb",
		"get-item",
		"read-request-unit",
		Whole(1),
	).WithLabel("table", "orders")

	if got := RecorderFromContext(ctx); got != recorder {
		t.Fatalf("RecorderFromContext returned %#v, want recorder", got)
	}
	if err := Record(ctx, usage); err != nil {
		t.Fatalf("Record returned error: %v", err)
	}

	usage.Labels["table"] = "mutated"
	event := recorder.Snapshot()
	if event.Operation.ID != "op_root" || event.Operation.RootID != "op_root" {
		t.Fatalf("operation = %#v, want root operation metadata", event.Operation)
	}
	if len(event.Usage) != 1 {
		t.Fatalf("usage units = %d, want 1", len(event.Usage))
	}
	if event.Usage[0].Labels["table"] != "orders" {
		t.Fatalf("usage label = %q, want orders", event.Usage[0].Labels["table"])
	}
}

func TestRecorderRecordsUsageAndBuildsEvent(t *testing.T) {
	recorder := NewRecorder(Operation{ID: "op_usage"})
	usage := NewUsageUnit(" aws ", " dynamodb ", " get-item ", " read-request-unit ", Whole(1)).
		WithLabel(" table ", " orders ")

	if err := recorder.Record(usage); err != nil {
		t.Fatalf("Record returned error: %v", err)
	}

	usage.Labels["table"] = "mutated"
	units := recorder.Usage()
	if len(units) != 1 {
		t.Fatalf("usage units = %d, want 1", len(units))
	}
	if units[0].Provider != "aws" || units[0].Service != "dynamodb" ||
		units[0].SKU != "get-item" || units[0].Unit != "read-request-unit" {
		t.Fatalf("usage unit was not normalized: %#v", units[0])
	}
	if units[0].Labels["table"] != "orders" {
		t.Fatalf("label = %q, want orders", units[0].Labels["table"])
	}
	if units[0].ObservedAt.IsZero() {
		t.Fatal("ObservedAt was not filled")
	}

	units[0].Labels["table"] = "changed"
	if got := recorder.Usage()[0].Labels["table"]; got != "orders" {
		t.Fatalf("stored label = %q, want orders", got)
	}

	event := recorder.Event(
		RequestMetadata{ID: "req_123", Status: 201},
		EndpointMetadata{OperationID: "create_order"},
	)
	if event.Request.ID != "req_123" || event.Request.Status != 201 {
		t.Fatalf("request metadata = %#v", event.Request)
	}
	if event.Endpoint.OperationID != "create_order" {
		t.Fatalf("endpoint metadata = %#v", event.Endpoint)
	}
	if len(event.Usage) != 1 {
		t.Fatalf("event usage units = %d, want 1", len(event.Usage))
	}
}

func TestStartChildLinksRecorderToParent(t *testing.T) {
	parent := NewRecorder(Operation{
		ID:                 "op_parent",
		RootID:             "op_root",
		TraceID:            "trace_root",
		CausationRequestID: "req_123",
	})
	parentCtx := ContextWithRecorder(context.Background(), parent)

	childCtx, child := StartChild(parentCtx, Operation{
		ID:   "op_child",
		Kind: "workflow",
	})

	if got := RecorderFromContext(parentCtx); got != parent {
		t.Fatalf("parent context recorder = %#v, want parent", got)
	}
	if got := RecorderFromContext(childCtx); got != child {
		t.Fatalf("child context recorder = %#v, want child", got)
	}

	operation := child.Operation()
	if operation.ID != "op_child" ||
		operation.ParentID != "op_parent" ||
		operation.RootID != "op_root" ||
		operation.TraceID != "trace_root" ||
		operation.CausationRequestID != "req_123" {
		t.Fatalf("child operation = %#v, want inherited linkage", operation)
	}
}

func TestRecorderRejectsMalformedUsage(t *testing.T) {
	recorder := NewRecorder(Operation{ID: "op_invalid_usage"})

	if err := recorder.Record(NewUsageUnit("", "dynamodb", "get-item", "request", Whole(1))); err == nil {
		t.Fatal("Record returned nil error for missing provider")
	}
	if err := recorder.Record(NewUsageUnit("aws", "dynamodb", "get-item", "request", Whole(0))); err == nil {
		t.Fatal("Record returned nil error for zero quantity")
	}

	usage := NewUsageUnit("aws", "dynamodb", "get-item", "request", Whole(1)).
		WithLabel("", "bad")
	if err := recorder.Record(usage); err == nil {
		t.Fatal("Record returned nil error for empty label name")
	}
}

func TestRecorderNilAndNoopBehavior(t *testing.T) {
	if err := Record(context.Background(), UsageUnit{}); err != nil {
		t.Fatalf("Record without recorder returned error: %v", err)
	}
	if err := Record(nil, UsageUnit{}); err != nil {
		t.Fatalf("Record with nil context returned error: %v", err)
	}
	if got := RecorderFromContext(ContextWithRecorder(context.Background(), nil)); got != nil {
		t.Fatalf("nil recorder was attached: %#v", got)
	}

	var recorder *Recorder
	if err := recorder.Record(UsageUnit{}); err != nil {
		t.Fatalf("nil recorder Record returned error: %v", err)
	}
	if usage := recorder.Usage(); usage != nil {
		t.Fatalf("nil recorder usage = %#v, want nil", usage)
	}
	if event := recorder.Snapshot(); !event.Empty() {
		t.Fatalf("nil recorder snapshot = %#v, want empty event", event)
	}
}

func TestRecorderFlushUsesSinkBoundary(t *testing.T) {
	recorder := NewRecorder(Operation{ID: "op_estimate"})
	if err := recorder.Record(NewUsageUnit(
		"gcp",
		"cloud-run",
		"cpu",
		"vcpu-second",
		MustDecimal(125, 2),
	).WithLabel("service", "api")); err != nil {
		t.Fatalf("Record returned error: %v", err)
	}

	var got OperationEvent
	var sinkLabel string
	sink := OperationSinkFunc(func(ctx context.Context, event OperationEvent) error {
		sinkLabel = event.Usage[0].Labels["service"]
		got = event
		event.Usage[0].Labels["service"] = "mutated"
		return nil
	})

	if err := recorder.Flush(context.Background(), sink); err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}
	if got.Operation.ID != "op_estimate" {
		t.Fatalf("sink operation = %#v, want op_estimate", got.Operation)
	}
	if got.Operation.CompletedAt.IsZero() {
		t.Fatal("sink event did not complete operation")
	}
	if len(got.Usage) != 1 || sinkLabel != "api" {
		t.Fatalf("sink usage = %#v, want cloned usage", got.Usage)
	}
	if recorder.Usage()[0].Labels["service"] != "api" {
		t.Fatal("sink mutated stored usage")
	}
}

func TestRecorderFlushReturnsSinkError(t *testing.T) {
	want := errors.New("sink unavailable")
	recorder := NewRecorder(Operation{ID: "op_estimate"})
	sink := OperationSinkFunc(func(context.Context, OperationEvent) error {
		return want
	})

	err := recorder.Flush(context.Background(), sink)
	if !errors.Is(err, want) {
		t.Fatalf("Flush error = %v, want %v", err, want)
	}
	if err := recorder.Flush(context.Background(), nil); err != nil {
		t.Fatalf("Flush with nil sink returned error: %v", err)
	}
}
