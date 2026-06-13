package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	URL        string
	Model      string
	HTTPClient *http.Client
}

func NewClient(url string, model string) *Client {
	return &Client{
		URL:        url,
		Model:      model,
		HTTPClient: http.DefaultClient,
	}
}

func (c *Client) Chat(ctx context.Context, messages []Message, tools []Tool) (ChatResponse, error) {
	reqBody := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
		Options: map[string]any{
			"temperature": 0,
		},
	}

	respBody, err := c.doRequest(ctx, reqBody)
	if err != nil {
		return ChatResponse{}, err
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return ChatResponse{}, fmt.Errorf("解析响应失败: %w\n原始响应: %s", err, string(respBody))
	}

	return chatResp, nil
}

func (c *Client) ChatStream(
	ctx context.Context,
	messages []Message,
	tools []Tool,
	onContentDelta func(delta string),
) (Message, error) {
	reqBody := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Tools:    tools,
		Stream:   true,
		Options: map[string]any{
			"temperature": 0,
		},
	}
	// Go struct 转 json 字节流
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return Message{}, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.URL,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return Message{}, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return Message{}, fmt.Errorf("Ollama 返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}
	// decoder可以解析流式响应
	decoder := json.NewDecoder(resp.Body)
	// 最终回答
	assembled := Message{
		Role: "assistant",
	}

	for {
		var chunk ChatResponse
		err := decoder.Decode(&chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			return Message{}, fmt.Errorf("解析流式 chunk 失败: %w", err)
		}

		if chunk.Message.Role != "" {
			assembled.Role = chunk.Message.Role
		}

		if chunk.Message.Thinking != "" {
			assembled.Thinking += chunk.Message.Thinking
		}

		if chunk.Message.Content != "" {
			assembled.Content += chunk.Message.Content
			if onContentDelta != nil {  // 实时把增量内容传给回调函数
				onContentDelta(chunk.Message.Content)
			}
		}

		if len(chunk.Message.ToolCalls) > 0 {
			assembled.ToolCalls = append(assembled.ToolCalls, chunk.Message.ToolCalls...)
		}

		if chunk.Done {
			break
		}
	}

	return assembled, nil
}

func (c *Client) doRequest(ctx context.Context, reqBody ChatRequest) ([]byte, error) {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.URL,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Ollama 返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}