package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
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
			Timeout: 12 * time.Minute,
		},
	}
}

type embedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embedResponse struct {
	Embedding []float64 `json:"embedding"`
}

func (c *OllamaClient) Embed(ctx context.Context, text string) ([]float64, error) {
	reqBody := embedRequest{
		Model:  c.Model,
		Prompt: text,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化 embedding 请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建 embedding 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送 embedding 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 embedding 响应失败: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Ollama embedding 返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	var result embedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析 embedding 响应失败: %w\n原始响应: %s", err, string(respBody))
	}

	if len(result.Embedding) == 0 {
		return nil, fmt.Errorf("embedding 为空")
	}

	for i, v := range result.Embedding {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("embedding 包含非法数值，index=%d, value=%v", i, v)
		}
	}

	return result.Embedding, nil
}