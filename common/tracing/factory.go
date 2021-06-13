package tracing

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
)

const maxSpanLength = 64

type spanOpt struct {
	isPreExec bool
	eFunc     spanOptFunc
}

type spanOptFunc func(span opentracing.Span)

// NewServerSpanFromMap must return an finish function
func NewServerSpanFromHTTP(r *http.Request, opts ...*spanOpt) func() {
	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	return newSpan(spanCtx, false, r.URL.Path, opts...)
}

// NewServerSpanFromMap must return an finish function
func NewServerSpanFromMap(name string, m map[string]string, opts ...*spanOpt) func() {
	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.TextMap,
		opentracing.TextMapCarrier(m))
	return newSpan(spanCtx, false, name, opts...)
}

// NewClientSpan must return an finish function
func NewClientSpan(ctx context.Context, name string, opts ...*spanOpt) func() {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return func() {}
	}
	return newSpan(span.Context(), true, name, opts...)
}

// WithPre returns option for pre exec func
func WithPre(optFunc spanOptFunc) *spanOpt {
	return &spanOpt{
		isPreExec: true,
		eFunc:     optFunc,
	}
}

// WithPost returns option for post exec func
func WithPost(optFunc spanOptFunc) *spanOpt {
	return &spanOpt{
		isPreExec: false,
		eFunc:     optFunc,
	}
}

// WithPostCheck set span tag/logs if cFunc returns error
func WithPostCheck(cFunc func(span opentracing.Span) error) *spanOpt {
	return WithPost(func(span opentracing.Span) {
		if err := cFunc(span); err != nil {
			span.SetTag("error", true)
			span.LogFields(
				olog.String("message", err.Error()),
			)
		}
	})
}

func newSpan(spanCtx opentracing.SpanContext, isClient bool, name string, opts ...*spanOpt) func() {
	if len(name) > maxSpanLength {
		name = name[:maxSpanLength] + "..."
	}
	var span opentracing.Span
	if isClient {
		span = opentracing.StartSpan(name, ext.SpanKindRPCClient, opentracing.ChildOf(spanCtx))
	} else {
		span = opentracing.StartSpan(name, ext.SpanKindRPCServer, ext.RPCServerOption(spanCtx))
	}
	for _, f := range opts {
		if f.isPreExec {
			f.eFunc(span)
		}
	}
	return func() {
		for _, f := range opts {
			if !f.isPreExec {
				f.eFunc(span)
			}
		}
		span.Finish()
	}
}
