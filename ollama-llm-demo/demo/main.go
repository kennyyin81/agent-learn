// demo 展示把 Ollama 调用封装成客户端之后，main 函数如何更简洁地使用它。
// 请在本目录运行 go run .，这样 main.go 和 ollama_client.go 会一起参与编译。
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// NewOllamaClient 返回一个可复用的客户端，避免在 main 中重复写 HTTP 细节。
	client := NewOllamaClient("qwen3.6:27b")

	// 给本次请求设置超时，模型长时间无响应时程序可以自动结束。
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	// system 消息负责设定助手行为，user 消息是本轮用户问题。
	answer, err := client.Chat(ctx, []Message{
		{
			Role:    "system",
			Content: "你是一个 Go 语言老师，请用中文回答。",
		},
		{
			Role:    "user",
			Content: "什么是 goroutine？用50个字以内回答。",
		},
	})
	if err != nil {
		fmt.Println("调用失败:", err)
		return
	}

	fmt.Println(answer)
}
