// Package erreur defines structured error response bodies and constructors.
package erreur

// Error is the structured error object returned by httpapi response helpers.
//
// Code is the most specific machine-readable identifier. Type, Cause, and
// FixCode group related errors so clients can choose display, retry, and
// remediation behavior without string-matching every Code.
type Error struct {
	// Message is a concise, human-readable description
	// of the error.  This provides a quick summary of
	// what went wrong without extensive detail.
	Message string `json:"message,omitzero"`

	// FixCode is a machine-readable suggestion for how
	// to resolve the error.  It translates the problem
	// into a specific action (e.g.,
	// "retry_with_exponential_backoff").
	FixCode string `json:"fix_code,omitzero"`

	// Detail is a comprehensive explanation of what went
	// wrong, including context about the operation and
	// actionable guidance for resolution.
	Detail string `json:"detail,omitzero"`

	// Status is the HTTP status code associated with this
	// error.  Not exported in JSON response - used internally
	// for setting response status.
	Status int `json:"-"`

	// Cause identifies the underlying category of
	// failure—which system or validation rule triggered the
	// error (e.g., "validation_failure", "payment_declined").
	Cause string `json:"cause,omitzero"`

	// Type categorizes the error into broad classes
	// (e.g., "invalid_request_parameter", "transient_error")
	// to help clients determine retry strategies.
	Type string `json:"type"`

	// Code is a unique machine-readable identifier for this
	// specific error condition.  Unlike Type and Cause which
	// group errors, Code pinpoints the exact problem.
	Code string `json:"code"`

	// URL is a link to the error reference documentation for
	// this specific error code.  Points to detailed explanation,
	// common causes, and resolution guidance.
	URL string `json:"url"`
}
