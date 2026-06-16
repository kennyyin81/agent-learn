package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"workflow-multiagent-demo/internal/llm"
	"workflow-multiagent-demo/internal/tools"
)

type Summarizer struct {
	LLM *llm.Client
}

func NewSummarizer(client *llm.Client) *Summarizer {
	return &Summarizer{LLM: client}
}

type SummaryInput struct {
	UserTask  string
	Plan      Plan
	Files     []CodeFile
	Execution tools.ExecutionResult
	Review    ReviewResult
	Attempts  int
}

func (s *Summarizer) Run(ctx context.Context, input SummaryInput) (FinalReport, error) {
	planBytes, _ := json.MarshalIndent(input.Plan, "", "  ")
	filesBytes, _ := json.MarshalIndent(input.Files, "", "  ")
	execBytes, _ := json.MarshalIndent(input.Execution, "", "  ")
	reviewBytes, _ := json.MarshalIndent(input.Review, "", "  ")

	var userPrompt strings.Builder
	userPrompt.WriteString("用户需求：\n")
	userPrompt.WriteString(input.UserTask)
	userPrompt.WriteString("\n\n计划：\n")
	userPrompt.Write(planBytes)
	userPrompt.WriteString("\n\n文件：\n")
	userPrompt.Write(filesBytes)
	userPrompt.WriteString("\n\n执行结果：\n")
	userPrompt.Write(execBytes)
	userPrompt.WriteString("\n\n审查结果：\n")
	userPrompt.Write(reviewBytes)
	userPrompt.WriteString("\n\n尝试次数：")
	userPrompt.WriteString(fmt.Sprintf("%d", input.Attempts))

	messages := []llm.Message{
		{
			Role: "system",
			Content: `
				你是 Summarizer Agent，负责把多 Agent 工作流的结果总结给用户。

				你必须只返回 JSON object，不要 Markdown，不要解释文字。

				JSON 格式：
				{
					"summary": "总结",
					"files": ["文件1", "文件2"],
					"test_result": "测试结果",
					"review": "审查结论",
					"next_steps": ["建议1", "建议2"]
				}

				要求：
				1. 用中文。
				2. 说明是否成功。
				3. 说明生成了哪些文件。
				4. 说明 go test 是否通过。
				5. 如果失败，要说明失败原因和下一步建议。
				`,
		},
		{
			Role:    "user",
			Content: userPrompt.String(),
		},
	}

	text, err := s.LLM.ChatJSON(ctx, messages, 0.2)
	if err != nil {
		return FinalReport{}, err
	}

	var result FinalReport
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return FinalReport{}, fmt.Errorf("解析 Summarizer JSON 失败: %w\n原始输出: %s", err, text)
	}

	return result, nil
}