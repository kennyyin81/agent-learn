package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"workflow-multiagent-demo/internal/llm"
	"workflow-multiagent-demo/internal/tools"
)

type Reviewer struct {
	LLM *llm.Client
}

func NewReviewer(client *llm.Client) *Reviewer {
	return &Reviewer{LLM: client}
}

type ReviewInput struct {
	UserTask  string
	Plan      Plan
	Files     []CodeFile
	Execution tools.ExecutionResult
}

func (r *Reviewer) Run(ctx context.Context, input ReviewInput) (ReviewResult, error) {
	planBytes, _ := json.MarshalIndent(input.Plan, "", "  ")
	filesBytes, _ := json.MarshalIndent(input.Files, "", "  ")
	execBytes, _ := json.MarshalIndent(input.Execution, "", "  ")

	var userPrompt strings.Builder
	userPrompt.WriteString("用户需求：\n")
	userPrompt.WriteString(input.UserTask)
	userPrompt.WriteString("\n\n开发计划：\n")
	userPrompt.Write(planBytes)
	userPrompt.WriteString("\n\n生成文件：\n")
	userPrompt.Write(filesBytes)
	userPrompt.WriteString("\n\n执行结果：\n")
	userPrompt.Write(execBytes)

	messages := []llm.Message{
		{
			Role: "system",
			Content: `
				你是 Reviewer Agent，负责审查 Go 代码是否满足用户需求。

				你必须只返回 JSON object，不要 Markdown，不要解释文字。

				JSON 格式：
				{
					"passed": true,
					"issues": [],
					"suggestions": [],
					"repair_instruction": ""
				}

				判断标准：
				1. 如果 go test 失败，passed 必须是 false。
				2. 如果代码不满足用户需求，passed 必须是 false。
				3. 如果缺少测试，passed 必须是 false。
				4. 如果有安全风险，例如执行系统命令、读取环境变量、访问网络、删除文件，passed 必须是 false。
				5. repair_instruction 要给 Coder Agent 使用，说明如何修复。
				`,
		},
		{
			Role:    "user",
			Content: userPrompt.String(),
		},
	}

	text, err := r.LLM.ChatJSON(ctx, messages, 0.1)
	if err != nil {
		return ReviewResult{}, err
	}

	var result ReviewResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return ReviewResult{}, fmt.Errorf("解析 Reviewer JSON 失败: %w\n原始输出: %s", err, text)
	}

	return result, nil
}
