/**
* @Author: TongTongLiu
* @Date: 2019-08-07 20:26
**/

package libs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"golang.org/x/net/context"

	"api-test/configs"
)

var (
	commonLog *Logger
	trailLog  *Logger
)

func InitLog() {
	configs.FlagParse()
	commonLog = NewLogger("", configs.FlagLogPidName)
	trailLog = NewLogger("msgTrail", configs.FlagLogPidName)
}

type Logger struct {
	beeLogger *logs.BeeLogger
}

func NewLogger(prefix string, postfix string) *Logger {
	logger := &Logger{}
	log := logs.NewLogger(1000)
	//debug := configs.AppDebug()
	logRunTime := configs.GetGlobalConfig().Log.Runtime
	if logRunTime {
		log.EnableFuncCallDepth(true)
		log.SetLogFuncCallDepth(3)
		log.EnableFuncCallDepth(true)
	}
	seg := []string{configs.GetGlobalConfig().Log.Dir + "/"}
	if prefix != "" {
		seg = append(seg, prefix+".")
	}
	seg = append(seg, configs.GetAppName())
	if postfix != "" {
		seg = append(seg, "."+postfix)
	}
	logFile := strings.Join(seg, "")
	logLevel := logs.LevelInfo
	if configs.GetGlobalConfig().Log.Debug == true {
		logLevel = logs.LevelDebug
	}
	log_config := map[string]interface{}{
		"filename": logFile,
		"maxsize":  1 << 28,
		"daily":    true,
		"maxdays":  30,
		"level":    logLevel,
		"perm":     "775",
	}
	fileConfig, _ := json.Marshal(log_config)
	log.SetLogger("file", string(fileConfig))
	log.Info("version is v19.11.18")
	log.Info("log path is (%v)", logFile)
	fmt.Println("log path is", logFile)
	logger.beeLogger = log
	return logger
}

func (l *Logger) extractTraceId(ctx context.Context) string {
	return ExtractTraceId(ctx)
}

func (l *Logger) Info(ctx context.Context, format string, v ...interface{}) {
	traceId := l.extractTraceId(ctx)
	l.beeLogger.Info(fmt.Sprintf("{traceId: %v} ", traceId)+format, v...)
}

func (l *Logger) Error(ctx context.Context, format string, v ...interface{}) {
	traceId := l.extractTraceId(ctx)
	l.beeLogger.Error(fmt.Sprintf("{traceId: %v} ", traceId)+format, v...)
}

func (l *Logger) Warn(ctx context.Context, format string, v ...interface{}) {
	traceId := l.extractTraceId(ctx)
	l.beeLogger.Warn(fmt.Sprintf("{traceId: %v} ", traceId)+format, v...)
}

func (l *Logger) Debug(ctx context.Context, format string, v ...interface{}) {
	traceId := l.extractTraceId(ctx)
	l.beeLogger.Debug(fmt.Sprintf("{traceId: %v} ", traceId)+format, v...)
}

func Log() *Logger {
	if commonLog == nil {
		commonLog = NewLogger("", configs.FlagLogPidName)
	}
	return commonLog
}

func TrailLog() *Logger {
	if trailLog == nil {
		trailLog = NewLogger("msgTrail", configs.FlagLogPidName)
	}
	return trailLog
}
