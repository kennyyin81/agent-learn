package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"trpc.group/trpc-go/trpc-agent-go/agent/llmagent"
	"trpc.group/trpc-go/trpc-agent-go/event"
	"trpc.group/trpc-go/trpc-agent-go/model"
	"trpc.group/trpc-go/trpc-agent-go/model/openai"
	"trpc.group/trpc-go/trpc-agent-go/runner"
	sessioninmemory "trpc.group/trpc-go/trpc-agent-go/session/inmemory"
	"trpc.group/trpc-go/trpc-agent-go/tool"
	"trpc.group/trpc-go/trpc-agent-go/tool/function"
)

func main() {
	if strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) == "" {
		log.Fatal("请先设置 OPENAI_API_KEY")
	}

	if strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")) == "" {
		fmt.Println("提示：如果你使用 DeepSeek，建议设置 OPENAI_BASE_URL=https://api.deepseek.com")
	}

	ctx := context.Background()

	modelName := env("MODEL", "deepseek-v4-flash")
	variant := env("VARIANT", "deepseek")

	modelInstance := openai.New(
		modelName,
		openai.WithVariant(openai.Variant(variant)),
	)

	calculatorTool := function.NewFunctionTool(
		calculator,
		function.WithName("calculator"),
		function.WithDescription("Perform arithmetic calculations. Use for add, subtract, multiply, divide and power."),
	)

	timeTool := function.NewFunctionTool(
		currentTime,
		function.WithName("current_time"),
		function.WithDescription("Get current date and time."),
	)

	genConfig := model.GenerationConfig{
		Stream:      true,
		Temperature: floatPtr(0.2),
		MaxTokens:   intPtr(2000),  // 限制模型本次最多输出多长
	}
	// Description 说明 Agent 的用途，Instruction 约束 Agent 的行为
	agent := llmagent.New(
		"study-assistant",
		llmagent.WithModel(modelInstance),
		llmagent.WithDescription("A study assistant with calculator and time tools."),
		llmagent.WithInstruction("你是一个中文 Agent 学习助手。遇到计算或时间问题时要调用工具，不要自己心算。"),
		llmagent.WithGenerationConfig(genConfig),
		llmagent.WithTools([]tool.Tool{
			calculatorTool,
			timeTool,
		}),
	)

	sessionService := sessioninmemory.NewSessionService()
	// 运行调度器 会话管理 + Agent 调度 + 事件流输出 + 生命周期管理
	r := runner.NewRunner(
		"trpc-agent-study-mini",
		agent,
		runner.WithSessionService(sessionService),
	)
	defer r.Close()

	userID := "user-001"
	sessionID := "session-001"

	userMessage := model.NewUserMessage("请查询当前时间，然后计算 15 * 23 + 100")
	// 返回的是未来会陆续产出事件的通道
	events, err := r.Run(ctx, userID, sessionID, userMessage)
	if err != nil {
		log.Fatal(err)
	}

	handleEvents(events)
}

func handleEvents(events <-chan *event.Event) {
	fmt.Println("Assistant:")

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
			fmt.Println("\n[Tool Calls]")
			for _, tc := range choice.Message.ToolCalls {
				fmt.Printf("- name=%s id=%s args=%s\n",
					tc.Function.Name,
					tc.ID,
					string(tc.Function.Arguments),
				)
			}
			continue
		}
		hasToolResponse := false
		for _, ch := range evt.Response.Choices {
			if ch.Message.Role == model.RoleTool && ch.Message.ToolID != "" {
				fmt.Printf("\n[Tool Result] id=%s result=%s\n",
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
			continue
		}

		if choice.Message.Content != "" && !evt.IsFinalResponse(){
			fmt.Print("\n" + choice.Message.Content)
		}

		if evt.Done || evt.IsFinalResponse() {
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
