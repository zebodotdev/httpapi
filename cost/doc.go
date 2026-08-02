// Package cost defines provider-neutral usage observations for httpapi
// requests, background jobs, workflows, activities, and other operations.
//
// The package intentionally records usage units and operation correlation, not
// prices. A service can observe that an operation consumed one DynamoDB read
// request unit, several Cloud Run CPU milliseconds, an internal referral
// request, or an outbound SMS segment without embedding billing-account
// knowledge, USD conversion, or invoice reconciliation in httpapi.
//
// Request handlers can attach usage through request.Req methods or by calling
// Record with the request context. Endpoint completion observers receive an
// OperationEvent with operation, request, endpoint, and usage metadata.
//
// Background jobs and workflows can create a Recorder directly, attach it to a
// context, start child recorders for async work, and Flush to a service-owned
// OperationSink when a durable estimate should be persisted. Sinks own pricing
// policy, storage, calibration, invoice matching, reconciliation, and retries.
package cost
