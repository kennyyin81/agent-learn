// demo07 展示使用 JSON Schema 约束模型输出结构，比单纯 format=json 更严格。
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

// ChatRequest 的 Format 使用 interface{}，这样既能传字符串 "json"，也能传 JSON Schema。
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

// Lesson 是本示例希望从模型输出中得到的 Go 结构体。
type Lesson struct {
	Title      string   `json:"title"`
	Difficulty string   `json:"difficulty"`
	Keywords   []string `json:"keywords"`
}

func main() {
	// schema 描述模型必须返回的 JSON 形状：字段类型、枚举值和必填字段。
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type": "string",
			},
			"difficulty": map[string]interface{}{
				"type": "string",
				"enum": []string{
					"beginner",
					"intermediate",
					"advanced",
				},
			},
			"keywords": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{
			"title",
			"difficulty",
			"keywords",
		},
	}

	messages := []Message{
		{
			Role:    "system",
			Content: "你是一个严格的 JSON 生成器。不要输出 Markdown，不要输出解释文字。",
		},
		{
			Role:    "user",
			Content: "请生成一个 Go 语言 interface 入门课程信息。",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 21*time.Minute)
	defer cancel()

	text, err := callOllamaWithSchema(ctx, messages, schema)
	if err != nil {
		fmt.Println("调用 Ollama 失败:", err)
		return
	}

	// 即使使用了 Schema，实际业务里仍然应该解析并检查返回值。
	var lesson Lesson
	if err := json.Unmarshal([]byte(text), &lesson); err != nil {
		fmt.Println("JSON 解析失败:", err)
		fmt.Println("模型原始输出:")
		fmt.Println(text)
		return
	}

	fmt.Printf("%+v\n", lesson)
}

// callOllamaWithSchema 把 JSON Schema 放进 format 字段，请求模型按该结构输出。
func callOllamaWithSchema(ctx context.Context, messages []Message, schema interface{}) (string, error) {
	reqBody := ChatRequest{
		Model:    modelName,
		Messages: messages,
		Stream:   false,
		Format:   schema,
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
