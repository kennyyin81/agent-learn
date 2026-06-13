package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"ollama-tool-demo/internal/agent"
	"ollama-tool-demo/internal/llm"
	"ollama-tool-demo/internal/localtools"
)

const (
	ollamaURL = "http://localhost:11434/api/chat"
	modelName = "qwen3.6:27b"
)

func main() {
	stream := flag.Bool("stream", false, "是否开启流式输出")
	flag.Parse()

	client := llm.NewClient(ollamaURL, modelName)

	registry := localtools.NewRegistry()
	tools := registry.Schemas()

	reader := bufio.NewReader(os.Stdin)

	messages := []llm.Message{
		{
			Role: "system",
			Content: `
				你是一个可以使用工具的中文助手。

				你可以使用这些工具：
				1. get_current_time：查询当前时间。
				2. calculator：做精确数学计算。
				3. random_number：生成随机整数。
				4. read_text_file：读取 workspace 目录下的文本文件。
				5. write_text_file：写入 workspace 目录下的文本文件。

				规则：
				1. 如果用户问当前时间，必须调用 get_current_time。
				2. 如果用户要求数学计算，必须调用 calculator，不要自己心算。
				3. 如果用户要求随机数，必须调用 random_number。
				4. 如果用户要求读取文件，必须调用 read_text_file。
				5. 如果用户要求写文件，必须调用 write_text_file。
				6. 复杂计算要拆成多步，每一步都调用 calculator。
				7. 工具返回结果后，再用中文给用户最终回答。
				8. 如果工具返回 error，要如实告诉用户失败原因。
				9. 你处在 Agent Loop 中，可以多轮调用工具。
				`,
		},
	}

	fmt.Println("本地 Ollama Tool Agent 已启动。")
	fmt.Println("输入 exit 或 quit 退出。")
	fmt.Println("示例：")
	fmt.Println("1. 帮我生成一个 1 到 100 的随机整数")
	fmt.Println("2. 读取 note.txt 的内容并总结")
	fmt.Println("3. 帮我写一个 hello.txt，内容是：你好，这是 Agent 写入的文件")
	fmt.Println("4. 查询新加坡当前时间，再计算 15 * 23 + 100")
	fmt.Println()

	for {
		fmt.Print("\n你> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("已退出。")
			return
		}

		messages = append(messages, llm.Message{
			Role:    "user",
			Content: input,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)

		if *stream {
			fmt.Print("助手> ")
		}

		answer, updatedMessages, err := agent.Run(
			ctx,
			client,
			registry,
			messages,
			tools,
			agent.Options{
				MaxSteps: 8,
				Stream:   *stream,
				Confirm:  agent.NewStdinConfirm(reader),
			},
		)

		cancel()

		if err != nil {
			fmt.Println()
			fmt.Println("运行失败:", err)
			continue
		}

		messages = updatedMessages

		if !*stream {
			fmt.Println("\n助手>", answer)
		} else {
			fmt.Println()
		}
	}
}