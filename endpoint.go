package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	e "github.com/zebodotdev/httpapi/erreur"
)

type HttpMethod = string
type ContentType = string
type Handler func(r *Req)

const (
	APIRequestsTable = "api_requests"

	OPTIONS HttpMethod = "OPTIONS"
	POST    HttpMethod = "POST"
	GET     HttpMethod = "GET"

	ApplicationJson           ContentType = "application/json"
	ApplicationFormURLEncoded ContentType = "application/x-www-form-urlencoded"
	MultipartFormData         ContentType = "multipart/form-data"
	TextHTML                  ContentType = "text/html"
	TextPlain                 ContentType = "text/plain; charset=utf-8"

	TAG = "[httpapi/endpoint]: "
)

var logr = log.New(os.Stdout, TAG, log.Flags()|log.Llongfile)

// Endpoint couples a URL endpoint to its handler function.
type Endpoint struct {
	accepts    []ContentType
	method     HttpMethod
	pattern    string
	handler    http.HandlerFunc
	rawHandler Handler
	idempotent bool
	resolver   IdempotencyScopeResolver
	access     endpointAccessPolicy
	route      RouteSpec
	priority   endpointPriorityPolicy
	timeout    endpointTimeoutPolicy
	authKeys   map[string]bool

	/* note-to-self(yaw): this thing should be here */
	// requiresAuth bool
}

func (e Endpoint) Handler() http.HandlerFunc { return e.handler }
func (e Endpoint) Method() HttpMethod        { return e.method }
func (e Endpoint) Pattern() string           { return e.pattern }
func (e Endpoint) Accepts() ContentType      { return primaryContentType(e.accepts) }
func (e Endpoint) AcceptedContentTypes() []ContentType {
	return cloneContentTypes(e.accepts)
}
func (e Endpoint) IsInternal() bool   { return e.accessPolicy().internal }
func (e Endpoint) IsIdempotent() bool { return e.idempotent }
func (e Endpoint) Authorization() AuthorizationRequirement {
	return e.accessPolicy().auth
}
func (e Endpoint) RequiresAuthorization() bool { return e.Authorization().Required }
func (e Endpoint) RouteSpec() RouteSpec        { return e.routeSpec() }
func (e Endpoint) Priority() EndpointPriority  { return e.priorityPolicy().priority }
func (e Endpoint) AuthKeys() map[string]bool   { return cloneEndpointAuthKeys(e.authKeys) }

// EndpointGroup groups endpoints together under one path
// prefix. All endpoints in an EndpointGroup are usually
// concerned with one domain/service/etc.
type EndpointGroup struct {
	PathPrefix string
	Endpoints  []Endpoint
	Internal   bool
	Auth       AuthorizationRequirement
	Route      RouteSpec
	Priority   EndpointPriority
	Timeout    EndpointTimeoutSpec
}

func (eg EndpointGroup) Authorization() AuthorizationRequirement {
	return normalizeAuthorizationRequirement(eg.Auth)
}
func (eg EndpointGroup) RequiresAuthorization() bool { return eg.Auth.Required }
func (eg EndpointGroup) RouteSpec() RouteSpec        { return eg.Route }

// ResolvedEndpoints returns endpoints with group-level metadata applied.
// Transcribers and other read-only consumers should use this instead of
// inspecting Endpoints directly when group defaults matter.
func (eg EndpointGroup) ResolvedEndpoints() []Endpoint {
	if len(eg.Endpoints) == 0 {
		return nil
	}
	endpoints := make([]Endpoint, 0, len(eg.Endpoints))
	for _, endpoint := range eg.Endpoints {
		endpoints = append(endpoints, eg.endpointWithGroupMetadata(endpoint))
	}
	return endpoints
}

// Add adds a new endpoint to the group. At the moment, it doesn't
// ensure that duplicates are rejected. This means that if duplicate
// endpoints are added (by path pattern), then the latest will be
// the only effective handler.
func (eg *EndpointGroup) Add(e Endpoint) {
	eg.Endpoints = append(eg.Endpoints, eg.endpointWithGroupMetadata(e))
}

// Mount attaches all endpoints in the group to the given
// multiplexer. It prefixes the endpoint's pattern with the
// group's prefix so that an endpoint that was originally defined
// as, say, `/endpoint`, in a group with `group` prefix eventually
// becomes `/group/endpoint`.
func (eg *EndpointGroup) Mount(mux *http.ServeMux) {
	for _, g := range eg.Endpoints {
		g = eg.endpointWithGroupMetadata(g)
		path, _ := url.JoinPath(eg.PathPrefix, g.pattern)
		path, _ = url.PathUnescape(path)
		logr.Printf(
			"attaching endpoint to multiplexer:"+
				" method=%s path=%s accepts=%s",
			g.method, path, joinContentTypes(g.accepts),
		)

		mux.HandleFunc(
			fmt.Sprintf("%s %s", g.method, path),
			g.Handler(),
		)
	}
}

