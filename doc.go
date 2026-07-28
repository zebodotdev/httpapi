// Package httpapi documents the module-level package layout for
// github.com/zebodotdev/httpapi.
//
// Runtime behavior intentionally lives in subpackages. Use endpoint for endpoint
// definitions and groups, request for safe request parsing and authentication
// attachment, caller for provider-neutral caller labels, param for reusable
// request-parameter definitions, response for rendering and response writing,
// auth for provider-neutral auth/session contracts, and openapi/* for spec
// transcription.
//
// The root package is intentionally doc-only so external consumers import the
// narrow package that owns the behavior they need.
//
// A typical service defines callers with caller.Define, configures request
// authentication and endpoint audit/idempotency adapters at startup, defines
// endpoint request parsers with param.JSON, defines caller-aware response shapes
// with response.Object or response.Project, then wires handlers through
// endpoint.DefineEndpoint.
package httpapi
