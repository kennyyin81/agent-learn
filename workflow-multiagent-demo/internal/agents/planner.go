package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"workflow-multiagent-demo/internal/llm"
)

type Planner struct {
	LLM *llm.Client
}

func NewPlanner(client *llm.Client) *Planner {
	return &Planner{LLM: client}
}

func (p *Planner) Run(ctx context.Context, userTask string) (Plan, error) {
	messages := []llm.Message{
		{
			Role: "system",
			Content: `
				你是 Planner Agent，负责把用户需求拆成清晰的开发计划。

				你必须只返回 JSON object，不要 Markdown，不要解释文字。

				JSON 格式：
				{
					"goal": "总体目标",
					"tasks": ["任务1", "任务2"],
					"constraints": ["约束1", "约束2"],
					"expected_files": ["go.mod", "xxx.go", "xxx_test.go"]
				}

				要求：
				1. 面向 Go 语言项目。
				2. 默认只使用 Go 标准库。
				3. 计划要小而可执行。
				4. expected_files 必须包含 go.mod 和至少一个 _test.go 文件。
				`,
		},
		{
			Role:    "user",
			Content: userTask,
		},
	}

	text, err := p.LLM.ChatJSON(ctx, messages, 0.1)
	if err != nil {
		return Plan{}, err
	}

	var plan Plan
	if err := json.Unmarshal([]byte(text), &plan); err != nil {
		return Plan{}, fmt.Errorf("解析 Planner JSON 失败: %w\n原始输出: %s", err, text)
	}

	if plan.Goal == "" {
		return Plan{}, fmt.Errorf("Planner 输出缺少 goal")
	}

	if len(plan.Tasks) == 0 {
		return Plan{}, fmt.Errorf("Planner 输出缺少 tasks")
	}

	return plan, nil
}