// NewEndpoint returns a server endpoint ready to handle
// requests in the way we approve of. That is, all in/outflows
// are recording for auditing, latency metrics are collected,
// etc.
func NewEndpoint(
	meth HttpMethod,
	pattern string,
	handler Handler,
	opts ...EndpointOption,
) Endpoint {
	return defineEndpointWithOptions(EndpointSpec{
		Method:  meth,
		Path:    pattern,
		Handler: handler,
	}, opts...)
}

func NewIdempotentEndpoint(
	meth HttpMethod,
	pattern string,
	handler Handler,
	opts ...EndpointOption,
) Endpoint {
	return defineEndpointWithOptions(EndpointSpec{
		Method:  meth,
		Path:    pattern,
		Handler: handler,
		Idempotency: EndpointIdempotencySpec{
			Enabled: true,
		},
	}, opts...)
}

func NewIdempotentEndpointWithScopeResolver(
	meth HttpMethod,
	pattern string,
	resolver IdempotencyScopeResolver,
	handler Handler,
	opts ...EndpointOption,
) Endpoint {
	return defineEndpointWithOptions(EndpointSpec{
		Method:  meth,
		Path:    pattern,
		Handler: handler,
		Idempotency: EndpointIdempotencySpec{
			Enabled:       true,
			ScopeResolver: resolver,
		},
	}, opts...)
}

func (endpoint Endpoint) withRebuiltHandler() Endpoint {
	endpoint.handler = endpoint.httpHandler()
	return endpoint
}

func (endpoint Endpoint) httpHandler() http.HandlerFunc {
	// close over the handler to only handle requests that have
	// the declared http method and content-type. it appears to me that there might
	// be a more elegant way to go about this task.
	return func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		auditContext := r.Context()
		timeout := endpoint.TimeoutSpec()
		r, cancelTimeout := requestWithEndpointHandlerTimeout(r, timeout, startedAt)
		if cancelTimeout != nil {
			defer cancelTimeout()
		}
		responseSizeBytes := 0
		req := &Req{
			AppID:  unAuthzReqAppID,
			RecdAt: startedAt,
			Req:    r,
			ID:     genReqID(),
		}
		defer func() {
			completedAt := time.Now()
			panicValue := recover()

			req.RecdAt = startedAt
			req.Dur = completedAt.Sub(startedAt)
			if req.Res != nil {
				req.Res.SentAt = completedAt
			} else if panicValue != nil {
				RenderErr(req, e.Unexpected())
			}

			if err := currentAuditSink().RecordRequest(auditContext, req); err != nil {
				logr.Printf(
					"attempt to record request audit failed"+
						" request_id=%s error=%v",
					req.ID,
					err,
				)
			}

			logr.Printf(
				"completed request handling:"+
					" request_method=%s request_path=%s request_id=%s"+
					" duration=%v response_size_bytes=%d",
				req.Method(), req.Path(), req.ID,
				time.Since(startedAt), responseSizeBytes,
			)

			if panicValue != nil {
				panic(panicValue)
			}
		}()

		if r.Method != endpoint.method {
			RenderErr(req, e.MethodNotAllowed(endpoint.method, r.Method))
			req.Dur = time.Since(startedAt)
			written, err := writeRenderedResponse(w, req, timeout)
			responseSizeBytes = written
			if err != nil {
				logr.Printf(
					"failed to write method not allowed response:"+
						" request_id=%s method=%s path=%s error=%v",
					req.ID, req.Method(), req.Path(), err,
				)
			}

			return
		}

		ct := ContentType(r.Header.Get(contentTypeHeaderKey))
		if r.Method != GET && ct != "" && validateEndpointContentType(ct, endpoint.accepts) != nil {
			RenderErr(req, e.UnsupportedContentType(string(ct), joinContentTypes(endpoint.accepts)))
			req.Dur = time.Since(startedAt)
			written, err := writeRenderedResponse(w, req, timeout)
			responseSizeBytes = written
			if err != nil {
				logr.Printf(
					"failed to write unsupported content type response:"+
						" request_id=%s method=%s path=%s error=%v",
					req.ID, req.Method(), req.Path(), err,
				)
			}

			return
		}

		readDeadlineSet := setEndpointReadBodyDeadline(w, req, timeout)
		parsedReq := NewReq(r)
		if readDeadlineSet {
			clearEndpointReadBodyDeadline(w, req)
		}
		if parsedReq == nil {
			RenderErr(req, e.InvalidRequestBody())
			req.Dur = time.Since(startedAt)
			written, err := writeRenderedResponse(w, req, timeout)
			responseSizeBytes = written
			if err != nil {
				logr.Printf(
					"failed to write invalid request body response:"+
						" request_id=%s method=%s path=%s error=%v",
					req.ID, req.Method(), req.Path(), err,
				)
			}

			return
		}
		req = parsedReq

		logr.Printf(
			"handling request: path=%s request=%s remote_addr=%s user_agent=%s",
			req.Path(), req.ID, req.RemoteAddr(),
			req.UserAgent(),
		)

		endpointHandler := endpoint.rawHandler
		if endpoint.idempotent {
			endpointHandler = func(req *Req) {
				handleIdempotently(
					req,
					endpoint.method,
					endpoint.pattern,
					endpoint.rawHandler,
					endpoint.resolver,
				)
			}
		}

		if accessErr := endpoint.accessError(req); accessErr != nil {
			RenderErr(req, accessErr)
		} else {
			endpointHandler(req)
		}

		if req.Res == nil {
			if endpointTimedOut(req) {
				RenderErr(req, e.RequestTimeout())
			} else {
				logr.Printf(
					"endpoint returned without setting response:"+
						" request_id=%s method=%s path=%s",
					req.ID, req.Method(), req.Path(),
				)
				RenderErr(req, e.Unexpected())
			}
		}

		req.Dur = time.Since(startedAt)
		if req.Res.BodyReader != nil {
			written, err := writeResponseStream(w, req, timeout)
			if written > 0 {
				responseSizeBytes = written
			}
			if err != nil {
				logr.Printf(
					"failed to write response stream:"+
						" request_id=%s method=%s path=%s error=%v",
					req.ID, req.Method(), req.Path(), err,
				)
			}
			return
		}

		body, err := req.ResponseBody()
		if err != nil {
			logr.Printf(
				"a really bad error occurred while encoding success"+
					" response to json and writing to connection: %v"+
					" returning internal server error instead.",
				err,
			)

			RenderErr(req, e.Unexpected())
			body, err = req.ResponseBody()
			if err != nil {
				logr.Printf(
					"failed to encode fallback error response:"+
						" request_id=%s method=%s path=%s error=%v",
					req.ID, req.Method(), req.Path(), err,
				)
				body = mustEncodeUnexpectedErr()
				responseSizeBytes = len(body)
				_, _ = writeResponseBody(w, req, body, timeout)

				return
			}
		}
		responseSizeBytes = len(body)

		written, err := writeResponseBody(w, req, body, timeout)
		if written > 0 {
			responseSizeBytes = written
		}
		if err != nil {
			logr.Printf(
				"failed to write response body:"+
					" request_id=%s method=%s path=%s error=%v",
				req.ID, req.Method(), req.Path(), err,
			)
		}
	}
}

