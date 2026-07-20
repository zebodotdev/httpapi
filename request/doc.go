// Package request owns safe request parsing, request audit serialization, and
// request-scoped authentication state for httpapi endpoint runtimes.
//
// Req is deliberately not an endpoint contract. It buffers and restores the
// incoming body, exposes normalized accessors, attaches sessions, and produces
// redacted audit JSON while endpoint.Endpoint decides what the request must
// satisfy.
package request
