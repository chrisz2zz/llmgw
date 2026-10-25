package api

import (
	"encoding/json"
	"errors"
	"llmgw/internal/protocol"
	"net/http"
)

// writeJSON 以 JSON 格式写出响应，统一设置 Content-Type。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError 是所有错误响应的统一出口：
// 通过 errors.As 提取 *protocol.Error，用其业务错误码决定 HTTP 状态码；
// 非业务错误（如内部错误）统一按 500 处理。
func writeError(w http.ResponseWriter, err error) {
	var perr *protocol.Error
	if errors.As(err, &perr) {
		writeJSON(w, perr.Code.HTTPStatus(), perr)
		return
	}

	writeJSON(w, http.StatusInternalServerError, &protocol.Error{
		Code:    protocol.ErrUpstreamError,
		Message: "internal error",
		Type:    "internal_error",
	})
}
