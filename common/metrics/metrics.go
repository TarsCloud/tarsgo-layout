package metrics

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/tarscloud/gopractice/common/log"

	"github.com/tarscloud/gopractice/common/ecode"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	labels = []string{"src_server", "server", "set", "action", "code"}

	requestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tars_rpc_request_total",
	}, labels)

	clientFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tars_rpc_client_failed",
	}, labels)

	serverFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tars_rpc_server_failed",
	}, labels)
	// TimeCostMS ...
	timeCostMS = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tars_rpc_cost_ms",
		Buckets: []float64{1, 5, 20, 50, 200, 500, 1000, 5000, 50000},
	}, labels)
)

// SetPrometheusStat set prometheus as stater
func SetPrometheusStat() {
	cfg := tars.GetServerConfig()
	if cfg == nil {
		log.Error("Only support in server")
		return
	}
	srcSvc := fmt.Sprintf("%s.%s", cfg.App, cfg.Server)
	log.Debug("SetPrometheusStat start")
	tars.ReportStat = func(msg *tars.Message, succ int32, timeout int32, exec int32) {
		code := "0"
		if msg.Resp != nil {
			code = strconv.Itoa(int(msg.Resp.IRet))
		}
		var svc = msg.Req.SServantName
		sNames := strings.Split(msg.Req.SServantName, ".")
		if len(sNames) > 2 {
			svc = fmt.Sprintf("%s.%s", sNames[0], sNames[1])
		}
		costMS := float64(msg.EndTime - msg.BeginTime)

		// 计数
		requestTotal.WithLabelValues(srcSvc, svc, cfg.Setdivision, msg.Req.SFuncName, code).Inc()
		if timeout > 0 || exec > 0 || msg.Resp == nil {
			serverFailed.WithLabelValues(srcSvc, svc, cfg.Setdivision, msg.Req.SFuncName, code).Inc()
		} else if msg.Resp != nil && msg.Resp.IRet != 0 {
			if ecode.IsClientErrorCode(msg.Resp.IRet) {
				clientFailed.WithLabelValues(srcSvc, svc, cfg.Setdivision, msg.Req.SFuncName, code).Inc()
			} else {
				serverFailed.WithLabelValues(srcSvc, svc, cfg.Setdivision, msg.Req.SFuncName, code).Inc()
			}
		}
		timeCostMS.WithLabelValues(srcSvc, svc, cfg.Setdivision, msg.Req.SFuncName, code).Observe(costMS)
	}
}

// Listen starts the prometheus handler
func Listen() {
	addr := ":8700-8800"
	if e := os.Getenv("PROMETHEUS_LISTEN_ADDR"); e != "" {
		addr = e
	}
	temps := strings.Split(addr, ":")
	if len(temps) != 2 {
		log.Error("bad format of addr %v", addr)
		return
	}
	sp := strings.Split(temps[1], "-")
	if len(sp) == 2 {
		startPort, _ := strconv.Atoi(sp[0])
		endPort, _ := strconv.Atoi(sp[1])
		port, err := getRandomPort(temps[0], "tcp", startPort, endPort)
		if err != nil {
			log.Error("getRandomPort error %v", addr)
			return
		}
		addr = fmt.Sprintf("%s:%d", temps[0], port)
	}
	log.Debug("prometheus listen on %s", addr)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Error("prometheus listen error %v", err)
	}
}
