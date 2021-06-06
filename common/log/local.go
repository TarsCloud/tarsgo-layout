package log

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/TarsCloud/TarsGo/tars/util/rogger"

	"github.com/TarsCloud/TarsGo/tars"
)

var (
	jsonLog  = tars.GetHourLogger("json", 24*2)
	debugLog = tars.GetLogger("debug")
	errorLog = tars.GetLogger("error")

	logKey   = "M"
	levelKey = "L"
	timeKey  = "T"
	lineKey  = "L"
	kvKey    interface{}

	enableDebug = true
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
		vv = make(map[interface{}]interface{})
	}
	for i := 1; i < len(kv); i += 2 {
		vv[kv[i-1]] = kv[i]
	}
	return context.WithValue(ctx, kvKey, vv)
}

func getKvFromContext(ctx context.Context) map[interface{}]interface{} {
	if val := ctx.Value(kvKey); val != nil {
		return val.(map[interface{}]interface{})
	}
	return nil
}

func writef(ctx context.Context, level rogger.LogLevel, format string, args []interface{}) {
	kv := getKvFromContext(ctx)
	newKV := make(map[interface{}]interface{})
	if kv != nil {
		for k, v := range kv {
			newKV[k] = v
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
		newKV[lineKey] = fmt.Sprintf("%s:%s:%d|", file, getFuncName(runtime.FuncForPC(pc).Name()), line)
	}
	newKV[timeKey] = time.Now().Format("2006-01-02 15:04:05.000")
	newKV[logKey] = fmt.Sprintf(format, args...)
	newKV[levelKey] = level.String()
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
