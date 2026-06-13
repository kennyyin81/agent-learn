// demo03 展示 system message 的作用：通过系统提示词约束模型的回答风格。
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
	ollamaURL = "http://localhost:11434/api/chat"
	modelName = "qwen3.6:27b"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 是发送给 Ollama 的 JSON 请求体。
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

func main() {
	// messages 按顺序组成上下文：先给系统规则，再给用户问题。
	messages := []Message{
		{
			Role:    "system",
			Content: "你是一个耐心的 Go 语言老师。请用中文回答，先讲概念，再给代码，适合初学者。",
		},
		{
			Role:    "user",
			Content: "Go 语言中的 interface 是什么？",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	answer, err := callOllama(ctx, messages)
	if err != nil {
		fmt.Println("调用 Ollama 失败:", err)
		return
	}

	fmt.Println(answer)
}

// callOllama 发送一次非流式请求，并返回 message.content。
func callOllama(ctx context.Context, messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    modelName,
		Messages: messages,
		Stream:   false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		ollamaURL,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Ollama 返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", err
	}

	return chatResp.Message.Content, nil
}
