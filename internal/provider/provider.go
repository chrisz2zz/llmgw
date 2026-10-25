package provider

import (
	"context"
	"llmgw/internal/protocol"
)

// Provider 抽象所有 LLM 供应商的对话能力。
// 各供应商（OpenAI/Anthropic/本地模型）实现此接口，
// 把供应商差异消化在各自的适配器内部，上层只面对统一模型。
type Provider interface {
	Name() string

	// Chat 发送非流式对话请求，返回统一响应
	// 出错时返回的error应能通过errors.As提取为*protocol.Error
	// 以便上层拿到业务错误码决定HTTP状态码
	Chat(ctx context.Context, req *protocol.Request) (*protocol.Response, error)

	// TODO 流式
	Stream(ctx context.Context, req *protocol.Request) (any, error)

	// TODO 模型列表
	Models(ctx context.Context) (any, error)
}