func writeRenderedResponse(
	w http.ResponseWriter,
	req *Req,
	timeout EndpointTimeoutSpec,
) (int, error) {
	body, err := req.ResponseBody()
	if err != nil {
		return 0, err
	}

	written, err := writeResponseBody(w, req, body, timeout)
	if written > 0 {
		return written, err
	}
	return len(body), err
}

func writeResponseBody(
	w http.ResponseWriter,
	req *Req,
	body []byte,
	timeout EndpointTimeoutSpec,
) (int, error) {
	writeDeadlineSet := setEndpointWriteDeadline(w, req, timeout)
	if writeDeadlineSet {
		defer clearEndpointWriteDeadline(w, req)
	}

	rsh := w.Header()
	for k, v := range req.Res.Header {
		for _, hv := range v {
			rsh.Add(k, hv)
		}
	}

	rsh.Add(contentTypeHeaderKey, req.Res.ContentType)
	rsh.Add(xReqTimingHeaderKey, req.Dur.String())
	rsh.Add(xReqIDHeaderKey, req.ID)
	rsh.Add(corsOriginHeaderKey, "*")
	rsh.Add(corsMethodsHeaderKey, "*")
	rsh.Add(corsHeadersHeaderKey, "*")
	w.WriteHeader(req.Res.Status)

	req.Res.Header = rsh
	return w.Write(body)
}

func writeResponseStream(
	w http.ResponseWriter,
	req *Req,
	timeout EndpointTimeoutSpec,
) (int, error) {
	writeDeadlineSet := setEndpointWriteDeadline(w, req, timeout)
	if writeDeadlineSet {
		defer clearEndpointWriteDeadline(w, req)
	}

	rsh := w.Header()
	for k, v := range req.Res.Header {
		for _, hv := range v {
			rsh.Add(k, hv)
		}
	}

	rsh.Add(contentTypeHeaderKey, req.Res.ContentType)
	rsh.Add(xReqTimingHeaderKey, req.Dur.String())
	rsh.Add(xReqIDHeaderKey, req.ID)
	rsh.Add(corsOriginHeaderKey, "*")
	rsh.Add(corsMethodsHeaderKey, "*")
	rsh.Add(corsHeadersHeaderKey, "*")
	w.WriteHeader(req.Res.Status)

	req.Res.Header = rsh
	if closer, ok := req.Res.BodyReader.(io.Closer); ok {
		defer closer.Close()
	}
	written, err := io.Copy(w, req.Res.BodyReader)
	return int(written), err
}

func mustEncodeUnexpectedErr() []byte {
	body, err := json.Marshal(ErrRes{Err: e.Unexpected()})
	if err != nil {
		panic(err)
	}

	return append(body, '\n')
}
