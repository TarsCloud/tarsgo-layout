package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/basef"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/util/current"
	"github.com/defool/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"

	"github.com/tarscloud/gopractice/apps/jsontarsproxy/config"
	"github.com/tarscloud/gopractice/common/ecode"
	"github.com/tarscloud/gopractice/common/log"
	"github.com/tarscloud/gopractice/common/tracing"
)

var (
	actionPrefix = "/apis/v1/"
	logKeyPrefix = "_log_status_"

	servantCache = make(map[string]*tars.ServantProxy)
	servantLock  = sync.RWMutex{}

	commCache = make(map[string]*tars.Communicator)
	commLock  = sync.RWMutex{}
)

type response struct {
	RequestId string
	Code      int32
	Error     string
	Data      map[string]interface{}
}

// HandlerFunc ...
func HandlerFunc(w http.ResponseWriter, r *http.Request) {
	// 初始化
	var tarsStatus = make(map[string]string)
	var reqBody []byte
	var actionName string
	var startTime = time.Now().UnixNano() / 1e6
	rsp := response{}
	ctx := context.Background()

	// 调用链
	ctx = current.ContextWithClientCurrent(ctx)
	preFunc := func(span opentracing.Span) {
		if jSpan, ok := span.(*jaeger.Span); ok {
			rsp.RequestId = jSpan.SpanContext().TraceID().String()
		} else {
			rsp.RequestId = uuid.UUID()
		}
		ctx = opentracing.ContextWithSpan(ctx, span)
	}
	checkFunc := func(span opentracing.Span) error {
		if actionName != "" {
			span.SetOperationName(actionName)
		}
		if rsp.Code != 0 && !ecode.IsClientErrorCode(rsp.Code) {
			return fmt.Errorf("dode=%d, error=%s", rsp.Code, rsp.Error)
		}
		return nil
	}
	defer tracing.NewServerSpanFromHTTP(r, tracing.WithPre(preFunc), tracing.WithPostCheck(checkFunc))()

	// 日志
	cfg := tars.GetServerConfig()
	ctx = log.WithFields(ctx,
		"ServerName", cfg.Server,
		"SetName", cfg.Setdivision,
		"ServerIp", cfg.LocalIP,
		"ReqId", rsp.RequestId,
		"ClientIp", strings.Split(r.RemoteAddr, ":")[0],
	)

	defer func() {
		if err := recover(); err != nil {
			rsp.Code = ecode.ServerError
			rsp.Error = "panic: " + fmt.Sprint(err)
		}
		rspData := rsp.write(ctx, w)

		// 日志
		logKv := []interface{}{
			"Action", actionName,
			"Code", rsp.Code,
			"Error", rsp.Error,
			"CostMS", time.Now().UnixNano()/1e6 - startTime,
		}
		for k, v := range tarsStatus {
			if strings.HasPrefix(k, logKeyPrefix) {
				logKv = append(logKv, k[len(logKeyPrefix):])
				logKv = append(logKv, v)
			}
		}
		ctx = log.WithFields(ctx, logKv...)
		log.Info(ctx, "req is %s, rsp is %s", string(reqBody), string(rspData))
	}()

	var err error
	reqBody, err = ioutil.ReadAll(r.Body)
	if err != nil {
		rsp.Code, rsp.Error = ecode.ClientError, "Read request error "+err.Error()
		return
	}

	reqBodyMap := make(map[string]interface{})
	err = json.Unmarshal(reqBody, &reqBodyMap)
	if err != nil {
		rsp.Code, rsp.Error = ecode.ClientError, "Request body is not json "+err.Error()
		return
	}

	reqJson := make(map[string]interface{})
	reqJson["req"] = reqBodyMap
	reqByte, _ := json.Marshal(reqJson)

	if !strings.HasPrefix(r.URL.Path, actionPrefix) {
		rsp.Code, rsp.Error = ecode.ClientError, "Request path should starts with "+actionPrefix
		return
	}
	actionName = r.URL.Path[len(actionPrefix):]
	sv := getServantProxy(actionName)
	if sv == nil {
		rsp.Code, rsp.Error = ecode.ClientError, fmt.Sprintf("Action '%s' not found", actionName)
		return
	}
	var resp = &requestf.ResponsePacket{}
	var tarsContext map[string]string
	var hashVal interface{}
	actionConf, _ := config.Get().ActionMap[actionName]
	for k, v := range reqBodyMap {
		if config.Get().Logging.ReqFields[k] {
			kk := logKeyPrefix + k
			tarsStatus[kk] = fmt.Sprint(v)
		}
		if k == actionConf.HashKey {
			hashVal = v
		}
	}
	tarsStatus[logKeyPrefix+"ReqId"] = rsp.RequestId

	// 按一致性hash调用
	if hashVal != nil {
		ctx = current.ContextWithClientCurrent(ctx)
		current.SetClientHash(ctx, int(tars.ConsistentHash), hashCode(hashVal))
	}

	if err := sv.Tars_invoke(ctx, 0, actionName, reqByte, tarsStatus, tarsContext, resp); err != nil {
		code := tars.GetErrorCode(err)
		if code == 1 {
			code = ecode.ServerError
		}
		rsp.Code, rsp.Error = code, "Tars invoke error "+err.Error()
		return
	}
	rspByte := codec.FromInt8(resp.SBuffer)
	jsonRsp := make(map[string]interface{})
	if err := json.Unmarshal(rspByte, &jsonRsp); err != nil {
		rsp.Code, rsp.Error = ecode.ServerError, fmt.Sprintf("Unmarshal rspByte error %v", err)
		return
	}
	rspData, ok := jsonRsp["rsp"].(map[string]interface{})
	if !ok {
		rsp.Code, rsp.Error = ecode.ServerError, "`rsp` not found in response"
		return
	}
	rsp.Data = rspData
}

func hashCode(s interface{}) uint32 {
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprint(s)))
	return h.Sum32()
}

func (r *response) write(ctx context.Context, w http.ResponseWriter) []byte {
	bs, _ := json.Marshal(r)
	_, err := w.Write(bs)
	if err != nil {
		log.Error(ctx, "Write error %v", err)
	}
	return bs
}

func getComm(setName string) *tars.Communicator {
	commLock.Lock()
	defer commLock.Unlock()
	if v, ok := commCache[setName]; ok {
		return v
	}

	comm := tars.NewCommunicator()
	if setName != "" {
		comm.SetProperty("enableset", true)
		comm.SetProperty("setdivision", setName)
	}
	commCache[setName] = comm
	return comm
}

func getServantProxy(action string) *tars.ServantProxy {
	actionConf, ok := config.Get().ActionMap[action]
	if !ok {
		return nil
	}
	cacheKey := fmt.Sprintf("%s-%s-%d", actionConf.Cluster, actionConf.Addr, actionConf.TimeoutMS)
	servantLock.Lock()
	defer servantLock.Unlock()
	if v, ok := servantCache[cacheKey]; ok {
		return v
	}
	// get from cache
	comm := getComm(actionConf.Cluster)
	sv := tars.NewServantProxy(comm, actionConf.Addr)
	sv.TarsSetTimeout(actionConf.TimeoutMS)
	sv.TarsSetVersion(basef.JSONVERSION)
	servantCache[cacheKey] = sv
	return sv
}
