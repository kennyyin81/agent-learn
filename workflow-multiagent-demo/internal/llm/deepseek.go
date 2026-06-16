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

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func NewClient(baseURL string, apiKey string, model string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []Message      `json:"messages"`
	Temperature    float64        `json:"temperature"`
	Stream         bool           `json:"stream"`
	ResponseFormat map[string]any `json:"response_format,omitempty"`  // 开启json模式
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

func (c *Client) ChatText(ctx context.Context, messages []Message, temperature float64) (string, error) {
	return c.chat(ctx, messages, temperature, false)
}

func (c *Client) ChatJSON(ctx context.Context, messages []Message, temperature float64) (string, error) {
	return c.chat(ctx, messages, temperature, true)
}

func (c *Client) chat(ctx context.Context, messages []Message, temperature float64, jsonMode bool) (string, error) {
	if strings.TrimSpace(c.APIKey) == "" {
		return "", fmt.Errorf("缺少 DEEPSEEK_API_KEY")
	}

	if strings.TrimSpace(c.BaseURL) == "" {
		return "", fmt.Errorf("缺少 DEEPSEEK_BASE_URL，应该是 https://api.deepseek.com")
	}

	if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
		return "", fmt.Errorf("DEEPSEEK_BASE_URL 配置错误：%q，应该是 https://api.deepseek.com", c.BaseURL)
	}

	reqBody := chatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		Stream:      false,
	}

	if jsonMode {
		reqBody.ResponseFormat = map[string]any{
			"type": "json_object",
		}
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