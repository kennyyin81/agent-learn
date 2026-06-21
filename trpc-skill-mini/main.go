package main

import (
	"bufio"
	"context"
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
	skillrepo "trpc.group/trpc-go/trpc-agent-go/skill"
	"trpc.group/trpc-go/trpc-agent-go/tool"
	skilltool "trpc.group/trpc-go/trpc-agent-go/tool/skill"
)

func main() {
	if strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) == "" {
		log.Fatal("请先设置 OPENAI_API_KEY")
	}

	if strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")) == "" {
		fmt.Println("提示：使用 DeepSeek 时建议设置 OPENAI_BASE_URL=https://api.deepseek.com")
	}

	ctx := context.Background()

	modelName := env("MODEL", "deepseek-v4-flash")

	modelInstance := openai.New(
		modelName,
		openai.WithVariant(openai.VariantDeepSeek),
	)

	repo, err := skillrepo.NewFSRepository("./skills")
	if err != nil {
		log.Fatalf("创建 skill repository 失败: %v", err)
	}

	loadSkillTool := skilltool.NewLoadTool(repo)

	genConfig := model.GenerationConfig{
		Stream:      true,
		Temperature: floatPtr(0.2),
		MaxTokens:   intPtr(3000),
	}

	agent := llmagent.New(
		"skill-study-agent",
		llmagent.WithModel(modelInstance),
		llmagent.WithDescription("一个用于学习 Go Agent Skills 的中文助手。"),
		llmagent.WithInstruction(`
			你是一个中文 Go Agent 学习助手。

			你可以使用 Agent Skills。
			当用户的问题涉及 Go Agent 学习路线、Tool Calling、Agent Loop、RAG、Memory、MCP、trpc-agent-go 或 Skill 时，
			优先使用 skill_load 加载相关 Skill，再根据 Skill 指南回答。

			重要规则：
			1. 不要声称自己加载了不存在的 Skill。
			2. 加载 Skill 后，要严格按照 SKILL.md 的说明组织回答。
			3. 如果 Skill 中没有覆盖用户问题，要如实说明，并用普通知识补充。
			4. 回答要用中文，适合初学者逐步学习。
			`),
		llmagent.WithGenerationConfig(genConfig),
		llmagent.WithTools([]tool.Tool{
			loadSkillTool,
		}),
	)

	r := runner.NewRunner("trpc-skill-mini", agent)
	defer r.Close()

	userID := "user-001"
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())

	fmt.Println("trpc-agent-go Skill Mini 已启动")
	fmt.Println("模型:", modelName)
	fmt.Println("Session:", sessionID)
	fmt.Println("输入 exit 退出")
	fmt.Println()

	reader := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("你> ")

		if !reader.Scan() {
			break
		}

		input := strings.TrimSpace(reader.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("已退出。")
			return
		}

		events, err := r.Run(ctx, userID, sessionID, model.NewUserMessage(input))
		if err != nil {
			fmt.Println("运行失败:", err)
			continue
		}

		handleEvents(events)
		fmt.Println()
	}

	if err := reader.Err(); err != nil {
		log.Fatal(err)
	}
}

func handleEvents(events <-chan *event.Event) {
	fmt.Print("助手> ")

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
			fmt.Println("[Skill Tool Calls]")
			for _, tc := range choice.Message.ToolCalls {
				fmt.Printf("- name=%s id=%s args=%s\n",
					tc.Function.Name,
					tc.ID,
					string(tc.Function.Arguments),
				)
			}
			fmt.Println("[Executing tools...]")
			continue
		}

		for _, ch := range evt.Response.Choices {
			if ch.Message.Role == model.RoleTool && ch.Message.ToolID != "" {
				fmt.Printf("\n[Skill Tool Result] id=%s result=%s\n",
					ch.Message.ToolID,
					strings.TrimSpace(ch.Message.Content),
				)
			}
		}

		if choice.Delta.Content != "" {
			fmt.Print(choice.Delta.Content)
		}

		if choice.Message.Content != "" {
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