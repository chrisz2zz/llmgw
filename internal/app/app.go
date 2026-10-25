// Package app 负责装配网关的各个依赖，并管理 HTTP 服务的生命周期。
// 它是依赖注入的载体：config、logger、以及后续的 provider/store 都在此统一装配。
package app

import (
	"context"
	"errors"
	"llmgw/internal/api"
	"llmgw/internal/config"
	"llmgw/internal/provider/openai"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// App 是网关应用，持有运行所需的全部依赖。
type App struct {
	cfg    *config.Config
	logger *slog.Logger
	server *http.Server
}

// New 装配应用：注册路由、创建 HTTP Server。
// 关键点：ctx 被设置为 server.BaseContext，使所有请求的 context 都派生自它。
// 这样退出信号触发 ctx 取消时，所有在途请求的 context 会级联取消，
// 从而能通过 provider 把取消传播给上游，避免退出后上游仍在生成（继续扣费）。
func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) *App {
	mux := chi.NewRouter()
	mux.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		// TODO 检查依赖（第 3 章接入 provider 后补充真实健康检查）。
		logger.Info("everything is ok")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	p := openai.New(cfg.OpenAI)
	mux.Post("/v1/chat/completions", api.ChatHandler(p, logger))

	return &App{
		cfg:    cfg,
		logger: logger,
		server: &http.Server{
			Addr:        cfg.Server.Addr,
			Handler:     mux,
			BaseContext: func(l net.Listener) context.Context { return ctx },
		},
	}
}

// Run 启动 HTTP 服务并阻塞，直到 ctx 被取消，然后执行优雅退出。
func (a *App) Run(ctx context.Context) error {
	go func() {
		a.logger.Info("server starting", "addr", a.cfg.Server.Addr)
		// ErrServerClosed 是正常关闭信号，不是错误，需排除。
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("server error", "err", err)
		}
	}()

	// 等待退出信号。
	<-ctx.Done()

	// 优雅关闭：给在途请求一个宽限期（5s）。
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("shutdown error", "err", err)
		// 宽限期耗尽仍未退完，强制关闭剩余连接。
		_ = a.server.Close()
	}
	a.logger.Info("server stopped")
	return nil
}
