# plog

简单轻量的 Go 日志库。支持文件输出、控制台彩色输出、按大小轮转与 gzip 压缩、异步写入，并内置 Gin / Fiber / GORM 集成。

## 特性

- 四级日志：`DEBUG` / `INFO` / `WARN` / `ERROR`
- 异步写入，格式化零额外字符串分配
- 按大小自动轮转，可选 gzip 压缩与过期备份清理
- 控制台 ANSI 彩色输出
- `Sub()` 子日志器，按 tag 区分模块
- 包级快捷函数，零配置即用

## 安装

```
go get github.com/phrynus/plog
```

## 快速开始

```go
package main

import "github.com/phrynus/plog"

func main() {
	defer plog.Close()
	plog.Info("hello plog")
	plog.Warnf("user %d not found", 42)
}
```

## 自定义配置

```go
l, _ := plog.New(plog.Config{
	Filename:    "app.log",
	MinLevel:    plog.Info,
	StdoutLevel: plog.Debug,
	Color:       true,
	FileLine:    true,
	MaxSize:     50 << 20, // 50MB
	Compress:    true,
	MaxBackups:  5,
})
defer l.Close()
l.Error("something went wrong")
```

## 模块子日志器

```go
db := l.Sub("DB")
db.Debugf("query: %s", sql)
```

## 框架集成

**Gin**

```go
r := gin.New()
r.Use(plog.GinLog(), plog.GinRecovery())
```

**Fiber**

```go
app := fiber.New()
app.Use(plog.FiberLog(), plog.FiberRecovery())
```

**GORM**

```go
db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
	Logger: myLog.GormLog(),
})
```

## License

MIT
