// Package protocol 定义网关内部统一的请求/响应协议模型。
// 它是各层之间传递的数据结构，字段对齐 OpenAI Chat Completions 协议，
// 作为防腐层隔离不同供应商的协议差异（供应商差异在 provider 适配层消化）。
package protocol

// Request 是统一的对话请求。
type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`

	// Temperature 和 MaxTokens 使用指针以区分「未传」与「传 0」，
	// 避免把零值误当作真实参数转发给上游。
	Temperature *float64 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`

	Stream bool `json:"stream"`
	Tools  any  `json:"tools,omitempty"` // 工具调用定义，第 13 章 Agent 使用
}

// Message 是一条对话消息。
type Message struct {
	Role    string `json:"role"` // system / user / assistant / tool
	Content string `json:"content"`
}

// Response 是统一的对话响应。
type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`  // 固定为 "chat.completion"
	Created int64    `json:"created"` // unix 秒时间戳
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Usage 是 token 用量，第 9 章成本计量依赖它。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice 是模型返回的一个候选回答。
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"` // stop / length / tool_calls 等
}
