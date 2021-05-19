package log

import (
	"encoding/json"

	"github.com/TarsCloud/TarsGo/tars/util/rogger"

	"github.com/TarsCloud/TarsGo/tars"
)

var (
	jsonLog  = tars.GetHourLogger("json", 24*2)
	debugLog = tars.GetLogger("debug")
	errorLog = tars.GetLogger("error")
)

// Write 将json格式的日志写入本地，用于远程上报
func Write(o interface{}) {
	bs, _ := json.Marshal(o)
	bs = append(bs, '\n')
	jsonLog.WriteLog(bs)
}

// Debug 写文件debug日志
func Debug(format string, args ...interface{}) {
	debugLog.Writef(0, rogger.DEBUG, format, args)
}

// Error 写文件error日志
func Error(format string, args ...interface{}) {
	errorLog.Writef(0, rogger.ERROR, format, args)
}
