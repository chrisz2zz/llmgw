// Package log 封装标准库 slog，提供日志器的初始化。
package log

import (
	"fmt"
	"log/slog"
	"os"
)

// New 根据日志级别字符串创建 *slog.Logger。
// 当前使用 TextHandler（开发期易读）；后续接入可观测性时可切换为 JSONHandler。
// 无效级别会返回 error，由调用方决定是否退出（fail fast）。
func New(level string) (*slog.Logger, error) {
	var l slog.Level
	err := l.UnmarshalText([]byte(level))
	if err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", level, err)
	}
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: l})
	return slog.New(h), nil
}
