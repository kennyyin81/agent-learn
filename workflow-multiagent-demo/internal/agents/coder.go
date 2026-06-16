package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"workflow-multiagent-demo/internal/llm"
)

type Coder struct {
	LLM *llm.Client
}

func NewCoder(client *llm.Client) *Coder {
	return &Coder{LLM: client}
}
/**
 * 代码输入结构
 * UserTask: 用户需求
 * Plan: 开发计划
 * PreviousFiles: 上一版生成的文件列表
 * ExecutionFeedback: 执行反馈信息 例如 go test ./... 报错内容
 * ReviewFeedback: 审查反馈信息
 * Attempt: 当前尝试次数
**/
type CodeInput struct {
	UserTask          string
	Plan              Plan
	PreviousFiles     []CodeFile
	ExecutionFeedback string
	ReviewFeedback    string
	Attempt           int
}

func (c *Coder) Run(ctx context.Context, input CodeInput) (CodeResult, error) {
	planBytes, _ := json.MarshalIndent(input.Plan, "", "  ")
	prevBytes, _ := json.MarshalIndent(input.PreviousFiles, "", "  ")

	var userPrompt strings.Builder

	userPrompt.WriteString("用户需求：\n")
	userPrompt.WriteString(input.UserTask)
	userPrompt.WriteString("\n\n开发计划：\n")
	userPrompt.Write(planBytes)

	if input.Attempt > 1 {
		userPrompt.WriteString("\n\n这是第 ")
		userPrompt.WriteString(fmt.Sprintf("%d", input.Attempt))
		userPrompt.WriteString(" 次生成/修复。")
		userPrompt.WriteString("\n\n上一版文件：\n")
		userPrompt.Write(prevBytes)
		userPrompt.WriteString("\n\n执行反馈：\n")
		userPrompt.WriteString(input.ExecutionFeedback)
		userPrompt.WriteString("\n\n审查反馈：\n")
		userPrompt.WriteString(input.ReviewFeedback)
	}

	messages := []llm.Message{
		{
			Role: "system",
			Content: `
				你是 Coder Agent，负责生成可运行、可测试的 Go 代码。

				你必须只返回 JSON object，不要 Markdown，不要解释文字。

				JSON 格式：
				{
					"files": [
						{
							"path": "go.mod",
							"content": "module generated\n\ngo 1.22\n"
						},
						{
							"path": "xxx.go",
							"content": "package generated\n\n..."
						},
						{
							"path": "xxx_test.go",
							"content": "package generated\n\n..."
						}
					],
					"notes": "简短说明"
				}

				硬性要求：
				1. 所有 path 必须是相对路径。
				2. 禁止使用绝对路径。
				3. 禁止使用 ../。
				4. 必须包含 go.mod。
				5. 必须包含至少一个 _test.go 文件。
				6. 只使用 Go 标准库。
				7. 不要生成需要网络访问、系统命令、文件删除、环境变量读取的代码。
				8. 代码必须能通过 go test ./...。
				9. package 名建议使用 generated。
				`,
		},
		{
			Role:    "user",
			Content: userPrompt.String(),
		},
	}

	text, err := c.LLM.ChatJSON(ctx, messages, 0.1)
	if err != nil {
		return CodeResult{}, err
	}

	var result CodeResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return CodeResult{}, fmt.Errorf("解析 Coder JSON 失败: %w\n原始输出: %s", err, text)
	}

	if len(result.Files) == 0 {
		return CodeResult{}, fmt.Errorf("Coder 没有生成任何文件")
	}

	return result, nil
}