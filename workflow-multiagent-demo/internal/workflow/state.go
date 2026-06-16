package workflow

import (
	"workflow-multiagent-demo/internal/agents"
	"workflow-multiagent-demo/internal/tools"
)

// State 代表了整个工作流的状态，包含了用户任务、计划、代码文件、执行结果、评审结果、最终报告等信息
type State struct {
	UserTask     string
	Plan         agents.Plan
	Files        []agents.CodeFile
	Execution    tools.ExecutionResult
	Review       agents.ReviewResult
	FinalReport  agents.FinalReport
	Attempts     int
	MaxAttempts  int
	WorkspaceDir string
}