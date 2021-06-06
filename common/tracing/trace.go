package tracing

import (
	"context"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/opentracing/opentracing-go"
	"github.com/tarscloud/gopractice/common/log"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

// EnableJaeger set jaeger as global tracer
func EnableJaeger() {
	var svcName string
	if cf := tars.GetServerConfig(); cf != nil {
		svcName = cf.Server
	}
	cfg := &jaegercfg.Configuration{
		ServiceName: svcName,
	}
	cfg, err := cfg.FromEnv()
	if err != nil {
		log.Error(context.Background(), "Could not parse jaeger env vars: %s", err.Error())
		return
	}
	tracer, _, err := cfg.NewTracer()
	if err != nil {
		log.Error(context.Background(), "Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	log.Debug(context.Background(), "Enable jaeger")
	opentracing.SetGlobalTracer(tracer)
}
