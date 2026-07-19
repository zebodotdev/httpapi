// Package httpapi provides a reusable HTTP endpoint contract for services.
//
// The package owns runtime endpoint definition, request wrapping, idempotency
// orchestration, and response rendering. Provider-neutral auth contracts live in
// the auth subpackage, endpoint metadata contracts live in the endpoint
// subpackage, and OpenAPI/gateway writers live in companion transcriber
// packages. Service-specific authentication, durable audit persistence, and
// idempotency storage are injected through interfaces.
package httpapi
