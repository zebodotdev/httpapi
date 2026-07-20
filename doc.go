// Package httpapi documents the module-level package layout.
//
// Runtime behavior intentionally lives in subpackages. Use endpoint for endpoint
// definitions and groups, request for safe request parsing and authentication
// attachment, response for rendering and response writing, auth for provider-
// neutral auth/session contracts, and openapi/* for spec transcription.
//
// The root package is intentionally doc-only so external consumers import the
// narrow package that owns the behavior they need.
package httpapi
