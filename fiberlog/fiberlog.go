// Package fiberlog 提供把 plog 接入 Fiber 的中间件。
//
// 这是一个可选子包：只有真正 import 它时，fiber（及其 fasthttp / quic-go 依赖链）
// 才会被编译进二进制；只使用 plog 核心日志功能的项目不会被引入任何第三方依赖。
//
// 使用示例：
//
//	app := fiber.New()
//	app.Use(fiberlog.Log(myLog), fiberlog.Recovery(myLog))
//
// 或使用 plog 的默认日志器：
//
//	app.Use(fiberlog.LogDefault(), fiberlog.RecoveryDefault())
package fiberlog

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/phrynus/plog"
)

// Log 返回一个 Fiber 中间件，将 HTTP 请求日志记入 l。
//
// 状态码、耗时、客户端 IP、方法、路径会被记录；
// 处理链返回错误或状态码 >= 400 / >= 500 时自动升级为 WARN / ERROR。
//
// 中间件运行在框架调用链内部，file:line 只会指向 fiber 内部实现，
// 因此子日志器关闭 FileLine（同时省去每个请求一次栈回溯）。
func Log(l *plog.Log) fiber.Handler {
	sub := l.Sub("FIBER", plog.WithFileLine(false))

	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		code := c.Response().StatusCode()
		ip := c.IP()
		method := c.Method()
		path := c.Path()

		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}

		switch {
		case errMsg != "":
			sub.Warnf("[%d] %s %s | %v | %s | %s", code, method, path, latency, ip, errMsg)
		case code >= 500:
			sub.Errorf("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		case code >= 400:
			sub.Warnf("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		default:
			sub.Debugf("[%d] %s %s | %v | %s", code, method, path, latency, ip)
		}

		return err
	}
}

// LogDefault 等价于 Log(plog.Default())，使用 plog 的包级默认日志器。
func LogDefault() fiber.Handler { return Log(plog.Default()) }

// Recovery 返回一个 Fiber 中间件，捕获 panic 并记入 l。
func Recovery(l *plog.Log) fiber.Handler {
	sub := l.Sub("FIBER", plog.WithFileLine(false))

	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				sub.Errorf("PANIC: %v | %s %s | IP: %s", r, c.Method(), c.Path(), c.IP())
				_ = c.Status(500).JSON(fiber.Map{
					"code":    500,
					"message": "internal server error",
				})
			}
		}()
		return c.Next()
	}
}

// RecoveryDefault 等价于 Recovery(plog.Default())，使用 plog 的包级默认日志器。
func RecoveryDefault() fiber.Handler { return Recovery(plog.Default()) }
