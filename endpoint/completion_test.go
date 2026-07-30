package endpoint

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zebodotdev/httpapi/param"
	"github.com/zebodotdev/httpapi/response"
)

func TestEndpointCompletionSinkReceivesSuccessfulRequest(t *testing.T) {
	events := captureEndpointCompletions(t)
	endpoint := DefineEndpoint(EndpointSpec{
		Method: POST,
		Path:   "/tasks/create",
		Handler: func(r *Req) {
			response.RenderJSON(r, http.StatusCreated, map[string]bool{"ok": true})
		},
		Route: RouteSpec{
			OperationID: "createTask",
			Summary:     "Create task",
		},
	})
	req := httptest.NewRequest(POST, "/tasks/create", strings.NewReader(`{}`))
	req.Header.Set(contentTypeHeaderKey, ApplicationJson)
	res := httptest.NewRecorder()

	endpoint.Handler()(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", res.Code, http.StatusCreated, res.Body.String())
	}
	event := onlyEndpointCompletion(t, events)
	if event.Outcome != CompletionOutcomeHandled {
		t.Fatalf("outcome = %q, want %q", event.Outcome, CompletionOutcomeHandled)
	}
	if event.Status != http.StatusCreated {
		t.Fatalf("completion status = %d, want %d", event.Status, http.StatusCreated)
	}
	if event.Endpoint.Method != POST || event.Endpoint.Pattern != "/tasks/create" {
		t.Fatalf("endpoint = %s %s", event.Endpoint.Method, event.Endpoint.Pattern)
	}
	if event.Endpoint.Route.OperationID != "createTask" {
		t.Fatalf("operation id = %q, want createTask", event.Endpoint.Route.OperationID)
	}
	if event.Request == nil || event.Request.ID == "" {
		t.Fatalf("request = %#v, want request id", event.Request)
	}
	if event.Duration <= 0 {
		t.Fatalf("duration = %v, want positive duration", event.Duration)
	}
	if event.ResponseSizeBytes == 0 {
		t.Fatal("response size was not recorded")
	}
}

func TestEndpointCompletionSinkReceivesTypedParseFailure(t *testing.T) {
	events := captureEndpointCompletions(t)
	called := false
	endpoint := DefineJSONEndpoint(JSONEndpointSpec[typedJSONParams]{
		Method:  POST,
		Path:    "/typed",
		Request: typedJSONRequest(),
		Handler: func(r *Req, params typedJSONParams) {
			called = true
			response.RenderJSON(r, http.StatusOK, params)
		},
	})
	req := httptest.NewRequest(POST, "/typed", strings.NewReader(`{}`))
	req.Header.Set(contentTypeHeaderKey, ApplicationJson)
	res := httptest.NewRecorder()

	endpoint.Handler()(res, req)

	if called {
		t.Fatal("typed handler was called")
	}
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", res.Code, http.StatusBadRequest, res.Body.String())
	}
	event := onlyEndpointCompletion(t, events)
	if event.Outcome != CompletionOutcomeParseFailed {
		t.Fatalf("outcome = %q, want %q", event.Outcome, CompletionOutcomeParseFailed)
	}
	if event.Error == nil || event.Error.Code != string(param.CodeMissing) {
		t.Fatalf("completion error = %#v, want missing parameter error", event.Error)
	}
}

func TestEndpointCompletionSinkReceivesAccessFailure(t *testing.T) {
	events := captureEndpointCompletions(t)
	called := false
	endpoint := DefineEndpoint(EndpointSpec{
		Method: POST,
		Path:   "/private",
		Access: EndpointAccessSpec{
			Authorization: RequiredAuthorization(AuthorizationKindBearer),
		},
		Handler: func(r *Req) {
			called = true
			response.RenderNoContent(r)
		},
	})
	req := httptest.NewRequest(POST, "/private", strings.NewReader(`{}`))
	req.Header.Set(contentTypeHeaderKey, ApplicationJson)
	res := httptest.NewRecorder()

	endpoint.Handler()(res, req)

	if called {
		t.Fatal("handler was called")
	}
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", res.Code, http.StatusUnauthorized, res.Body.String())
	}
	event := onlyEndpointCompletion(t, events)
	if event.Outcome != CompletionOutcomeAccessDenied {
		t.Fatalf("outcome = %q, want %q", event.Outcome, CompletionOutcomeAccessDenied)
	}
	if event.Request == nil || event.Request.AuthorizationFailure == nil {
		t.Fatalf("authorization failure = %#v", event.Request)
	}
	if event.Error == nil || event.Error.Code != endpointAuthenticationRequiredCode {
		t.Fatalf("completion error = %#v, want endpoint auth error", event.Error)
	}
}

func TestEndpointCompletionSinkReceivesPanicBeforePropagation(t *testing.T) {
	events := captureEndpointCompletions(t)
	endpoint := DefineEndpoint(EndpointSpec{
		Method: POST,
		Path:   "/panic",
		Handler: func(r *Req) {
			panic("boom")
		},
	})
	req := httptest.NewRequest(POST, "/panic", strings.NewReader(`{}`))
	req.Header.Set(contentTypeHeaderKey, ApplicationJson)
	res := httptest.NewRecorder()

	func() {
		defer func() {
			got := recover()
			if got != "boom" {
				t.Fatalf("panic = %#v, want boom", got)
			}
		}()
		endpoint.Handler()(res, req)
	}()

	event := onlyEndpointCompletion(t, events)
	if event.Outcome != CompletionOutcomePanicked {
		t.Fatalf("outcome = %q, want %q", event.Outcome, CompletionOutcomePanicked)
	}
	if event.Status != http.StatusInternalServerError {
		t.Fatalf("completion status = %d, want %d", event.Status, http.StatusInternalServerError)
	}
	if event.Panic == nil || event.Panic.Type != "string" || event.Panic.Value != "boom" {
		t.Fatalf("panic metadata = %#v", event.Panic)
	}
	if event.Error == nil {
		t.Fatal("completion error was not recorded for panic response")
	}
}

func captureEndpointCompletions(t *testing.T) *[]Completion {
	t.Helper()
	events := []Completion{}
	restore := ConfigureCompletionSink(CompletionSinkFunc(
		func(ctx context.Context, completion Completion) error {
			events = append(events, completion)
			return nil
		},
	))
	t.Cleanup(restore)
	return &events
}

func onlyEndpointCompletion(t *testing.T, events *[]Completion) Completion {
	t.Helper()
	if events == nil {
		t.Fatal("events pointer is nil")
	}
	if len(*events) != 1 {
		t.Fatalf("completion events = %d, want 1", len(*events))
	}
	return (*events)[0]
}
