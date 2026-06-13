// demo05 展示流式输出：模型生成一点，程序就打印一点，体验更像聊天。
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// ChatRequest 中 Stream 为 true 时，Ollama 会分块返回结果。
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type StreamChatResponse struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

func main() {
	messages := []Message{
		{
			Role:    "system",
			Content: "你是一个耐心的 Go 语言老师。请用中文回答。",
		},
		{
			Role:    "user",
			Content: "请写一段 300 字左右的 Go 语言学习建议。",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := streamOllama(ctx, messages); err != nil {
		fmt.Println("\n流式调用失败:", err)
		return
	}

	fmt.Println()
}

// streamOllama 发送流式请求，并逐行读取 Ollama 返回的 JSON 片段。
func streamOllama(ctx context.Context, messages []Message) error {
	reqBody := ChatRequest{
		Model:    modelName,
		Messages: messages,
		Stream:   true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		ollamaURL,
		bytes.NewReader(bodyBytes), // 把 []byte 包装成一个 Reader
	)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Ollama 返回错误状态码: %d", resp.StatusCode)
	}
	// 从 resp.Body 里持续读取数据的扫描器
	// 每次 Scan 读取一行；Ollama 的流式接口通常每行都是一个 JSON 对象。
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Bytes()

		// 将当前 JSON 片段解析成结构体，取出本次新增的 content。
		var chunk StreamChatResponse
		if err := json.Unmarshal(line, &chunk); err != nil {
			return fmt.Errorf("解析流式 JSON 失败: %w\n原始内容: %s", err, string(line))
		}

		fmt.Print(chunk.Message.Content)

		if chunk.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取流失败: %w", err)
	}

	return nil
}
