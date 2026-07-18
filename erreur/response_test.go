package erreur

import (
	"net/http"
	"testing"
)

func TestResponseStatusDefaultsToInternalServerError(t *testing.T) {
	if got := ResponseStatus(nil); got != http.StatusInternalServerError {
		t.Fatalf("nil status = %d, want %d", got, http.StatusInternalServerError)
	}

	if got := ResponseStatus(&Error{}); got != http.StatusInternalServerError {
		t.Fatalf("empty error status = %d, want %d", got, http.StatusInternalServerError)
	}
}

func TestUnexpectedIsStableAndNonLeaky(t *testing.T) {
	got := Unexpected()

	if got.Code != "request_failed" {
		t.Fatalf("code = %q, want request_failed", got.Code)
	}
	if got.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", got.Status, http.StatusInternalServerError)
	}
	if got.Type != TypeUnknown {
		t.Fatalf("type = %q, want %q", got.Type, TypeUnknown)
	}
}

func TestNotFoundClassifiesMissingResource(t *testing.T) {
	got := NotFound("customer_not_found", "customer not found", "missing")

	if got.Code != "customer_not_found" {
		t.Fatalf("code = %q, want customer_not_found", got.Code)
	}
	if got.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", got.Status, http.StatusNotFound)
	}
	if got.FixCode != FixCodeChangeParams {
		t.Fatalf("fix_code = %q, want %q", got.FixCode, FixCodeChangeParams)
	}
}

func TestConstructorsClassifySharedHTTPFailures(t *testing.T) {
	tests := []struct {
		name    string
		err     *Error
		status  int
		typ     string
		cause   string
		fixCode string
		code    string
	}{
		{
			name:    "authn",
			err:     Unauthenticated("api_key_invalid", "invalid api key", "authenticate"),
			status:  http.StatusUnauthorized,
			typ:     TypeAuthentication,
			cause:   CauseAuthnFailed,
			fixCode: FixCodeAuthenticate,
			code:    "api_key_invalid",
		},
		{
			name:    "request envelope",
			err:     UnsupportedContentType("text/plain", "application/json"),
			status:  http.StatusUnsupportedMediaType,
			typ:     TypeInvalidRequest,
			cause:   CauseUnsupportedContentType,
			fixCode: FixCodeUseSupportedMediaType,
			code:    "unsupported_content_type",
		},
		{
			name:    "missing param",
			err:     MissingParam("order_id_missing", "order id is required", "provide order_id"),
			status:  http.StatusBadRequest,
			typ:     TypeInvalidParam,
			cause:   CauseMissingParam,
			fixCode: FixCodeChangeParams,
			code:    "order_id_missing",
		},
		{
			name:    "idempotency",
			err:     IdempotencyConflict(),
			status:  http.StatusUnprocessableEntity,
			typ:     TypeIdempotency,
			cause:   CauseIdempotencyConflict,
			fixCode: FixCodeUseNewIdempotencyKey,
			code:    "idempotency_key_conflict",
		},
		{
			name:    "limit",
			err:     RateLimited("request_rate_limited", "too many requests", "wait"),
			status:  http.StatusTooManyRequests,
			typ:     TypeLimit,
			cause:   CauseRateLimited,
			fixCode: FixCodeBackoff,
			code:    "request_rate_limited",
		},
		{
			name:    "provider",
			err:     ProviderDeclined("payment_method_declined", "payment method declined", "use another method"),
			status:  http.StatusBadRequest,
			typ:     TypeProvider,
			cause:   CauseProviderDeclined,
			fixCode: FixCodeUseDifferentProviderMethod,
			code:    "payment_method_declined",
		},
		{
			name:    "configuration",
			err:     ConfigurationMissing("payments_not_configured", "payments are not configured", "configure payments"),
			status:  http.StatusBadRequest,
			typ:     TypeConfiguration,
			cause:   CauseConfigurationMissing,
			fixCode: FixCodeResolveAccountConfiguration,
			code:    "payments_not_configured",
		},
		{
			name:    "operation",
			err:     OperationInProgress("order_payment_in_progress", "payment is already in progress", "wait"),
			status:  http.StatusConflict,
			typ:     TypeResourceLock,
			cause:   CauseOperationInProgress,
			fixCode: FixCodeBackoff,
			code:    "order_payment_in_progress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Status != tt.status {
				t.Fatalf("status = %d, want %d", tt.err.Status, tt.status)
			}
			if tt.err.Type != tt.typ {
				t.Fatalf("type = %q, want %q", tt.err.Type, tt.typ)
			}
			if tt.err.Cause != tt.cause {
				t.Fatalf("cause = %q, want %q", tt.err.Cause, tt.cause)
			}
			if tt.err.FixCode != tt.fixCode {
				t.Fatalf("fix_code = %q, want %q", tt.err.FixCode, tt.fixCode)
			}
			if tt.err.Code != tt.code {
				t.Fatalf("code = %q, want %q", tt.err.Code, tt.code)
			}
			if tt.err.URL != "" {
				t.Fatalf("url = %q, want empty without configured URL builder", tt.err.URL)
			}
		})
	}
}
