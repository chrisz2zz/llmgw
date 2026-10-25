// Package openai 实现 OpenAI 兼容协议的 Provider 适配器。
// 它负责把统一请求翻译成供应商格式、发送请求，并把供应商响应翻译回统一模型；
// 供应商差异（错误码、字段命名等）都在此消化，上层只面对 protocol 统一模型。
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"llmgw/internal/config"
	"llmgw/internal/protocol"
	"net/http"
	"time"
)

// OpenAIProvider 是 OpenAI 兼容供应商（OpenAI / DeepSeek 等）的适配器实现。
type OpenAIProvider struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New 根据配置构造 OpenAIProvider。
// httpClient 在 Provider 生命周期内复用，以复用底层连接池。
func New(cfg config.OpenAIConfig) *OpenAIProvider {
	timeout, _ := time.ParseDuration(cfg.Timeout)
	return &OpenAIProvider{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Name 返回供应商名称，供第 5 章的注册表标识使用。
func (o *OpenAIProvider) Name() string {
	return "openai"
}

// Chat 发送非流式对话请求，返回统一响应。
// 流程：构造请求 → 发送 → 按状态码分支：非 2xx 归一错误码，2xx 解析统一响应。
func (o *OpenAIProvider) Chat(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	url := o.baseURL + "/chat/completions"
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Authorization", "Bearer "+o.apiKey)
	r.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 非 2xx：解析供应商错误体并归一为统一错误码。
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var ae apiError
		if json.Unmarshal(data, &ae) == nil && ae.Error.Code != "" {
			return nil, &protocol.Error{Code: mapErrCode(ae.Error.Code), Message: ae.Error.Message, Type: ae.Error.Type}
		}
		return nil, &protocol.Error{
			Code:    protocol.ErrUpstreamError,
			Message: resp.Status,
			Type:    "upstream_error",
		}
	}

	response := &protocol.Response{}
	err = json.Unmarshal(data, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Stream 流式响应，第 4 章实现。
func (o *OpenAIProvider) Stream(ctx context.Context, req *protocol.Request) (any, error) {
	return nil, nil
}

// Models 查询模型列表，后续章节实现。
func (o *OpenAIProvider) Models(ctx context.Context) (any, error) {
	return nil, nil
}

// apiError 是 OpenAI 兼容供应商的错误响应体。
type apiError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// mapErrCode 按供应商的 code 字符串归一为统一错误码。
// TODO 当前各供应商（如 DeepSeek）的 code 命名与 OpenAI 不完全一致，
// 后续需按具体供应商细化映射规则（可结合 HTTP 状态码共同判断）。
func mapErrCode(code string) protocol.ErrCode {
	switch code {
	case "invalid_api_key":
		return protocol.ErrInvalidAPIKey
	case "rate_limit_exceeded", "insufficient_quota":
		return protocol.ErrRateLimit
	case "model_not_found":
		return protocol.ErrModelNotFound
	case "context_length_exceeded":
		return protocol.ErrContextLengthExceeded
	default:
		return protocol.ErrUpstreamError
	}
}
