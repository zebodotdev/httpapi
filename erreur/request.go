package erreur

import (
	"fmt"
	"net/http"
	"strings"
)

// InvalidRequestBody returns a 400 error for request bodies that cannot be read
// or decoded as the expected JSON payload.
func InvalidRequestBody() *Error {
	code := "invalid_request_body"
	return &Error{
		Message: "invalid request body",
		FixCode: FixCodeChangeParams,
		Detail:  "send a valid json request body and try again.",
		Status:  http.StatusBadRequest,
		Cause:   CauseInvalidBody,
		Type:    TypeInvalidRequest,
		Code:    code,
		URL:     URL(code),
	}
}

// MethodNotAllowed returns a 405 error when a request uses a method the
// endpoint does not accept.
func MethodNotAllowed(allowed, got string) *Error {
	code := "method_not_allowed"
	allowed = strings.ToUpper(strings.TrimSpace(allowed))
	got = strings.ToUpper(strings.TrimSpace(got))

	return &Error{
		Message: "http method is not allowed",
		FixCode: FixCodeUseAllowedMethod,
		Detail: fmt.Sprintf(
			"this endpoint accepts %s requests. your request used %s.",
			allowed, got,
		),
		Status: http.StatusMethodNotAllowed,
		Cause:  CauseMethodNotAllowed,
		Type:   TypeInvalidRequest,
		Code:   code,
		URL:    URL(code),
	}
}

// UnsupportedContentType returns a 415 error when a request body uses an
// unsupported content type.
func UnsupportedContentType(got, want string) *Error {
	code := "unsupported_content_type"
	got = strings.TrimSpace(got)
	want = strings.TrimSpace(want)

	return &Error{
		Message: "unsupported content type",
		FixCode: FixCodeUseSupportedMediaType,
		Detail: fmt.Sprintf(
			"this endpoint accepts request bodies sent as %s. your request used %s.",
			want, got,
		),
		Status: http.StatusUnsupportedMediaType,
		Cause:  CauseUnsupportedContentType,
		Type:   TypeInvalidRequest,
		Code:   code,
		URL:    URL(code),
	}
}

// RequestTimeout returns a timeout error for endpoint handlers that exceed
// their configured runtime budget.
func RequestTimeout() *Error {
	code := "request_timeout"
	return &Error{
		Message: "request timed out",
		FixCode: FixCodeBackoff,
		Detail:  "the endpoint did not complete within its configured timeout. retry with backoff.",
		Status:  http.StatusGatewayTimeout,
		Cause:   CauseRequestTimeout,
		Type:    TypeTransient,
		Code:    code,
		URL:     URL(code),
	}
}
