// Package gormlog 提供把 plog 接入 GORM 的日志适配器。
//
// 这是一个可选子包：只有真正 import 它时，gorm（及其依赖链）才会被编译进二进制；
// 只使用 plog 核心日志功能的项目不会被引入任何第三方依赖。
//
// 使用示例：
//
//	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
//		Logger: gormlog.New(myLog),
//	})
//
// 或使用 plog 的默认日志器：
//
//	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
//		Logger: gormlog.NewDefault(),
//	})
package gormlog

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/phrynus/plog"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
)

// Logger 实现 gorm.io/gorm/logger.Interface，把 SQL 日志桥接到 plog。
//
// 字段与 gorm 官方 logger.Config 对齐，可在创建后按需调整。
type Logger struct {
	// LogLevel 记录级别，默认 gormLog.Info。
	LogLevel gormLog.LogLevel

	// SlowThreshold 慢 SQL 阈值，默认 200ms。
	SlowThreshold time.Duration

	// IgnoreRecordNotFoundError 是否忽略 gorm.ErrRecordNotFound，默认 true。
	IgnoreRecordNotFoundError bool

	// ShowCaller 是否在 SQL 日志中显示调用位置，继承父日志器的 FileLine 设置。
	ShowCaller bool

	log *plog.Log // 子日志器
}

// New 创建 GORM 日志适配器，日志写入 l。
//
// 内部使用 tag 为 GORM 的子日志器，并关闭其 file:line，
// 由适配器自行定位真正调用 SQL 的业务代码位置。
func New(l *plog.Log) *Logger {
	sub := l.Sub("GORM", plog.WithFileLine(false))

	return &Logger{
		LogLevel:                  gormLog.Info,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
		ShowCaller:                l.Config().FileLine, // 继承父日志器的 FileLine 设置
		log:                       sub,
	}
}

// NewDefault 等价于 New(plog.Default())，使用 plog 的包级默认日志器。
func NewDefault() *Logger { return New(plog.Default()) }

// LogMode 设置日志级别（实现 gormLog.Interface）。
// 返回副本，不影响原实例。
func (g *Logger) LogMode(level gormLog.LogLevel) gormLog.Interface {
	n := *g
	n.LogLevel = level
	return &n
}

// Info 实现 gormLog.Interface。
func (g *Logger) Info(_ context.Context, msg string, data ...any) {
	if g.LogLevel >= gormLog.Info {
		g.log.Debugf(msg, data...)
	}
}

// Warn 实现 gormLog.Interface。
func (g *Logger) Warn(_ context.Context, msg string, data ...any) {
	if g.LogLevel >= gormLog.Warn {
		g.log.Warnf(msg, data...)
	}
}

// Error 实现 gormLog.Interface。
func (g *Logger) Error(_ context.Context, msg string, data ...any) {
	if g.LogLevel >= gormLog.Error {
		g.log.Errorf(msg, data...)
	}
}

// Trace 实现 gormLog.Interface，记录 SQL 执行详情。
func (g *Logger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rows int64), err error) {
	if g.LogLevel <= gormLog.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	var caller string
	if g.ShowCaller {
		caller = g.callerInfo() + " "
	}

	switch {
	case err != nil && g.LogLevel >= gormLog.Error &&
		(!errors.Is(err, gorm.ErrRecordNotFound) || !g.IgnoreRecordNotFoundError):
		g.log.Warnf("%s[%.3fms] [rows:%d] %s | %v", caller, ms(elapsed), rows, sql, err)

	case elapsed > g.SlowThreshold && g.SlowThreshold > 0 && g.LogLevel >= gormLog.Warn:
		g.log.Warnf("%sSLOW(>=%v) [%.3fms] [rows:%d] %s", caller, g.SlowThreshold, ms(elapsed), rows, sql)

	case g.LogLevel >= gormLog.Info:
		g.log.Debugf("%s[%.3fms] [rows:%d] %s", caller, ms(elapsed), rows, sql)
	}
}

// callerInfo 返回调用 GORM 的业务代码位置，
// 跳过 gorm 自身与本库内部的栈帧。
func (g *Logger) callerInfo() string {
	var pcs [16]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for {
		fr, more := frames.Next()
		if !more {
			return ""
		}
		if strings.Contains(fr.File, "gorm.io/") ||
			strings.HasPrefix(fr.Function, plog.ModulePath+"/") ||
			strings.HasPrefix(fr.Function, plog.ModulePath+".") {
			continue
		}
		return fmt.Sprintf("%s:%d", shortFile(fr.File), fr.Line)
	}
}

// shortFile 返回不含目录的文件名，兼容 Windows 与 Unix 分隔符。
func shortFile(file string) string {
	if idx := strings.LastIndexAny(file, `/\`); idx >= 0 {
		return file[idx+1:]
	}
	return file
}

// ms 返回 float64 毫秒值。
func ms(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1e6
}
