// Package httpapi provides a reusable HTTP endpoint contract for services.
//
// The package owns endpoint definition, request wrapping, access metadata,
// runtime timeout budgets, idempotency orchestration, response rendering, and
// OpenAPI/gateway transcription. Service-specific authentication, durable audit
// persistence, and idempotency storage are injected through interfaces.
package httpapi
