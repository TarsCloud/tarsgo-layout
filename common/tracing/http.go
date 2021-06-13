package tracing

import (
	"context"
	"fmt"
	"io"
	"net/http"

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
	injectFunc := func(span opentracing.Span) {
		// 将tracing信息注入到http header中
		_ = opentracing.GlobalTracer().Inject(span.Context(),
			opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	}
	checkFunc := func(span opentracing.Span) error {
		if err != nil {
			return err
		}
		if rsp.StatusCode >= 500 {
			return fmt.Errorf("http status code %d", rsp.StatusCode)
		}
		return nil
	}
	defer NewClientSpan(ctx, name,
		WithPre(injectFunc),
		WithPostCheck(checkFunc),
	)()
	rsp, err = DefaultClient.raw.Do(req)
	return rsp, err
}
