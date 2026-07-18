package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// EndpointTimeoutSpec declares runtime timeout budgets for one endpoint.
//
// These budgets are enforced by the in-process HTTP wrapper. They are separate
// from RouteBackend.Timeout, which only describes generated gateway/backend
// deadlines for transcribed route specs.
type EndpointTimeoutSpec struct {
	ReadBody time.Duration
	Handler  time.Duration
	Write    time.Duration
}

type endpointTimeoutPolicy struct {
	timeout EndpointTimeoutSpec

	readBodyInherited bool
	handlerInherited  bool
	writeInherited    bool
}

type endpointDeadlineSetter func(*http.ResponseController, time.Time) error

// WithTimeoutSpec applies runtime timeout budgets to an endpoint.
func WithTimeoutSpec(spec EndpointTimeoutSpec) EndpointOption {
	spec = normalizeEndpointTimeoutSpec(spec)
	return func(e *Endpoint) {
		e.timeout = endpointTimeoutPolicy{timeout: spec}
	}
}

// ConfigureTimeoutSpec sets default runtime timeout budgets for endpoints in
// the group. Endpoint-level timeout fields override group defaults field by
// field.
func (eg *EndpointGroup) ConfigureTimeoutSpec(spec EndpointTimeoutSpec) {
	eg.Timeout = normalizeEndpointTimeoutSpec(spec)
	for i := range eg.Endpoints {
		eg.Endpoints[i].mutableTimeoutPolicy().inheritDefaults(eg.Timeout)
		eg.Endpoints[i] = eg.Endpoints[i].withRebuiltHandler()
	}
}

func (eg EndpointGroup) TimeoutSpec() EndpointTimeoutSpec {
	return normalizeEndpointTimeoutSpec(eg.Timeout)
}

func (e Endpoint) TimeoutSpec() EndpointTimeoutSpec {
	return e.timeoutSpec()
}

func (e Endpoint) timeoutSpec() EndpointTimeoutSpec {
	return normalizeEndpointTimeoutSpec(e.timeout.timeout)
}

func (e *Endpoint) mutableTimeoutPolicy() *endpointTimeoutPolicy {
	return &e.timeout
}

func (p *endpointTimeoutPolicy) inheritDefaults(defaults EndpointTimeoutSpec) {
	defaults = normalizeEndpointTimeoutSpec(defaults)
	p.timeout = normalizeEndpointTimeoutSpec(p.timeout)

	if p.timeout.ReadBody == 0 || p.readBodyInherited {
		p.timeout.ReadBody = defaults.ReadBody
		p.readBodyInherited = defaults.ReadBody != 0
	}
	if p.timeout.Handler == 0 || p.handlerInherited {
		p.timeout.Handler = defaults.Handler
		p.handlerInherited = defaults.Handler != 0
	}
	if p.timeout.Write == 0 || p.writeInherited {
		p.timeout.Write = defaults.Write
		p.writeInherited = defaults.Write != 0
	}
}

func normalizeEndpointTimeoutSpec(spec EndpointTimeoutSpec) EndpointTimeoutSpec {
	if spec.ReadBody < 0 {
		panic(fmt.Sprintf(
			"httpapi: endpoint read body timeout cannot be negative: %s",
			spec.ReadBody,
		))
	}
	if spec.Handler < 0 {
		panic(fmt.Sprintf(
			"httpapi: endpoint handler timeout cannot be negative: %s",
			spec.Handler,
		))
	}
	if spec.Write < 0 {
		panic(fmt.Sprintf(
			"httpapi: endpoint write timeout cannot be negative: %s",
			spec.Write,
		))
	}

	return spec
}

func requestWithEndpointHandlerTimeout(
	req *http.Request,
	timeout EndpointTimeoutSpec,
	startedAt time.Time,
) (*http.Request, context.CancelFunc) {
	timeout = normalizeEndpointTimeoutSpec(timeout)
	if req == nil || timeout.Handler == 0 {
		return req, nil
	}

	ctx, cancel := context.WithDeadline(
		req.Context(),
		startedAt.Add(timeout.Handler),
	)
	return req.WithContext(ctx), cancel
}

func endpointTimedOut(req *Req) bool {
	if req == nil {
		return false
	}

	return errors.Is(req.Context().Err(), context.DeadlineExceeded)
}

func setEndpointReadBodyDeadline(
	w http.ResponseWriter,
	req *Req,
	timeout EndpointTimeoutSpec,
) bool {
	timeout = normalizeEndpointTimeoutSpec(timeout)
	if timeout.ReadBody == 0 {
		return false
	}

	return applyEndpointDeadline(
		w,
		req,
		time.Now().Add(timeout.ReadBody),
		"read_body",
		(*http.ResponseController).SetReadDeadline,
	)
}

func clearEndpointReadBodyDeadline(w http.ResponseWriter, req *Req) {
	applyEndpointDeadline(
		w,
		req,
		time.Time{},
		"read_body",
		(*http.ResponseController).SetReadDeadline,
	)
}

func setEndpointWriteDeadline(
	w http.ResponseWriter,
	req *Req,
	timeout EndpointTimeoutSpec,
) bool {
	timeout = normalizeEndpointTimeoutSpec(timeout)
	if timeout.Write == 0 {
		return false
	}

	return applyEndpointDeadline(
		w,
		req,
		time.Now().Add(timeout.Write),
		"write",
		(*http.ResponseController).SetWriteDeadline,
	)
}

func clearEndpointWriteDeadline(w http.ResponseWriter, req *Req) {
	applyEndpointDeadline(
		w,
		req,
		time.Time{},
		"write",
		(*http.ResponseController).SetWriteDeadline,
	)
}

func applyEndpointDeadline(
	w http.ResponseWriter,
	req *Req,
	deadline time.Time,
	phase string,
	set endpointDeadlineSetter,
) bool {
	if w == nil {
		return false
	}

	if err := set(http.NewResponseController(w), deadline); err != nil {
		if errors.Is(err, http.ErrNotSupported) {
			return false
		}

		logr.Printf(
			"failed to set endpoint %s deadline:"+
				" request_id=%s method=%s path=%s deadline=%v error=%v",
			phase,
			requestIDForDeadlineLog(req),
			methodForDeadlineLog(req),
			pathForDeadlineLog(req),
			deadline,
			err,
		)
		return false
	}

	return true
}

func requestIDForDeadlineLog(req *Req) string {
	if req == nil {
		return ""
	}

	return req.ID
}

func methodForDeadlineLog(req *Req) string {
	if req == nil || req.Req == nil {
		return ""
	}

	return req.Req.Method
}

func pathForDeadlineLog(req *Req) string {
	if req == nil || req.Req == nil || req.Req.URL == nil {
		return ""
	}

	return req.Req.URL.Path
}
