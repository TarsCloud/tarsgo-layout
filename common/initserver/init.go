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
	"github.com/TarsCloud/TarsGo/tars/util/current"
	"github.com/opentracing/opentracing-go"

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
			var iReq, iRsp interface{} = req, rsp
			if len(req) == 1 {
				iReq = req[0]
			}
			if len(rsp) == 1 {
				iRsp = rsp[0]
			}
			rbs, _ := json.Marshal(iReq)
			sbs, _ := json.Marshal(iRsp)
			log.Info(ctx, "req is %s, rsp is %s", string(rbs), string(sbs))
		},

		// 客户端filter: 调用链
		clientFilter: func(ctx context.Context, msg *tars.Message, invoke tars.Invoke, timeout time.Duration) error {
			var invokeErr error

			// 传递status
			st, _ := current.GetRequestStatus(ctx)
			if msg.Req.Status == nil {
				msg.Req.Status = make(map[string]string)
			}
			for k, v := range st {
				msg.Req.Status[k] = v
			}

			// 调用链
			injectOpt := func(span opentracing.Span) {
				if msg.Req.Status == nil {
					msg.Req.Status = make(map[string]string)
				}
				opentracing.GlobalTracer().Inject(span.Context(),
					opentracing.TextMap, opentracing.TextMapCarrier(msg.Req.Status))
			}
			checkOpt := tracing.WithPostCheck(func(span opentracing.Span) error {
				if ip, ok := current.GetServerIPFromContext(ctx); ok {
					span.SetTag("peer.ipv4", ip)
				}
				if invokeErr != nil && !ecode.IsClientError(invokeErr) {
					return invokeErr
				}
				return nil
			})
			defer tracing.NewClientSpan(ctx, msg.Req.SFuncName, tracing.WithPre(injectOpt), checkOpt)()

			invokeErr = invoke(ctx, msg, timeout)
			return invokeErr
		},

		// 服务filter: 日志、调用链
		serverFilter: func(ctx context.Context, d tars.Dispatch, f interface{}, req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (invokeErr error) {
			// 日志
			cfg := tars.GetServerConfig()
			cIp, _ := current.GetClientIPFromContext(ctx)
			startTime := time.Now()
			logKv := []interface{}{
				"ServerName", cfg.Server,
				"SetName", cfg.Setdivision,
				"ServerIp", cfg.LocalIP,
				"ClientIp", cIp,
				"Action", req.SFuncName,
			}
			for k, v := range req.Status {
				if strings.HasPrefix(k, statusLogPrefix) {
					kk := k[len(statusLogPrefix):]
					logKv = append(logKv, kk)
					logKv = append(logKv, v)
				}
			}
			current.SetRequestStatus(ctx, req.Status)
			ctx = log.WithFields(ctx, logKv...)

			// 调用链
			preFunc := func(span opentracing.Span) {
				ctx = opentracing.ContextWithSpan(ctx, span)
			}
			checkFunc := func(span opentracing.Span) error {
				if invokeErr != nil && !ecode.IsClientError(invokeErr) {
					return invokeErr
				}
				return nil
			}
			defer tracing.NewServerSpanFromMap(req.SFuncName, req.Status, tracing.WithPre(preFunc), tracing.WithPostCheck(checkFunc))()

			defer func() {
				// recover处理
				if rr := recover(); rr != nil {
					resp.IRequestId = req.IRequestId
					invokeErr = ecode.Server("panic: %v", rr)
					resp.IRet = ecode.ServerError
					resp.SResultDesc = invokeErr.Error()

					buf := make([]byte, 16*1014)
					n := runtime.Stack(buf, false)
					log.Error(ctx, "%v\n%s", rr, string(buf[:n]))
				}

				// 日志
				logKv := []interface{}{
					"CostMS", (time.Now().UnixNano() - startTime.UnixNano()) / 1e6,
					"Code", tars.GetErrorCode(invokeErr),
				}
				if invokeErr != nil {
					logKv = append(logKv, "Error", invokeErr.Error())
				}
				ctx = log.WithFields(ctx, logKv...)
				log.Info(ctx, "done")
			}()

			// 业务逻辑
			invokeErr = d(ctx, f, req, resp, withContext)
			return invokeErr
		},
	}
}

// DoInit start the initialization of server
func (opt *initOption) DoInit() error {
	if opt.enableMetrics {
		go metrics.Listen()
		metrics.SetPrometheusStat()
	}
	if opt.enableTracing {
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
