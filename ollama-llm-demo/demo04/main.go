// demo04 展示多轮对话：把历史 user/assistant 消息一起发给模型。
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

// ChatRequest 是调用 Ollama chat 接口需要的请求结构。
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	// messages 会保存整段对话历史。第一条 system 消息负责设定助手人设。
	messages := []Message{
		{
			Role:    "system",
			Content: "你是一个耐心的 Go 语言老师。请用中文回答，适合初学者。",
		},
	}

	fmt.Println("Ollama 多轮对话 Demo 已启动。")
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

		// 把本轮用户输入追加到历史消息中。
		messages = append(messages, Message{
			Role:    "user",
			Content: input,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)

		answer, err := callOllama(ctx, messages)

		cancel()

		if err != nil {
			fmt.Println("调用 Ollama 失败:", err)
			continue
		}

		fmt.Println("\n模型：")
		fmt.Println(answer)

		// 把模型回答也加入历史，这样下一轮模型才知道之前聊过什么。
		messages = append(messages, Message{
			Role:    "assistant",
			Content: answer,
		})
	}
}

// callOllama 把完整对话历史发送给模型，返回本轮 assistant 的回答。
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
