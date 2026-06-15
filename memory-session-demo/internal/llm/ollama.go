package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OllamaClient struct {
	URL        string
	Model      string
	HTTPClient *http.Client
}

func NewOllamaClient(url string, model string) *OllamaClient {
	return &OllamaClient{
		URL:   url,
		Model: model,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Minute,
		},
	}
}

func (c *OllamaClient) Chat(ctx context.Context, messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   false,
		Options: map[string]any{
			"temperature": 0.3,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.URL,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
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

	if chatResp.Message.Content == "" {
		return "", fmt.Errorf("模型返回内容为空")
	}

	return chatResp.Message.Content, nil
}