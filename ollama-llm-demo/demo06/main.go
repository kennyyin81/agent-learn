// demo06 展示 Ollama 的 JSON 格式输出：要求模型只返回可解析的 JSON 字符串。
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

// ChatRequest 的 Format 字段设置为 "json" 时，会提示 Ollama 尽量返回 JSON。
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`

	// format 可以传 "json"，也可以传 JSON Schema。
	// 这里先用 "json"，便于初学。
	Format string `json:"format,omitempty"`
}

type ChatResponse struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// Lesson 是我们希望模型最终生成的数据结构。
// 如果模型返回的 JSON 字段名和 json 标签匹配，就可以直接 Unmarshal 到这里。
type Lesson struct {
	Title      string   `json:"title"`
	Difficulty string   `json:"difficulty"`
	Keywords   []string `json:"keywords"`
}

func main() {
	messages := []Message{
		{
			Role: "system",
			Content: `
				你是一个 JSON 生成器。
				你只能输出合法 JSON。
				不要输出 Markdown。
				不要输出解释文字。
				`,
		},
		{
			Role: "user",
			Content: `
				请根据下面主题生成一个 JSON：

				主题：Go 语言 interface 入门

				JSON 字段要求：
				{
					"title": "string",
					"difficulty": "beginner | intermediate | advanced",
					"keywords": ["string"]
				}
				`,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 21*time.Minute)
	defer cancel()

	text, err := callOllamaJSON(ctx, messages)
	if err != nil {
		fmt.Println("调用 Ollama 失败:", err)
		return
	}

	// 模型返回的是字符串，需要再解析为业务结构体 Lesson。
	var lesson Lesson
	if err := json.Unmarshal([]byte(text), &lesson); err != nil {
		fmt.Println("JSON 解析失败:", err)
		fmt.Println("模型原始输出:")
		fmt.Println(text)
		return
	}

	fmt.Println("标题:", lesson.Title)
	fmt.Println("难度:", lesson.Difficulty)
	fmt.Println("关键词:", lesson.Keywords)
	fmt.Printf("%+v\n", lesson)
}

// callOllamaJSON 调用 Ollama，并通过 format=json 要求模型输出 JSON。
func callOllamaJSON(ctx context.Context, messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    modelName,
		Messages: messages,
		Stream:   false,
		Format:   "json",
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
