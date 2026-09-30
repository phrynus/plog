// Package ginlog 提供把 plog 接入 Gin 的中间件。
//
// 这是一个可选子包：只有真正 import 它时，gin 才会被编译进二进制；
// 只使用 plog 核心日志功能的项目不会被引入任何第三方依赖。
//
// 使用示例：
//
//	r := gin.New()
//	r.Use(ginlog.Log(myLog), ginlog.Recovery(myLog))
//
// 或使用 plog 的默认日志器：
//
//	r.Use(ginlog.LogDefault(), ginlog.RecoveryDefault())
package ginlog

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phrynus/plog"
)

// Log 返回一个 Gin 中间件，将 HTTP 请求日志记入 l。
//
// 状态码、耗时、客户端 IP、方法、路径会被记录；
// 命中 c.Errors 或状态码 >= 400 / >= 500 时自动升级为 WARN / ERROR。
//
// 中间件运行在框架调用链内部，file:line 只会指向 gin 内部实现，
// 因此子日志器关闭 FileLine（同时省去每个请求一次栈回溯）。
func Log(l *plog.Log) gin.HandlerFunc {
	sub := l.Sub("GIN", plog.WithFileLine(false))

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		code := c.Writer.Status()
		ip := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		errMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()

		switch {
		case errMsg != "":
			sub.Warnf("[%d] %s %s | %v | %s | %s", code, method, path, latency, ip, errMsg)
		case code >= 500:
			sub.Errorf("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		case code >= 400:
			sub.Warnf("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		default:
			sub.Infof("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		}
	}
}

// LogDefault 等价于 Log(plog.Default())，使用 plog 的包级默认日志器。
func LogDefault() gin.HandlerFunc { return Log(plog.Default()) }

// Recovery 返回一个 Gin 中间件，捕获 panic 并记入 l。
func Recovery(l *plog.Log) gin.HandlerFunc {
	sub := l.Sub("GIN", plog.WithFileLine(false))

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				sub.Errorf("PANIC: %v | %s %s | IP: %s", r, c.Request.Method, c.Request.URL.Path, c.ClientIP())
				c.AbortWithStatusJSON(500, gin.H{
					"code":    500,
					"message": "internal server error",
				})
			}
		}()
		c.Next()
	}
}

// RecoveryDefault 等价于 Recovery(plog.Default())，使用 plog 的包级默认日志器。
func RecoveryDefault() gin.HandlerFunc { return Recovery(plog.Default()) }
