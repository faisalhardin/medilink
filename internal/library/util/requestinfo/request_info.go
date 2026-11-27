package requestinfo

import (
	"context"
	"net/http"
	"time"
)

type RequestInfo struct {
	StartRequest time.Time
	Host         string
	SourceIP     string
	RequestURL   string
	Method       string
	Error        error
	HTTPStatus   int
	UserAgent    string
}

type ctxRequestInfo struct{}

func SetRequestContext(ctx context.Context, r *http.Request) context.Context {
	requestInfo := RequestInfo{
		StartRequest: time.Now(),
		Host:         r.Host,
		SourceIP:     r.RemoteAddr,
		// Never store the raw path: patient NIK and other identifiers can appear in URLs.
		RequestURL: "unknown",
		Method:       r.Method,
		UserAgent:    r.UserAgent(),
	}

	ctx = context.WithValue(ctx, ctxRequestInfo{}, requestInfo)
	return ctx
}

func GetRequestInfo(ctx context.Context) RequestInfo {
	requestInfo, ok := ctx.Value(ctxRequestInfo{}).(RequestInfo)
	if !ok {
		return RequestInfo{}
	}
	return requestInfo
}

// SetHTTPStatus is function set http status
func (req *RequestInfo) SetHTTPStatus(status int) {
	req.HTTPStatus = status
}
