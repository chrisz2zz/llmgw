// Package api 是接入层：负责 HTTP 请求的解析、调用 provider、写出响应。
package api

import (
	"encoding/json"
	"llmgw/internal/protocol"
	"llmgw/internal/provider"
	"log/slog"
	"net/http"
)

// ChatHandler 返回 /v1/chat/completions 的处理器。
// 职责：解析请求 → 调用 provider → 写出统一响应或错误。
func ChatHandler(p provider.Provider, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("handle chat", "provider", p.Name())

		defer r.Body.Close()

		var req protocol.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("parse request body", "err", err)

			writeError(w, &protocol.Error{
				Code:    protocol.ErrInvalidRequest,
				Message: "invalid request body",
				Type:    "invalid_request_error",
			})
			return
		}

		resp, err := p.Chat(r.Context(), &req)
		if err != nil {
			logger.Error("provider call", "err", err)

			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}
