package plog

import "testing"

func TestInt(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Level
	}{
		{"全小写", "debug", LevelDebug},
		{"首字母大写", "Debug", LevelDebug},
		{"全大写", "DEBUG", LevelDebug},
		{"混合大小写", "wArN", LevelWarn},
		{"前后空白", "  info  ", LevelInfo},
		{"大写错误级别", "Error", LevelError},
		{"关闭级别", "NONE", LevelNone},
		{"空字符串", "", LevelNone},
		{"未知名称", "verbose", LevelNone},
		{"近似名称", "WARNING", LevelNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int(tt.in); got != tt.want {
				t.Fatalf("Int(%q) = %d, 期望 %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestLevelStringRoundTrip(t *testing.T) {
	for lvl := LevelDebug; lvl <= LevelNone; lvl++ {
		if got := Int(lvl.String()); got != lvl {
			t.Fatalf("Int(%q) = %d, 期望 %d", lvl.String(), got, lvl)
		}
	}
}
