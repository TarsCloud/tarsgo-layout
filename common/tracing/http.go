package tracing

import (
	"context"
	"io"
	"net/http"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/defool/uuid"

	"github.com/opentracing/opentracing-go"
)

var (
	// DefaultClient is wraps the http.DefaultClient
	DefaultClient = &Client{http.DefaultClient}
)

// Client wraps http.Client
type Client struct {
	raw *http.Client
}

// Get wraps the http.Get
func Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return DoContext(ctx, req)
}

// Post wraps the http.Post
func Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return DoContext(ctx, req)
}

// DoContext wraps the http client.Do
func DoContext(ctx context.Context, req *http.Request) (rsp *http.Response, err error) {
	name := req.Method + " " + req.URL.String()
	defer NewFactory(ctx).NewSpan(name,
		PreExec(func(span opentracing.Span) {
			// 将tracing信息注入到http header中
			_ = opentracing.GlobalTracer().Inject(span.Context(),
				opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		}),
		PostExec(func(span opentracing.Span) {
			if err != nil {
				span.SetTag("error", err.Error())
			}
			span.SetTag("http.status", rsp.StatusCode)
		}),
	)()
	rsp, err = DefaultClient.raw.Do(req)
	return rsp, err
}

// SpanFromRequest start span from http request
func SpanFromRequest(r *http.Request) (opentracing.Span, string) {
	traceID := r.Header.Get("Uber-Trace-Id")
	if traceID == "" {
		traceID = uuid.UUID()
		r.Header.Set("Uber-Trace-Id", traceID)
	}
	var span opentracing.Span
	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	wireContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil {
		span = opentracing.StartSpan(r.URL.Path)
	} else {
		span = opentracing.StartSpan(r.URL.Path, opentracing.ChildOf(wireContext))
	}
	return span, traceID
}

// SpanFromRequest start span from http request
func SpanFromMap(name string, m map[string]string) opentracing.Span {
	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.TextMap,
		opentracing.TextMapCarrier(m))
	return opentracing.StartSpan(name, ext.SpanKindRPCServer, ext.RPCServerOption(spanCtx))
}
