// demo01 展示最基础的一次性聊天调用：构造请求、发送 HTTP POST、解析响应。
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// Ollama 默认监听 11434 端口，/api/chat 是对话接口。
	ollamaURL = "http://localhost:11434/api/chat"

	// 如果你实际使用的是 Ollama 官方 Qwen3.6 27B 模型，改成：
	modelName = "qwen3.6:27b"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 是请求 Ollama 时发送的 JSON 结构。
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

func main() {
	// 大模型响应可能比较慢，所以这里设置较长的超时时间。
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	answer, err := callOllama(ctx, []Message{
		{
			Role:    "user",
			Content: "请用中文解释 Go 语言中的 interface，适合初学者。",
		},
	})
	if err != nil {
		fmt.Println("调用 Ollama 失败:", err)
		return
	}

	fmt.Println(answer)
}

// callOllama 封装一次完整的非流式 Ollama 调用。
func callOllama(ctx context.Context, messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    modelName,
		Messages: messages,
		Stream:   false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		ollamaURL,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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
