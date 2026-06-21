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

	"trpc.group/trpc-go/trpc-agent-go/agent/llmagent"
	"trpc.group/trpc-go/trpc-agent-go/event"
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/model/openai"
	"trpc.group/trpc-go/trpc-agent-go/runner"
	"trpc.group/trpc-go/trpc-agent-go/tool"
	"trpc.group/trpc-go/trpc-agent-go/tool/mcp"
)

func main() {
	modelName := flag.String("model", env("MODEL", "deepseek-v4-flash"), "model name")
	flag.Parse()

	if strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) == "" {
		log.Fatal("请先设置 OPENAI_API_KEY")
	}

	if strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")) == "" {
		fmt.Println("提示：使用 DeepSeek 时请设置 OPENAI_BASE_URL=https://api.deepseek.com")
	}

	ctx := context.Background()

	modelInstance := openai.New(*modelName)

	stdioToolSet := mcp.NewMCPToolSet(
		mcp.ConnectionConfig{
			Transport: "stdio",
			Command:   "go",
			Args:      []string{"run", "./mcpserver/main.go"},
			Timeout:   10 * time.Second,
		},
		mcp.WithName("study-stdio-mcp"),
		mcp.WithToolFilterFunc(tool.NewIncludeToolNamesFilter("echo", "add", "read_note")),
	)

	if err := stdioToolSet.Init(ctx); err != nil {
		log.Fatalf("初始化 MCP ToolSet 失败: %v", err)
	}
	defer stdioToolSet.Close()

	fmt.Println("MCP ToolSet 初始化成功。")

	genConfig := model.GenerationConfig{
		MaxTokens:   intPtr(2000),
		Temperature: floatPtr(0.2),
		Stream:      true,
	}

	agent := llmagent.New(
		"mcp-study-agent",
		llmagent.WithModel(modelInstance),
		llmagent.WithDescription("A Chinese study assistant with MCP tools."),
		llmagent.WithInstruction(`
			你是一个中文 Agent 学习助手。

			你可以使用 MCP 工具：
			1. echo：回显文本。
			2. add：做两个数字相加。
			3. read_note：读取内置学习笔记，topic 支持 mcp、rag、agent。

			规则：
			- 用户问 MCP、RAG、Agent 基础概念时，优先调用 read_note。
			- 用户要求相加时，调用 add。
			- 工具返回结果后，用中文解释给用户。
			- 不要声称自己调用了没有暴露的工具。
			`),
		llmagent.WithGenerationConfig(genConfig),
		llmagent.WithToolSets([]tool.ToolSet{
			stdioToolSet,
		}),
	)

	r := runner.NewRunner(
		"mcp-trpc-agent-demo",
		agent,
	)
	defer r.Close()

	userID := "user-001"
	sessionID := fmt.Sprintf("mcp-session-%d", time.Now().Unix())

	fmt.Println("MCP + trpc-agent-go Demo 已启动")
	fmt.Println("模型:", *modelName)
	fmt.Println("Session:", sessionID)
	fmt.Println("输入 exit 退出")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("你> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("已退出。")
			return
		}

		msg := model.NewUserMessage(input)

		events, err := r.Run(ctx, userID, sessionID, msg)
		if err != nil {
			fmt.Println("运行失败:", err)
			continue
		}

		handleEvents(events)
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func handleEvents(events <-chan *event.Event) {
	fmt.Print("助手> ")

	printedStreaming := false

	for evt := range events {
		if evt.Error != nil {
			fmt.Printf("\n[Error] %s\n", evt.Error.Message)
			continue
		}

		if evt.Response == nil || len(evt.Response.Choices) == 0 {
			continue
		}

		choice := evt.Response.Choices[0]

		if len(choice.Message.ToolCalls) > 0 {
			fmt.Println()
			fmt.Println("[MCP Tool Calls]")
			for _, tc := range choice.Message.ToolCalls {
				fmt.Printf("- name=%s id=%s args=%s\n",
					tc.Function.Name,
					tc.ID,
					string(tc.Function.Arguments),
				)
			}
			fmt.Println("[Executing MCP tools...]")
			continue
		}

		hasToolResponse := false
		for _, ch := range evt.Response.Choices {
			if ch.Message.Role == model.RoleTool && ch.Message.ToolID != "" {
				fmt.Printf("\n[MCP Tool Result] id=%s result=%s\n",
					ch.Message.ToolID,
					strings.TrimSpace(ch.Message.Content),
				)
				hasToolResponse = true
			}
		}
		if hasToolResponse {
			continue
		}

		if choice.Delta.Content != "" {
			fmt.Print(choice.Delta.Content)
			printedStreaming = true
			continue
		}

		if choice.Message.Content != "" && !printedStreaming {
			fmt.Print(choice.Message.Content)
		}

		if evt.IsFinalResponse() {
			fmt.Println()
			break
		}
	}
}

func intPtr(v int) *int {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}

func env(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
