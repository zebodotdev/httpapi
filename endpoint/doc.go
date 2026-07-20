// Package endpoint contains provider-neutral endpoint metadata contracts used by
// httpapi runtime endpoints and spec transcribers.
//
// DefineEndpoint is the preferred entry point for new endpoints. It collects
// route, access, idempotency, priority, timeout, content-type, and handler
// configuration into one EndpointSpec so request parsing can remain independent
// from endpoint expectations.
package endpoint
