package erreur

const (
	TypeBadRequest = "bad_request"

	// TypeAuthentication indicates the request did not provide
	// usable credentials. Clients must authenticate before the
	// request can be processed.
	TypeAuthentication = "authentication_error"

	// TypeAuthorization indicates the request credentials are valid
	// but not permitted to perform the attempted operation.
	TypeAuthorization = "authorization_error"

	// TypeInvalidParam indicates the request failed due to invalid
	// or incorrectly formatted parameters. This error type signals
	// that the client must modify the request before retrying—the
	// same request will always fail without changes to input data.
	TypeInvalidParam = "invalid_request_parameter"

	// TypeInvalidRequest indicates the HTTP envelope itself is
	// invalid: method, content type, body shape, or similar request
	// framing failed before endpoint-specific validation could run.
	TypeInvalidRequest = "invalid_request"

	// TypeIdempotency indicates the request failed because its
	// idempotency key or replay state cannot be accepted for this
	// operation.
	TypeIdempotency = "idempotency_error"

	// TypeLimit indicates the request exceeds a platform, account,
	// or endpoint limit and must be reduced or delayed.
	TypeLimit = "limit_error"

	// TypeProvider indicates an upstream processor or provider
	// rejected or failed the operation in a way clients may need to
	// surface distinctly from local validation.
	TypeProvider = "provider_error"

	// TypeConfiguration indicates the account, application, or
	// requested capability is not configured to perform the operation.
	TypeConfiguration = "configuration_error"

	// TypeResourceLock indicates the target resource or operation is
	// temporarily locked by another write or in-flight workflow.
	TypeResourceLock = "resource_lock"

	// TypeTransient indicates a temporary failure that may succeed
	// if retried. These errors occur due to momentary system load,
	// network issues, or resource contention that is expected to
	// resolve quickly. Clients should retry with exponential backoff.
	TypeTransient = "transient_error"

	// TypeStateConflict indicates the target resource changed state
	// before or during request processing. Clients should refresh the
	// resource before deciding whether another request is necessary.
	TypeStateConflict = "state_conflict"

	// TypeUnknown indicates an unexpected failure that doesn't fit
	// established error categories. These errors are typically logged
	// for investigation and may represent edge cases or system issues
	// that require engineering attention to properly classify.
	TypeUnknown = "unknown_error"
)
