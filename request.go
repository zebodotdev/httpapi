package httpapi

import (
	"net/http"

	requestpkg "github.com/zebodotdev/httpapi/request"
)

const (
	unAuthzReqAppID = requestpkg.UnauthorizedAppID

	// intentionally case-insensitive. we're accessing the value through
	// http.Header.Get, which doesn't respect the case anyways.
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

	ReqTable    = requestpkg.ReqTable
	ReqPartKeyK = requestpkg.ReqPartKeyK
	ReqSortKeyK = requestpkg.ReqSortKeyK
	IdType      = requestpkg.IdType
)

type Req = requestpkg.Req
type RequestAudit = requestpkg.RequestAudit
type ResponseAudit = requestpkg.ResponseAudit

func NewReq(req *http.Request) *Req {
	return requestpkg.NewReq(req)
}

func genReqID() string {
	return requestpkg.NewID()
}
