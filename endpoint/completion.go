package endpoint

import (
	"context"
	"fmt"
	"sync"
	"time"

	callerpkg "github.com/zebodotdev/httpapi/caller"
	e "github.com/zebodotdev/httpapi/erreur"
	responsepkg "github.com/zebodotdev/httpapi/response"
)

// CompletionOutcome identifies the runtime path that completed an endpoint
// request.
type CompletionOutcome string

const (
	// CompletionOutcomeHandled reports that the request reached the application
	// handler and completed without a more specific runtime outcome.
	CompletionOutcomeHandled CompletionOutcome = "handled"

	// CompletionOutcomeMethodNotAllowed reports that the request used an HTTP
	// method the endpoint does not accept.
	CompletionOutcomeMethodNotAllowed CompletionOutcome = "method_not_allowed"

	// CompletionOutcomeUnsupportedContentType reports that the request body was
	// sent with a content type the endpoint does not accept.
	CompletionOutcomeUnsupportedContentType CompletionOutcome = "unsupported_content_type"

	// CompletionOutcomeRequestTooLarge reports that the request exceeded the
	// endpoint's configured request-size limit.
	CompletionOutcomeRequestTooLarge CompletionOutcome = "request_too_large"

	// CompletionOutcomeInvalidRequestBody reports that httpapi could not read
	// the inbound request body.
	CompletionOutcomeInvalidRequestBody CompletionOutcome = "invalid_request_body"

	// CompletionOutcomeParseFailed reports that a typed endpoint parser rendered
	// a request-parameter error response.
	CompletionOutcomeParseFailed CompletionOutcome = "parse_failed"

	// CompletionOutcomeAccessDenied reports that endpoint access policy rejected
	// the request before the application handler ran.
	CompletionOutcomeAccessDenied CompletionOutcome = "access_denied"

	// CompletionOutcomeTimedOut reports that the endpoint handler exceeded its
	// configured runtime timeout before producing a response.
	CompletionOutcomeTimedOut CompletionOutcome = "timed_out"

	// CompletionOutcomePanicked reports that the handler panicked. httpapi still
	// records completion, then re-panics to preserve net/http panic behavior.
	CompletionOutcomePanicked CompletionOutcome = "panicked"
)

// Completion describes one finished endpoint request for logging, audit,
// metrics, notification, and other service-owned side effects.
//
// Completion is intentionally service-neutral. It carries httpapi request and
// endpoint metadata, but it does not define where events should be stored or
// how they should be routed.
type Completion struct {
	// Request is the completed httpapi request. It may include buffered request
	// body bytes, so sinks that persist it should apply their own redaction or
	// use Req's audit serialization.
	Request *Req

	// Endpoint is a safe snapshot of endpoint metadata captured at completion
	// time.
	Endpoint CompletionEndpoint

	// StartedAt is when the endpoint runtime began handling the request.
	StartedAt time.Time

	// CompletedAt is when the endpoint runtime finished handling the request.
	CompletedAt time.Time

	// Duration is the elapsed runtime duration.
	Duration time.Duration

	// Status is the response status selected by the endpoint runtime or handler.
	// It is zero when no response was selected.
	Status int

	// ResponseSizeBytes is the number of response bytes written when known.
	ResponseSizeBytes int

	// Outcome identifies the runtime path that completed the request.
	Outcome CompletionOutcome

	// Error is the structured httpapi error rendered for the response, when the
	// response used the standard error envelope.
	Error *e.Error

	// Panic is populated when the handler panicked.
	Panic *CompletionPanic
}

// CompletionEndpoint is a stable endpoint metadata snapshot included in
// completion events.
type CompletionEndpoint struct {
	Method               HttpMethod
	Pattern              string
	AcceptedContentTypes []ContentType
	Internal             bool
	Authorization        AuthorizationRequirement
	Callers              []callerpkg.Caller
	Idempotent           bool
	Route                RouteSpec
	Priority             EndpointPriority
	AuthKeys             map[string]bool
}

// CompletionPanic contains audit-safe panic metadata. Value is the recovered
// panic value for in-process observers; Type is suitable for logs and durable
// records.
type CompletionPanic struct {
	Value any
	Type  string
}

// CompletionSink observes endpoint completion events.
//
// Implementations should return quickly and avoid panicking. Sink errors are
// logged by the endpoint runtime and do not change the client response.
type CompletionSink interface {
	CompleteEndpoint(context.Context, Completion) error
}

