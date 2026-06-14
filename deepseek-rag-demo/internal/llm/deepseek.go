package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type DeepSeekClient struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func NewDeepSeekClient(baseURL string, apiKey string, model string) *DeepSeekClient {
	return &DeepSeekClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Minute,
		},
	}
}

type chatRequest struct {
	Model       string              `json:"model"`
	Messages    []map[string]string `json:"messages"`
	Temperature float64             `json:"temperature"`
	Stream      bool                `json:"stream"`
	Thinking    map[string]string   `json:"thinking,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func (c *DeepSeekClient) Chat(ctx context.Context, messages []map[string]string) (string, error) {
	if strings.TrimSpace(c.APIKey) == "" {
		return "", fmt.Errorf("缺少 DEEPSEEK_API_KEY 环境变量")
	}

	reqBody := chatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: 0.2,
		Stream:      false,
		Thinking: map[string]string{
			"type": "disabled",
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化 DeepSeek 请求失败: %w", err)
	}

	url := c.BaseURL + "/chat/completions"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("创建 DeepSeek 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送 DeepSeek 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 DeepSeek 响应失败: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("DeepSeek 返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	var result chatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析 DeepSeek 响应失败: %w\n原始响应: %s", err, string(respBody))
	}

	if result.Error != nil {
		return "", fmt.Errorf("DeepSeek API 错误: type=%s code=%s message=%s",
			result.Error.Type,
			result.Error.Code,
			result.Error.Message,
		)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("DeepSeek 没有返回 choices")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("DeepSeek 返回内容为空")
	}

	return content, nil
}