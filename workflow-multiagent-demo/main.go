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

	"workflow-multiagent-demo/internal/agents"
	"workflow-multiagent-demo/internal/llm"
	"workflow-multiagent-demo/internal/workflow"
)

func main() {
	baseURL := flag.String("base-url", env("DEEPSEEK_BASE_URL", "https://api.deepseek.com"), "DeepSeek Base URL")
	model := flag.String("model", env("DEEPSEEK_MODEL", "deepseek-v4-flash"), "DeepSeek model")
	workspaceDir := flag.String("workspace", "workspace", "代码生成和测试目录")
	maxAttempts := flag.Int("max-attempts", 2, "最多生成/修复次数")
	task := flag.String("task", "", "用户任务")
	flag.Parse()

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if strings.TrimSpace(apiKey) == "" {
		log.Fatal("缺少 DEEPSEEK_API_KEY 环境变量")
	}

	userTask := strings.TrimSpace(*task)
	if userTask == "" {
		fmt.Println("请输入一个代码任务，例如：")
		fmt.Println("帮我写一个 Go 函数 Add(a,b int) int，并写单元测试")
		fmt.Print("\n你> ")

		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		userTask = strings.TrimSpace(line)
	}

	if userTask == "" {
		log.Fatal("任务不能为空")
	}

	client := llm.NewClient(*baseURL, apiKey, *model)

	planner := agents.NewPlanner(client)
	coder := agents.NewCoder(client)
	reviewer := agents.NewReviewer(client)
	summarizer := agents.NewSummarizer(client)

	runner := workflow.NewRunner(planner, coder, reviewer, summarizer)

	state := &workflow.State{
		UserTask:     userTask,
		MaxAttempts:  *maxAttempts,
		WorkspaceDir: *workspaceDir,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	if err := runner.Run(ctx, state); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n========== 最终报告 ==========")
	fmt.Println("总结:", state.FinalReport.Summary)
	fmt.Println("文件:")
	for _, file := range state.FinalReport.Files {
		fmt.Println("-", file)
	}
	fmt.Println("测试结果:", state.FinalReport.TestResult)
	fmt.Println("审查结论:", state.FinalReport.Review)

	if len(state.FinalReport.NextSteps) > 0 {
		fmt.Println("下一步建议:")
		for _, step := range state.FinalReport.NextSteps {
			fmt.Println("-", step)
		}
	}
}

func env(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}