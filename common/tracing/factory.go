package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

const maxSpanLength = 64

type spanOpt struct {
	isPreExec bool
	eFunc     func(span opentracing.Span)
}

// Factory is the tracing factory
type Factory struct {
	ctx context.Context
}

// NewFactory return the tracing factory
func NewFactory(ctx context.Context) *Factory {
	return &Factory{
		ctx: ctx,
	}
}

// NewSpan must return an finish function
func (t *Factory) NewSpan(name string, opts ...*spanOpt) func() {
	if len(name) > maxSpanLength {
		name = name[:maxSpanLength] + "..."
	}
	if opentracing.SpanFromContext(t.ctx) != nil {
		span, _ := opentracing.StartSpanFromContext(t.ctx, name)
		if span == nil {
			return func() {}
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
	return func() {}
}

// PreExec execute the func after span start
func PreExec(f func(span opentracing.Span)) *spanOpt {
	return &spanOpt{
		isPreExec: true,
		eFunc:     f,
	}
}

// PostExec execute the func before span finish
func PostExec(f func(span opentracing.Span)) *spanOpt {
	return &spanOpt{
		isPreExec: false,
		eFunc:     f,
	}
}
