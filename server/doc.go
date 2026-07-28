// Package server provides httpapi's default HTTP server wiring.
//
// The package is intentionally small: applications still own service
// configuration, authentication setup, observability, CORS policy selection,
// and shutdown orchestration. The reusable server package turns an endpoint.Mux
// or custom http.Handler into an http.Server with conservative defaults,
// optional CORS middleware, and an explicit middleware chain.
//
// Config-level CORS is applied before the middleware chain so browser preflight
// requests do not enter auth-like middleware. Middleware is applied in
// declaration order. Given Middleware: []Middleware{A, B}, actual requests flow
// through A, then B, then the mux.
package server
