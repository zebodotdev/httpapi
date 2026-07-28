// Package endpoint contains provider-neutral endpoint metadata contracts used by
// httpapi runtime endpoints and spec transcribers.
//
// DefineEndpoint is the preferred entry point for new endpoints. It collects
// route, access, caller availability, idempotency, priority, timeout,
// content-type, and handler configuration into one EndpointSpec so request
// parsing can remain independent from endpoint expectations.
//
// Endpoint requirements belong on EndpointSpec. A request.Req only represents a
// safely parsed incoming HTTP request with caller and session state attached; it
// should not know which route will eventually handle it. The endpoint runtime
// combines the Req with the EndpointSpec and enforces method, content type,
// authorization, caller availability, request size limits, idempotency, and
// timeout policy before or around the handler.
//
// Endpoint metadata is intentionally provider-neutral. RouteSpec describes
// operation IDs, summaries, backend addresses, path forwarding behavior, and
// backend timeout intent without embedding a cloud provider's document format.
// Target transcribers under openapi/* translate that metadata into concrete
// OpenAPI or gateway documents.
//
// EndpointGroup applies shared defaults to several endpoints. Group caller
// availability narrows endpoint availability; it cannot widen an endpoint that
// was already restricted. Empty caller availability means available to all.
//
// NewEndpoint, NewIdempotentEndpoint, and
// NewIdempotentEndpointWithScopeResolver exist for compatibility with older
// integrations. Prefer DefineEndpoint for new code because a named EndpointSpec
// remains readable as endpoint requirements grow.
package endpoint
