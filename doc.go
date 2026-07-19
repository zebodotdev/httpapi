// Package httpapi provides a reusable HTTP endpoint contract for services.
//
// The package owns endpoint definition, request wrapping, access metadata,
// runtime timeout budgets, idempotency orchestration, and response rendering.
// OpenAPI and gateway writers live in companion transcriber packages so the
// endpoint model can stay provider-neutral. Service-specific authentication,
// durable audit persistence, and idempotency storage are injected through
// interfaces.
package httpapi
