package httpapi

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	e "github.com/zebodotdev/httpapi/erreur"
)

const (
	authTypeBearer  = "bearer"
	authTypeService = "service"
	unAuthzReqAppID = "UNAUTHORIZED_REQUEST"

	// intentionally case-insensitive. we're accessing
	// the value through `r.Header.Get` method, which
	// doesn't respect the case anyways.
	contentTypeHeaderKey = "content-type"
	originHeaderKey      = "origin"
	authHeaderKey        = "authorization"
	fwdAuthHeaderKey     = "x-forwarded-authorization"
	idempotencyHeaderKey = "idempotency-key"
	xReqIDHeaderKey      = "x-request-id"
	xReqTimingHeaderKey  = "x-request-timing"
	traceParentHeaderKey = "traceparent"

	corsOriginHeaderKey  = "access-control-allow-origin"
	corsMethodsHeaderKey = "access-control-allow-methods"
	corsHeadersHeaderKey = "access-control-allow-headers"

	ReqTable    = "api_requests"
	ReqPartKeyK = "application_id"
	ReqSortKeyK = "id"
	IdType      = "req"
)

type jsonb map[string]string

// Req is the parameter we should pass to all endpoint
// handlers. the interior `Req` will contain all they need
// to successfully respond, then they'll write their response
// to the interior `Res` field.
type Req struct {
	AppID                string        `json:"application_id"`
	SessID               string        `json:"session_id,omitempty"`
	IdemKey              string        `json:"idempotency_key,omitzero"`
	RecdAt               time.Time     `json:"received_at"`
	Req                  *http.Request `json:"request,omitempty"`
	Body                 []byte        `json:"body"`
	Res                  *Res          `json:"response,omitempty"`
	Err                  *e.Error      `json:"error,omitempty"`
	Dur                  time.Duration `json:"duration,omitzero"`
	ID                   string        `json:"id"`
	Sess                 *Session      `json:"session"`
	AuthFailure          *AuthFailure  `json:"auth_failure,omitempty"`
	AuthorizationFailure *AuthFailure  `json:"authorization_failure,omitempty"`
}

type Res struct {
	ContentType string      `json:"content_type"`
	Status      int         `json:"status"`
	SentAt      time.Time   `json:"sent_at"`
	Header      http.Header `json:"header"`
	Body        any         `json:"body"`
	BodyReader  io.Reader   `json:"-"`
}

func (r *Req) partKey() string { return r.AppID }

func (r Req) MarshalJSON() ([]byte, error) {
	safeReq, safeh, reqURL := r.auditRequest()
	bdy := auditBody(r.Body)
	res := r.auditResponse()

	return json.Marshal(struct {
		URL                  string         `json:"url"`
		AppID                string         `json:"application_id"`
		SessID               string         `json:"session_id,omitempty"`
		IdemKey              string         `json:"idempotency_key,omitzero"`
		RecdAt               time.Time      `json:"received_at"`
		Header               http.Header    `json:"header,omitempty"`
		Req                  *RequestAudit  `json:"request,omitempty"`
		Body                 any            `json:"body,omitempty"`
		Res                  *ResponseAudit `json:"response,omitempty"`
		Err                  *e.Error       `json:"error,omitempty"`
		Dur                  time.Duration  `json:"duration"`
		ID                   string         `json:"id"`
		Auth                 AuthAudit      `json:"auth"`
		AuthFailure          *AuthFailure   `json:"auth_failure,omitempty"`
		AuthorizationFailure *AuthFailure   `json:"authorization_failure,omitempty"`
		Session              *SessionAudit  `json:"session,omitempty"`
	}{
		URL:                  reqURL,
		AppID:                r.AppID,
		SessID:               r.SessID,
		IdemKey:              auditSecret(r.IdemKey),
		RecdAt:               r.RecdAt,
		Body:                 bdy,
		Header:               safeh,
		Req:                  safeReq,
		Res:                  res,
		Err:                  r.Err,
		Dur:                  r.Dur,
		ID:                   r.ID,
		Auth:                 r.authAudit(),
		AuthFailure:          r.AuthFailure,
		AuthorizationFailure: r.AuthorizationFailure,
		Session:              sessionAudit(r.Sess),
	})
}

func (r *Req) Context() context.Context {
	if r == nil || r.Req == nil {
		return context.Background()
	}

	return r.Req.Context()
}

type RequestAudit struct {
	Method        string      `json:"method,omitempty"`
	URL           string      `json:"url,omitempty"`
	Path          string      `json:"path,omitempty"`
	RequestURI    string      `json:"request_uri,omitempty"`
	Host          string      `json:"host,omitempty"`
	RemoteAddr    string      `json:"remote_addr,omitempty"`
	UserAgent     string      `json:"user_agent,omitempty"`
	Referer       string      `json:"referer,omitempty"`
	Header        http.Header `json:"header,omitempty"`
	QueryRedacted bool        `json:"query_redacted,omitempty"`
}

