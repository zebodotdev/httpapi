package erreur

const (
	// CauseUnknown indicates the error cause could not be
	// determined.  This occurs for unexpected failures that
	// don't map to known categories. Log these errors for
	// investigation as they may represent edge cases requiring
	// code changes.
	CauseUnknown = "unknown"

	// CauseInvalidParam indicates one or more request
	// parameters contain invalid values that violate format,
	// type, or business logic constraints. Check the error
	// detail for specific parameter issues and correct the
	// values before retrying.
	CauseInvalidParam = "invalid_request_parameter"

	// CauseMissingParam indicates required request parameters
	// were not provided. Review the API documentation for the
	// endpoint to identify which parameters are mandatory, then
	// include them in your request and retry.
	CauseMissingParam = "missing_request_parameter"

	// CauseAuthnFailed indicates authentication failed due to
	// missing, invalid, or expired credentials. Verify your API
	// key is correct, properly formatted in the Authorization
	// header, and hasn't been revoked or expired.
	CauseAuthnFailed = "authentication_failed"

	// CauseAuthzFailed indicates your credentials were
	// authenticated successfully but lack permission to perform
	// the requested operation. This typically means the resource
	// belongs to a different application or your API key's
	// permissions don't include the attempted action.
	CauseAuthzFailed = "authorization_failed"

	// CauseNotFound indicates the requested resource does not exist or is not
	// visible to the caller.
	CauseNotFound = "resource_not_found"

	// CauseInvalidBody indicates the request body could not be read
	// or decoded as the expected JSON payload.
	CauseInvalidBody = "invalid_request_body"

	// CauseUnsupportedContentType indicates the request used a
	// content type this endpoint does not accept.
	CauseUnsupportedContentType = "unsupported_content_type"

	// CauseMethodNotAllowed indicates the endpoint exists but the
	// HTTP method used for this request is not accepted.
	CauseMethodNotAllowed = "method_not_allowed"

	// CauseRequestTimeout indicates endpoint processing exceeded its configured
	// runtime budget before a response could be produced.
	CauseRequestTimeout = "request_timeout"

	// CausePayloadTooLarge indicates the request payload exceeds the
	// size limit accepted by the API or endpoint.
	CausePayloadTooLarge = "payload_too_large"

	// CausePreconditionUnmet indicates the caller must satisfy a prerequisite
	// before the operation can run.
	CausePreconditionUnmet = "precondition_unmet"

	// CauseStateConflict indicates the resource changed or is already in a
	// conflicting state.
	CauseStateConflict = "resource_state_conflict"

	// CauseStateInvalid indicates the resource's current state cannot support
	// the attempted operation.
	CauseStateInvalid = "resource_state_invalid"

	// CauseIdempotencyConflict indicates an idempotency key was
	// reused with a different operation payload.
	CauseIdempotencyConflict = "idempotency_conflict"

	// CauseIdempotencyInProgress indicates an idempotency key is
	// already reserved by an operation that has not completed yet.
	CauseIdempotencyInProgress = "idempotency_in_progress"

	// CauseRateLimited indicates the client is sending requests too
	// quickly for the current rate window.
	CauseRateLimited = "rate_limited"

	// CauseQuotaExceeded indicates an account or application quota
	// has been exhausted.
	CauseQuotaExceeded = "quota_exceeded"

	// CauseLimitExceeded indicates a non-rate request or resource
	// limit was exceeded.
	CauseLimitExceeded = "limit_exceeded"

	// CauseServiceUnavailable indicates the service is temporarily unable to
	// process the request.
	CauseServiceUnavailable = "service_unavailable"

	// CauseDependencyUnavailable indicates an internal dependency is
	// temporarily unavailable.
	CauseDependencyUnavailable = "dependency_unavailable"

	// CauseDependencyFailed indicates an internal dependency returned a terminal
	// or unexpected failure.
	CauseDependencyFailed = "dependency_failed"

	// CauseProviderUnavailable indicates an external provider or
	// processor is temporarily unavailable.
	CauseProviderUnavailable = "provider_unavailable"

	// CauseProviderDeclined indicates an external provider actively
	// rejected the attempted operation.
	CauseProviderDeclined = "provider_declined"

	// CauseConfigurationMissing indicates required account or
	// application configuration is missing.
	CauseConfigurationMissing = "configuration_missing"

	// CauseCapabilityUnsupported indicates the account, application,
	// or resource does not support the requested capability.
	CauseCapabilityUnsupported = "capability_unsupported"

	// CauseResourceLocked indicates the resource is locked by another
	// write or workflow.
	CauseResourceLocked = "resource_locked"

	// CauseOperationInProgress indicates the requested operation is
	// already running and should not be started again.
	CauseOperationInProgress = "operation_in_progress"
)
