// demo02 在 demo01 的基础上加入命令行循环，实现简单的单轮问答。
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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

// ChatRequest 是 Ollama chat 接口的请求体。
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

func main() {
	// bufio.Reader 可以从终端读取包含空格的一整行输入。
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Ollama LLM Demo 已启动。")
	fmt.Println("输入 exit 退出。")

	for {
		fmt.Print("\n你：")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入失败:", err)
			return
		}

		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		if input == "exit" {
			fmt.Println("退出。")
			return
		}

		// 每次用户输入都创建新的 context，保证每次请求都有独立的超时控制。
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)

		answer, err := callOllama(ctx, []Message{
			{
				Role:    "user",
				Content: input,
			},
		})

		cancel()

		if err != nil {
			fmt.Println("调用 Ollama 失败:", err)
			continue
		}

		fmt.Println("\n模型：")
		fmt.Println(answer)
	}
}

// callOllama 负责把用户输入发送给 Ollama，并返回模型的文本回答。
func callOllama(ctx context.Context, messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    modelName,
		Messages: messages,
		Stream:   false,
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