type ResponseAudit struct {
	ContentType string      `json:"content_type,omitempty"`
	Status      int         `json:"status,omitempty"`
	SentAt      time.Time   `json:"sent_at,omitzero"`
	Header      http.Header `json:"header,omitempty"`
	BodyPresent bool        `json:"body_present,omitempty"`
	Streamed    bool        `json:"streamed,omitempty"`
}

func (r Req) auditRequest() (*RequestAudit, http.Header, string) {
	if r.Req == nil {
		return nil, nil, ""
	}

	header := auditHeader(r.Req.Header)
	url := auditURLString(r.Req)
	path := ""
	queryRedacted := false
	if r.Req.URL != nil {
		path = r.Req.URL.Path
		queryRedacted = r.Req.URL.RawQuery != ""
	}

	audit := &RequestAudit{
		Method:        r.Req.Method,
		URL:           url,
		Path:          path,
		RequestURI:    auditRequestURI(r.Req),
		Host:          r.Req.Host,
		RemoteAddr:    r.Req.RemoteAddr,
		UserAgent:     r.Req.UserAgent(),
		Referer:       auditRawURLString(r.Req.Referer()),
		Header:        header,
		QueryRedacted: queryRedacted,
	}

	return audit, header, url
}

func (r Req) auditResponse() *ResponseAudit {
	if r.Res == nil {
		return nil
	}

	return &ResponseAudit{
		ContentType: r.Res.ContentType,
		Status:      r.Res.Status,
		SentAt:      r.Res.SentAt,
		Header:      auditResponseHeader(r.Res.Header),
		BodyPresent: r.Res.Body != nil || r.Res.BodyReader != nil,
		Streamed:    r.Res.BodyReader != nil,
	}
}

func auditHeader(header http.Header) http.Header {
	out := make(http.Header)
	for _, key := range []string{
		contentTypeHeaderKey,
		"user-agent",
		"accept",
		originHeaderKey,
		xReqIDHeaderKey,
		xReqTimingHeaderKey,
		traceParentHeaderKey,
	} {
		copyHeader(out, header, key)
	}
	redactHeader(out, header, authHeaderKey)
	redactHeader(out, header, fwdAuthHeaderKey)
	redactHeader(out, header, idempotencyHeaderKey)
	redactHeader(out, header, "cookie")
	redactHeader(out, header, "proxy-authorization")
	redactHeader(out, header, "x-api-key")
	return out
}

func auditResponseHeader(header http.Header) http.Header {
	out := make(http.Header)
	for _, key := range []string{
		contentTypeHeaderKey,
		"cache-control",
		xReqIDHeaderKey,
		xReqTimingHeaderKey,
	} {
		copyHeader(out, header, key)
	}
	for _, key := range []string{
		"location",
		"set-cookie",
	} {
		redactHeader(out, header, key)
	}
	return out
}

func copyHeader(dst, src http.Header, key string) {
	if strings.TrimSpace(src.Get(key)) == "" {
		return
	}
	dst.Set(key, src.Get(key))
}

func redactHeader(dst, src http.Header, key string) {
	if strings.TrimSpace(src.Get(key)) == "" {
		return
	}

	dst.Set(key, "REDACTED")
}

func auditBody(body []byte) any {
	if len(body) == 0 {
		return nil
	}
	return "REDACTED"
}

func auditSecret(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return "REDACTED"
}

func auditURLString(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	u := auditURL(*req.URL)
	return u.String()
}

func auditRawURLString(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "REDACTED"
	}
	safeURL := auditURL(*u)
	return safeURL.String()
}

func auditURL(u url.URL) url.URL {
	u.RawQuery = ""
	u.ForceQuery = false
	u.Fragment = ""
	return u
}

func auditRequestURI(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	return req.URL.EscapedPath()
}

