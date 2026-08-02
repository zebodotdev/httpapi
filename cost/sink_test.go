package cost

import (
	"context"
	"testing"
)

func TestEstimatorAndEstimateSinkFunctionAdapters(t *testing.T) {
	event := OperationEvent{
		Operation: Operation{ID: "op_123", RootID: "op_root"},
		Usage: []UsageUnit{
			NewUsageUnit(
				"aws",
				"dynamodb",
				"query",
				"read-request-unit",
				Whole(3),
			).WithLabel("table", "orders"),
		},
	}

	estimator := EstimatorFunc(func(ctx context.Context, event OperationEvent) (EstimateEvent, error) {
		event.Usage[0].Labels["table"] = "mutated"
		return NewEstimateEvent(event, "usd", MustDecimal(42, 2)).
			WithLabel("policy", "standard"), nil
	})

	estimate, err := estimator.EstimateOperation(context.Background(), event)
	if err != nil {
		t.Fatalf("EstimateOperation returned error: %v", err)
	}
	if event.Usage[0].Labels["table"] != "orders" {
		t.Fatal("estimator mutated source operation event")
	}
	if estimate.Currency != "USD" || estimate.Amount != MustDecimal(42, 2) {
		t.Fatalf("estimate amount = %s %s, want USD 0.42", estimate.Currency, estimate.Amount)
	}
	if estimate.Labels["policy"] != "standard" {
		t.Fatalf("estimate labels = %#v, want policy label", estimate.Labels)
	}

	var got EstimateEvent
	sink := EstimateSinkFunc(func(ctx context.Context, estimate EstimateEvent) error {
		got = estimate
		estimate.Labels["policy"] = "mutated"
		return nil
	})

	if err := sink.RecordEstimate(context.Background(), estimate); err != nil {
		t.Fatalf("RecordEstimate returned error: %v", err)
	}
	if got.Operation.ID != "op_123" {
		t.Fatalf("estimate operation = %#v, want op_123", got.Operation)
	}
	if estimate.Labels["policy"] != "standard" {
		t.Fatal("estimate sink mutated source estimate")
	}
}

func TestNoopSinks(t *testing.T) {
	if err := NoopOperationSink().RecordOperation(context.Background(), OperationEvent{}); err != nil {
		t.Fatalf("NoopOperationSink returned error: %v", err)
	}
	if err := NoopEstimateSink().RecordEstimate(context.Background(), EstimateEvent{}); err != nil {
		t.Fatalf("NoopEstimateSink returned error: %v", err)
	}
}
