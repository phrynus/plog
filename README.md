# plog

简单轻量的 Go 日志库。支持文件输出、控制台彩色输出、按大小轮转与 gzip 压缩、异步写入，并提供 Gin / Fiber / GORM 可选适配子包。

## 特性

- 四级日志：`DEBUG` / `INFO` / `WARN` / `ERROR`
- 异步写入，格式化零额外字符串分配
- 按大小自动轮转，可选 gzip 压缩与过期备份清理
- 控制台 ANSI 彩色输出
- `Sub()` 子日志器，按 tag 区分模块
- 包级快捷函数，零配置即用
- **核心包零第三方依赖**，Web 框架适配器独立成子包按需引入（不引入则不占体积）

## 安装

```
go get github.com/phrynus/plog
```

## 快速开始

```go
package plog

import "github.com/phrynus/plog"

func main() {
	defer plog.Close()
	plog.Info("hello plog")
	plog.Warnf("user %d not found", 42)
}
```

## 默认配置

使用包级快捷函数前，可用 `SetDefault` 自定义默认日志器（必须在首次调用前设置，推荐放在 `init()` 中）。

```go
func init() {
	plog.SetDefault(plog.Config{
		Filename:    "app.log",       // 日志文件路径（必填）
		MaxSize:     50 << 20,        // 单个文件最大 50MB，超过自动轮转
		MinLevel:    plog.LevelDebug, // 最低记录级别 Debug
		StdoutLevel: 1,               // 控制台输出最低级别 Info
		Color:       true,            // 控制台 ANSI 彩色
		FileLine:    true,            // 记录调用位置 file:line
		Tag:         "MAIN",          // 日志标识符
		BufferSize:  4096,            // 异步缓冲大小
		Compress:    true,            // 轮转后 gzip 压缩
		MaxBackups:  5,               // 保留的压缩备份数
	})
}
```

## 自定义配置

```go
l, _ := plog.New(plog.Config{
	Filename:    "app.log",       // 日志文件路径（必填）
	MaxSize:     50 << 20,        // 单个文件最大 50MB，超过自动轮转
	MinLevel:    plog.LevelDebug, // 最低记录级别 Debug
	StdoutLevel: 1,               // 控制台输出最低级别 Info
	Color:       true,            // 控制台 ANSI 彩色
	FileLine:    true,            // 记录调用位置 file:line
	Tag:         "MAIN",          // 日志标识符
	BufferSize:  4096,            // 异步缓冲大小
	Compress:    true,            // 轮转后 gzip 压缩
	MaxBackups:  5,               // 保留的压缩备份数
})
defer l.Close()
l.Error("something went wrong")
```

| 字段          | 默认值  | 说明                                  |
| ------------- | ------- | ------------------------------------- |
| `Filename`    | —       | 日志文件路径（必填），目录自动创建    |
| `MaxSize`     | 50MB    | 单个文件超过后自动轮转                |
| `MinLevel`    | `DEBUG` | 低于该级别的日志被丢弃                |
| `StdoutLevel` | `INFO`  | 控制台输出最低级别，`NONE` 关闭控制台 |
| `Color`       | `true`  | 控制台 ANSI 彩色输出                  |
| `FileLine`    | `true`  | 在日志中记录调用位置                  |
| `Tag`         | `MAIN`  | 日志标识符，出现在 `[TAG]`            |
| `BufferSize`  | 4096    | 异步缓冲大小，满时丢弃兜底            |
| `Compress`    | `false` | 轮转后是否 gzip 压缩旧文件            |
| `MaxBackups`  | 5       | 保留的压缩备份数量                    |

## 级别名称转换

`plog.Int` 把级别名称转成 `Level`，大小写不敏感（`debug` / `Debug` / `DEBUG` 等价），
名称前后空白会被忽略，无法识别时返回 `LevelNone`（`4`），即关闭输出。
常用于从配置或环境变量读取级别：

```go
// 无法识别的名称返回 LevelNone，不会输出任何日志
cfg.MinLevel = plog.Int(os.Getenv("LOG_LEVEL"))

plog.Int("debug") // LevelDebug
plog.Int("Warn")  // LevelWarn
plog.Int("NONE")  // LevelNone
plog.Int("???")   // LevelNone
```

## 模块子日志器

```go
// 子日志器与父日志器写同一文件、共享异步通道，只有 tag 不同。
db := l.Sub("DB")
db.Debugf("query: %s", sql)
```

子日志器默认继承父日志器的全部配置，可用 `SubOption` 覆盖：

```go
// 例如让这个子日志器不记录 file:line
trace := l.Sub("TRACE", plog.WithFileLine(false))
```

## 框架集成

Gin / Fiber / GORM 适配器都是**可选子包**：只有真正 `import` 时才会把对应框架编译进二进制。
只写文件日志的项目不会再被拖入 Web 框架全家桶。

**Gin**（`github.com/phrynus/plog/ginlog`）

```go
import "github.com/phrynus/plog/ginlog"

r := gin.New()
r.Use(ginlog.Log(myLog), ginlog.Recovery(myLog))

// 使用包级默认日志器：
r.Use(ginlog.LogDefault(), ginlog.RecoveryDefault())
```

**Fiber**（`github.com/phrynus/plog/fiberlog`）

```go
import "github.com/phrynus/plog/fiberlog"

app := fiber.New()
app.Use(fiberlog.Log(myLog), fiberlog.Recovery(myLog))

// 使用包级默认日志器：
app.Use(fiberlog.LogDefault(), fiberlog.RecoveryDefault())
```

**GORM**（`github.com/phrynus/plog/gormlog`）

```go
import "github.com/phrynus/plog/gormlog"

db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
	Logger: gormlog.New(myLog), // 或 gormlog.NewDefault()
})
```

`gormlog.Logger` 的字段可直接调整：

```go
lg := gormlog.New(myLog)
lg.SlowThreshold = 500 * time.Millisecond
lg.IgnoreRecordNotFoundError = false
db, _ := gorm.Open(dialector, &gorm.Config{Logger: lg})
```