func (r *Req) Origin() string      { return r.Req.Header.Get("origin") }
func (r *Req) Path() string        { return r.Req.URL.Path }
func (r *Req) URL() string         { return r.Req.URL.String() }
func (r *Req) ContentType() string { return r.Req.Header.Get("content-type") }
func (r *Req) Authorization() string {
	auth := strings.TrimSpace(r.Req.Header.Get(fwdAuthHeaderKey))
	if auth == "" {
		auth = r.Req.Header.Get(authHeaderKey)
	}
	return auth
}
func (r *Req) DownstreamAuthorization() string {
	if r == nil {
		return ""
	}

	if r.Sess != nil && r.Sess.serviceSession() {
		token := strings.TrimSpace(r.Sess.Token)
		if token == "" {
			return ""
		}
		return currentAuthorizationSchemes().Bearer + " " + token
	}

	return r.Authorization()
}
func (r *Req) RemoteAddr() string  { return r.Req.RemoteAddr }
func (r *Req) Method() string      { return r.Req.Method }
func (r *Req) Referer() string     { return r.Req.Referer() }
func (r *Req) RequestBody() []byte { return r.Body }
func (r *Req) Errored() bool       { return r.Err != nil }
func (r *Req) Duration() *int64 {
	if r.Res == nil {
		return nil
	}

	diff := r.Res.SentAt.UnixMilli() - r.RecdAt.UnixMilli()
	return &diff
}
func (r *Req) UserAgent() string {
	ua := strings.TrimSpace(r.Req.UserAgent())
	if ua == "" {
		ua = "unknown_user_agent"
	}

	return ua
}

func (r *Req) ResponseHeaders() *jsonb {
	if r.Res == nil {
		return nil
	}
	combined := make(jsonb)
	for k, v := range r.Res.Header {
		combined[k] = strings.Join(
			v,
			"|",
		)
	}

	return &combined
}

func (r *Req) ResponseBody() ([]byte, error) {
	if r.Res.BodyReader != nil {
		return nil, fmt.Errorf("httpapi: streamed response body cannot be buffered")
	}
	if r.Res.ContentType == TextHTML {
		if body, ok := r.Res.Body.(string); ok {
			return []byte(body), nil
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(r.Res.Body); err != nil {
		logr.Printf(
			"failed to encode response as json:"+
				" request_id=%s error=%v",
			r.ID,
			err,
		)
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *Req) ResponseContentType() *string {
	if r.Res == nil {
		return nil
	}

	return &r.Res.ContentType
}

func (r *Req) ResponseStatus() *int {
	if r.Res == nil {
		return nil
	}

	return &r.Res.Status
}

func (r *Req) Headers() jsonb {
	combined := make(jsonb)
	for k, v := range r.Req.Header {
		combined[k] = strings.Join(
			v,
			" | ",
		)
	}
	return combined
}

func genReqID() string {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err == nil {
		return "req_" + hex.EncodeToString(buf)
	}

	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func NewReq(req *http.Request) *Req {
	if req.Body == nil {
		req.Body = http.NoBody
	}

	body, err := io.ReadAll(req.Body)
	if err != nil && err != io.EOF {
		logr.Printf(
			"error while reading request body:"+
				" remote_addr=%s user_agent=%s method=%s"+
				" request_url=%s content_type=%s error=%v",
			req.RemoteAddr, req.UserAgent(), req.Method,
			req.URL.String(), req.Header.Get("content-type"),
			err,
		)

		return nil
	}
	req.Body = io.NopCloser(bytes.NewReader(body))

	r := Req{
		AppID:  unAuthzReqAppID,
		RecdAt: time.Now(),
		Req:    req,
		Body:   body,
		ID:     genReqID(),
	}
	if session := SessionFromContext(req.Context()); session != nil {
		r.AttachSession(session)
		return &r
	}

	auth := r.Authorization()
	authScheme := strings.SplitN(auth, " ", 2)[0]
	schemes := currentAuthorizationSchemes()
	switch strings.ToLower(authScheme) {
	case strings.ToLower(schemes.Bearer):
		if err := r.authenticate(authTypeBearer); err != nil {
			r.recordAuthFailure(authTypeBearer, err)
			logr.Printf(
				"error while attempting to authenticate request:"+
					" remote_addr=%s user_agent=%s method=%s"+
					" request_url=%s content_type=%s error=%v",
				req.RemoteAddr, req.UserAgent(), req.Method,
				req.URL.String(), req.Header.Get("content-type"),
				err,
			)
		}
	case serviceAuthorizationSchemeCase(schemes, authScheme):
		if err := r.authenticate(authTypeService); err != nil {
			r.recordAuthFailure(authTypeService, err)
			logr.Printf(
				"error while attempting to authenticate service request:"+
					" remote_addr=%s user_agent=%s method=%s"+
					" request_url=%s content_type=%s error=%v",
				req.RemoteAddr, req.UserAgent(), req.Method,
				req.URL.String(), req.Header.Get("content-type"),
				err,
			)
		}
	default:
		if strings.TrimSpace(auth) != "" {
			r.recordAuthFailure(authTypeUnknown, ErrUnsupportedAuthorizationScheme)
		}
	}

	return &r
}

func serviceAuthorizationSchemeCase(schemes AuthorizationSchemes, candidate string) string {
	candidate = strings.ToLower(strings.TrimSpace(candidate))
	for _, scheme := range schemes.ServiceAuthorizationSchemes() {
		if candidate == strings.ToLower(scheme) {
			return candidate
		}
	}
	return "\x00"
}
