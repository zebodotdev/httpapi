package erreur

const (
	// FixCodeAbandon indicates the request cannot succeed and should
	// be abandoned entirely. This typically signals a fundamental
	// constraint violation where no modification or retry strategy
	// will resolve the problem. Log the error for investigation.
	FixCodeAbandon = "abandon_request"

	// FixCodeChangeParams indicates the request parameters are
	// invalid and must be corrected before retrying. Examine the
	// error detail for specific parameter issues, update the request
	// with valid values, then resubmit.
	FixCodeChangeParams = "change_request_parameters"

	// FixCodeAuthenticate indicates the caller must provide valid
	// authentication credentials before retrying.
	FixCodeAuthenticate = "authenticate_request"

	// FixCodeUseAuthorizedCredentials indicates the caller should
	// retry with credentials that are allowed to perform the operation.
	FixCodeUseAuthorizedCredentials = "use_authorized_credentials"

	// FixCodeContactSupport indicates the error requires manual
	// intervention from support. This occurs for account-level
	// issues, configuration problems, or edge cases that cannot be
	// resolved through API calls alone. Contact support with error details.
	FixCodeContactSupport = "contact_customer_support"

	// FixCodeUseAllowedMethod indicates the caller must send the
	// request with one of the endpoint's accepted HTTP methods.
	FixCodeUseAllowedMethod = "use_allowed_method"

	// FixCodeUseSupportedMediaType indicates the caller must send the
	// request body using a content type accepted by the endpoint.
	FixCodeUseSupportedMediaType = "use_supported_media_type"

	// FixCodeReduceRequest indicates the caller must reduce the body,
	// page size, batch size, or other request dimensions before retrying.
	FixCodeReduceRequest = "reduce_request_size"

	// FixCodeRepeatSame indicates the request should be retried
	// without modification. This signals a transient failure where
	// the same request is expected to succeed on retry. Use exponential
	// backoff to avoid overwhelming the system during recovery.
	FixCodeRepeatSame = "repeat_same_request"

	// FixCodeBackoff indicates the caller may retry the same request
	// later, but should wait and use exponential backoff before doing so.
	FixCodeBackoff = "retry_with_backoff"

	// FixCodeRefreshState indicates the request was made against stale
	// resource state. Fetch the resource again, inspect the latest status,
	// and only issue a follow-up request if the desired state was not
	// already reached.
	FixCodeRefreshState = "refresh_resource_state"

	// FixCodeSatisfyPrecond indicates a required precondition must be
	// satisfied before the requested operation can proceed. Examine the
	// error detail to identify which precondition failed, complete the
	// necessary prerequisite action, then retry the operation.
	FixCodeSatisfyPrecond = "satisfy_precondition"

	// FixCodeUseNewIdempotencyKey indicates the caller must choose a
	// fresh idempotency key because the current key already belongs to
	// a different operation.
	FixCodeUseNewIdempotencyKey = "use_new_idempotency_key"

	// FixCodeResolveAccountConfiguration indicates the account or
	// application configuration must be updated before retrying.
	FixCodeResolveAccountConfiguration = "resolve_account_configuration"

	// FixCodeUseDifferentProviderMethod indicates the caller should
	// choose a different provider-backed method, instrument, or route.
	FixCodeUseDifferentProviderMethod = "use_different_provider_method"
)
