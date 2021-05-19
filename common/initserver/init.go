package initserver

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/tarscloud/gopractice/common/ecode"
	"github.com/tarscloud/gopractice/common/log"
	"github.com/tarscloud/gopractice/common/metrics"
	"github.com/tarscloud/gopractice/common/remoteconf"
	"github.com/tarscloud/gopractice/common/tracing"
)

var (
	statusLogPrefix = "_log_status_"
	logKey          = "_tars_log_"
)

type initOption struct {
	dispatch     tars.DispatchReporter
	clientFilter tars.ClientFilter
	serverFilter tars.ServerFilter

	enableMetrics bool
	enableTracing bool

	configMap map[string]configCallback
}

type configCallback func(string) error

// NewOption ...
func NewOption() *initOption {
	return &initOption{
		enableMetrics: true,
		enableTracing: true,
		// 日志: 增加req/rsp
		dispatch: func(ctx context.Context, req []interface{}, rsp []interface{}, returns []interface{}) {
			logObj, ok := ctx.Value(logKey).(map[string]interface{})
			if !ok {
				log.Debug("log key not found in ctx")
				return
			}
			var iReq, iRsp interface{} = req, rsp
			if len(req) == 1 {
				iReq = req[0]
			}
			if len(rsp) == 1 {
				iRsp = rsp[0]
			}
			bs, _ := json.Marshal(iReq)
			logObj["Req"] = string(bs)
			bs, _ = json.Marshal(iRsp)
			logObj["Rsp"] = string(bs)
		},

		// 客户端filter: 调用链
		clientFilter: func(ctx context.Context, msg *tars.Message, invoke tars.Invoke, timeout time.Duration) error {
			var span opentracing.Span
			// 只有ctx有调用链才会注入
			if opentracing.SpanFromContext(ctx) != nil {
				span = opentracing.StartSpan(msg.Req.SFuncName,
					ext.SpanKindRPCClient,
				)
				// inject to context
				if msg.Req.Status == nil {
					msg.Req.Status = make(map[string]string)
				}
				_ = opentracing.GlobalTracer().Inject(span.Context(),
					opentracing.TextMap, opentracing.TextMapCarrier(msg.Req.Status))
			}
			defer func() {
				if span != nil {
					span.Finish()
				}
			}()
			err := invoke(ctx, msg, timeout)
			return err
		},

		// 服务filter: 日志、调用链
		serverFilter: func(ctx context.Context, d tars.Dispatch, f interface{}, req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) error {
			// 日志
			logObj := make(map[string]interface{})
			now := time.Now()
			logObj["StartMS"] = now.Format("2006-01-02 15:04:05")
			ctx = context.WithValue(ctx, logKey, logObj)
			startTime := now.UnixNano() / 1e6

			// 调用链
			spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.TextMap,
				opentracing.TextMapCarrier(req.Status))
			span := opentracing.StartSpan(req.SFuncName, ext.SpanKindRPCServer, ext.RPCServerOption(spanCtx))
			ctx = opentracing.ContextWithSpan(ctx, span)

			log.Debug("status is", req.Status)
			var invokeErr error
			defer func() {
				// recover处理
				if rr := recover(); rr != nil {
					resp.IRequestId = req.IRequestId
					invokeErr = ecode.Server("panic: %v", rr)

					buf := make([]byte, 16*1014)
					n := runtime.Stack(buf, false)
					log.Error("%v\n%s", rr, string(buf[:n]))
				}
				// 调用链
				if span != nil {
					span.Finish()
				}

				// 日志
				if invokeErr != nil {
					logObj["Code"] = tars.GetErrorCode(invokeErr)
					logObj["Error"] = invokeErr.Error()
				}
				logObj["CostMS"] = time.Now().UnixNano()/1e6 - startTime
				for k, v := range req.Status {
					if strings.HasPrefix(k, statusLogPrefix) {
						kk := k[len(statusLogPrefix):]
						logObj[kk] = v
					}
				}
				log.Write(logObj)
			}()

			// 业务逻辑
			invokeErr = d(ctx, f, req, resp, withContext)
			return invokeErr
		},
	}
}

func (opt *initOption) DoInit() error {
	if opt.enableMetrics {
		go metrics.Listen()
		metrics.SetPrometheusStat()
	}
	if opt.enableTracing {
		fmt.Println("EnableJaeger")
		tracing.EnableJaeger()
	}
	if opt.dispatch != nil {
		tars.RegisterDispatchReporter(opt.dispatch)
	}
	if opt.clientFilter != nil {
		tars.RegisterClientFilter(opt.clientFilter)
	}
	if opt.serverFilter != nil {
		tars.RegisterServerFilter(opt.serverFilter)
	}
	if opt.configMap != nil {
		cfg := tars.GetServerConfig()
		for name, callback := range opt.configMap {
			if err := remoteconf.DownloadConfig(cfg.BasePath, name); err != nil {
				return fmt.Errorf("download config %s error %v", name, err)
			}
			path := filepath.Join(cfg.BasePath, name)
			if err := callback(path); err != nil {
				return fmt.Errorf("init config %s error %v", name, err)
			}
		}
	}
	return nil
}

// WithRemoteConf ...
func (opt *initOption) WithRemoteConf(name string, callback configCallback) *initOption {
	if opt.configMap == nil {
		opt.configMap = make(map[string]configCallback)
	}
	opt.configMap[name] = callback
	return opt
}
