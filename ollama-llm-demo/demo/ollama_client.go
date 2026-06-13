// 本文件把访问 Ollama 的 HTTP 细节封装成 OllamaClient。
// 初学时可以先看 main.go 理解怎么使用，再回到这里看内部实现。
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaClient struct {
	// BaseURL 是 Ollama 服务地址，默认本机 11434 端口。
	BaseURL string
	// Model 是每次请求使用的模型名称。
	Model string
	// Client 是标准库 HTTP 客户端，保留为字段方便以后自定义超时、代理等配置。
	Client *http.Client
}

// Message 表示 Ollama chat 接口中的一条消息。
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 是发送给 Ollama /api/chat 的请求体。
type ChatRequest struct {
	Model    string      `json:"model"`
	Messages []Message   `json:"messages"`
	Stream   bool        `json:"stream"`
	Format   interface{} `json:"format,omitempty"`
}

type ChatResponse struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// NewOllamaClient 创建一个使用指定模型的 Ollama 客户端。
func NewOllamaClient(model string) *OllamaClient {
	return &OllamaClient{
		BaseURL: "http://localhost:11434",
		Model:   model,
		Client:  http.DefaultClient,
	}
}

// Chat 发起普通文本聊天，返回模型回答的 content。
func (c *OllamaClient) Chat(ctx context.Context, messages []Message) (string, error) {
	return c.chat(ctx, messages, false, nil)
}

// ChatJSON 要求模型返回 JSON 字符串。
func (c *OllamaClient) ChatJSON(ctx context.Context, messages []Message) (string, error) {
	return c.chat(ctx, messages, false, "json")
}

// ChatWithSchema 使用 JSON Schema 约束模型输出结构。
func (c *OllamaClient) ChatWithSchema(ctx context.Context, messages []Message, schema interface{}) (string, error) {
	return c.chat(ctx, messages, false, schema)
}

// chat 是内部通用方法，三个公开方法最终都会走到这里。
func (c *OllamaClient) chat(ctx context.Context, messages []Message, stream bool, format interface{}) (string, error) {
	reqBody := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   stream,
		Format:   format,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	url := c.BaseURL + "/api/chat"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Ollama 返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w\n原始响应: %s", err, string(respBody))
	}

	return chatResp.Message.Content, nil
}