// CompletionSinkFunc adapts a function to CompletionSink.
type CompletionSinkFunc func(context.Context, Completion) error

// CompleteEndpoint calls f(ctx, completion).
func (f CompletionSinkFunc) CompleteEndpoint(ctx context.Context, completion Completion) error {
	return f(ctx, completion)
}

type noopCompletionSink struct{}

// CompleteEndpoint ignores the completion event.
func (noopCompletionSink) CompleteEndpoint(context.Context, Completion) error {
	return nil
}

type endpointCompletionState struct {
	startedAt         time.Time
	responseSizeBytes int
	outcome           CompletionOutcome
	panicValue        any
}

var completionSinkState = struct {
	sync.RWMutex
	sink CompletionSink
}{
	sink: noopCompletionSink{},
}

// ConfigureCompletionSink installs the package-level endpoint completion sink.
// It returns a restore function for tests and short-lived overrides.
func ConfigureCompletionSink(sink CompletionSink) func() {
	if sink == nil {
		sink = noopCompletionSink{}
	}

	completionSinkState.Lock()
	prev := completionSinkState.sink
	completionSinkState.sink = sink
	completionSinkState.Unlock()

	return func() {
		completionSinkState.Lock()
		completionSinkState.sink = prev
		completionSinkState.Unlock()
	}
}

func currentCompletionSink() CompletionSink {
	completionSinkState.RLock()
	defer completionSinkState.RUnlock()
	return completionSinkState.sink
}

func (endpoint Endpoint) completeEndpoint(
	ctx context.Context,
	req *Req,
	completedAt time.Time,
	state endpointCompletionState,
) (err error) {
	defer func() {
		if panicValue := recover(); panicValue != nil {
			err = fmt.Errorf("completion sink panicked: %v", panicValue)
		}
	}()

	completion := Completion{
		Request:           req,
		Endpoint:          endpoint.completionEndpoint(),
		StartedAt:         state.startedAt,
		CompletedAt:       completedAt,
		Duration:          completedAt.Sub(state.startedAt),
		Status:            completionStatus(req),
		ResponseSizeBytes: state.responseSizeBytes,
		Outcome:           completionOutcome(req, state),
		Error:             completionError(req),
		Panic:             completionPanic(state.panicValue),
	}

	return currentCompletionSink().CompleteEndpoint(ctx, completion)
}

func (endpoint Endpoint) completionEndpoint() CompletionEndpoint {
	return CompletionEndpoint{
		Method:               endpoint.Method(),
		Pattern:              endpoint.Pattern(),
		AcceptedContentTypes: endpoint.AcceptedContentTypes(),
		Internal:             endpoint.IsInternal(),
		Authorization:        endpoint.Authorization(),
		Callers:              endpoint.AvailableCallers(),
		Idempotent:           endpoint.IsIdempotent(),
		Route:                endpoint.RouteSpec(),
		Priority:             endpoint.Priority(),
		AuthKeys:             endpoint.AuthKeys(),
	}
}

func completionStatus(req *Req) int {
	if req == nil || req.Res == nil {
		return 0
	}
	return req.Res.Status
}

func completionOutcome(req *Req, state endpointCompletionState) CompletionOutcome {
	if state.panicValue != nil {
		return CompletionOutcomePanicked
	}
	if state.outcome != "" {
		return state.outcome
	}
	if req != nil && req.AuthorizationFailure != nil {
		return CompletionOutcomeAccessDenied
	}
	if err := completionError(req); endpointCompletionParseError(err) {
		return CompletionOutcomeParseFailed
	}
	return CompletionOutcomeHandled
}

func completionError(req *Req) *e.Error {
	if req == nil {
		return nil
	}
	if req.Err != nil {
		return req.Err
	}
	if req.Res == nil {
		return nil
	}

	switch body := req.Res.Body.(type) {
	case responsepkg.ErrorBody:
		return body.Err
	case *responsepkg.ErrorBody:
		if body == nil {
			return nil
		}
		return body.Err
	default:
		return nil
	}
}

func endpointCompletionParseError(err *e.Error) bool {
	if err == nil {
		return false
	}
	return err.Cause == e.CauseInvalidParam ||
		err.Cause == e.CauseMissingParam ||
		err.Cause == e.CauseInvalidBody ||
		err.Type == e.TypeInvalidParam
}

func completionPanic(value any) *CompletionPanic {
	if value == nil {
		return nil
	}
	return &CompletionPanic{
		Value: value,
		Type:  fmt.Sprintf("%T", value),
	}
}
