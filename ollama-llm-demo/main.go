// demo 展示如何直接调用 Ollama 的 /api/chat 接口，并以流式方式打印模型输出。
// 运行前需要先启动 Ollama 服务，并确保本机已经拉取了下面 modelName 指定的模型。
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
	ollamaURL = "http://127.0.0.1:11434/api/chat"
	modelName = "qwen3.6:27b"
)

type Message struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	Thinking string `json:"thinking,omitempty"`
}

// ChatRequest 是发送给 Ollama 的请求体。
// json 标签决定结构体字段序列化成 JSON 时使用的字段名。
type ChatRequest struct {
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages"`
	Stream   bool                   `json:"stream"`
	Think    bool                   `json:"think"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type StreamChatResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Done    bool    `json:"done"`
	Error   string  `json:"error,omitempty"`
}

func main() {
	// context.WithTimeout 用来限制本次请求最长等待时间，避免程序一直卡住。
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	err := streamOllama(ctx, []Message{
		{
			Role:    "user",
			Content: "hello, reply with one short sentence.",
		},
	})

	if err != nil {
		fmt.Println("\n调用 Ollama 失败:", err)
		return
	}

	fmt.Println("\n完成。")
}

// streamOllama 发送一次流式聊天请求，并把模型返回的内容边接收边打印。
func streamOllama(ctx context.Context, messages []Message) error {
	reqBody := ChatRequest{
		Model:    modelName,
		Messages: messages,
		Stream:   true,

		// 关键：关闭 thinking，避免 content 为空
		Think: false,

		Options: map[string]interface{}{
			"num_predict": 128,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		ollamaURL,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	fmt.Println("正在请求 Ollama，下面会开始流式输出：")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Ollama 返回错误状态码: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)

	hasOutput := false

	for scanner.Scan() {
		line := scanner.Bytes()

		var chunk StreamChatResponse
		if err := json.Unmarshal(line, &chunk); err != nil {
			return fmt.Errorf("解析流式响应失败: %w\n原始内容: %s", err, string(line))
		}

		if chunk.Error != "" {
			return fmt.Errorf("Ollama 返回错误: %s", chunk.Error)
		}

		if chunk.Message.Content != "" {
			hasOutput = true
			fmt.Print(chunk.Message.Content)
		}

		// 调试用：如果还有 thinking 字段，说明模型在输出思考内容
		if chunk.Message.Thinking != "" {
			hasOutput = true
			fmt.Print(chunk.Message.Thinking)
		}

		if chunk.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取流失败: %w", err)
	}

	if !hasOutput {
		fmt.Println("\n没有收到 content/thinking 文本。建议检查模型是否支持 chat 接口或换一个小模型测试。")
	}

	return nil
}
