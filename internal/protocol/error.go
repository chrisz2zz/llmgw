package protocol

import (
	"fmt"
)

// ErrCode 是网关统一的业务错误码（机器可读），与 HTTP 状态码分离。
type ErrCode string

// 统一错误码枚举。各层将供应商的差异错误归一为这些码，下游据此稳定处理。
const (
	ErrInvalidRequest        ErrCode = "invalid_request"         // 请求参数错误
	ErrInvalidAPIKey         ErrCode = "invalid_api_key"         // 鉴权失败
	ErrRateLimit             ErrCode = "rate_limit"              // 限流
	ErrModelNotFound         ErrCode = "model_not_found"         // 模型不存在
	ErrContextLengthExceeded ErrCode = "context_length_exceeded" // 上下文超长
	ErrUpstreamError         ErrCode = "upstream_error"          // 上游供应商错误
)

// HTTPStatus 返回错误码对应的 HTTP 状态码。
// 注意：HTTP 状态码属于传输层语义，严格来说应由接入层（api）维护；
// 这里为 MVP 便利暂放在 protocol 包内。
func (c ErrCode) HTTPStatus() int {
	switch c {
	case ErrInvalidRequest:
		return 400
	case ErrInvalidAPIKey:
		return 401
	case ErrRateLimit:
		return 429
	case ErrModelNotFound:
		return 404
	case ErrContextLengthExceeded:
		return 400
	case ErrUpstreamError:
		return 502
	default:
		return 500
	}
}

// Error 是统一的错误响应体，各层返回错误时统一使用。
type Error struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
	Type    string  `json:"type"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
