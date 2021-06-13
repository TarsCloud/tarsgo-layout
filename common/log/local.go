package log

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/util/rogger"
)

var (
	jsonLog = tars.GetDayLogger("json", 3)

	logKey   = "M"
	levelKey = "C"
	timeKey  = "T"
	lineKey  = "L"

	kvKey       interface{} = struct{}{}
	enableDebug             = true
)

// Debug 写文件debug日志
func Debug(ctx context.Context, format string, args ...interface{}) {
	if !enableDebug {
		return
	}
	writef(ctx, rogger.DEBUG, format, args)
}

// Info 写文件error日志
func Info(ctx context.Context, format string, args ...interface{}) {
	writef(ctx, rogger.INFO, format, args)
}

// Error 写文件error日志
func Error(ctx context.Context, format string, args ...interface{}) {
	writef(ctx, rogger.ERROR, format, args)
}

// WithFields 将kv加入到日志中
func WithFields(ctx context.Context, kv ...interface{}) context.Context {
	vv := getKvFromContext(ctx)
	if vv == nil {
		vv = make(map[string]interface{})
	}
	for i := 1; i < len(kv); i += 2 {
		vv[fmt.Sprint(kv[i-1])] = kv[i]
	}
	return context.WithValue(ctx, kvKey, vv)
}

func getKvFromContext(ctx context.Context) map[string]interface{} {
	if val := ctx.Value(kvKey); val != nil {
		return val.(map[string]interface{})
	}
	return nil
}

func writef(ctx context.Context, level rogger.LogLevel, format string, args []interface{}) {
	logKv := map[string]interface{}{
		timeKey:  time.Now().Format("2006-01-02 15:04:05.000"),
		logKey:   fmt.Sprintf(format, args...),
		levelKey: level.String(),
	}
	kv := getKvFromContext(ctx)
	if kv != nil {
		for k, v := range kv {
			logKv[k] = v
		}
	}
	if level == rogger.DEBUG {
		pc, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "???"
			line = 0
		} else {
			file = filepath.Base(file)
		}
		logKv[lineKey] = fmt.Sprintf("%s:%s:%d", file, getFuncName(runtime.FuncForPC(pc).Name()), line)
	}

	bs, _ := json.Marshal(logKv)
	bs = append(bs, '\n')
	jsonLog.WriteLog(bs)
}

func getFuncName(name string) string {
	idx := strings.LastIndexByte(name, '/')
	if idx != -1 {
		name = name[idx:]
		idx = strings.IndexByte(name, '.')
		if idx != -1 {
			name = strings.TrimPrefix(name[idx:], ".")
		}
	}
	return name
}
