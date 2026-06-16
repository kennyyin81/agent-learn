package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"workflow-multiagent-demo/internal/agents"
	"workflow-multiagent-demo/internal/tools"
)

type Runner struct {
	Planner    *agents.Planner
	Coder      *agents.Coder
	Reviewer   *agents.Reviewer
	Summarizer *agents.Summarizer
}

func NewRunner(
	planner *agents.Planner,
	coder *agents.Coder,
	reviewer *agents.Reviewer,
	summarizer *agents.Summarizer,
) *Runner {
	return &Runner{
		Planner:    planner,
		Coder:      coder,
		Reviewer:   reviewer,
		Summarizer: summarizer,
	}
}

func (r *Runner) Run(ctx context.Context, state *State) error {
	if state.MaxAttempts <= 0 {
		state.MaxAttempts = 2
	}

	fmt.Println("\n========== Step 1: Planner Agent ==========")

	plan, err := r.Planner.Run(ctx, state.UserTask)
	if err != nil {
		return fmt.Errorf("Planner 失败: %w", err)
	}
	state.Plan = plan

	printJSON("Plan", plan)

	var executionFeedback string
	var reviewFeedback string

	for attempt := 1; attempt <= state.MaxAttempts; attempt++ {
		state.Attempts = attempt

		fmt.Printf("\n========== Step 2: Coder Agent，Attempt %d ==========\n", attempt)

		codeResult, err := r.Coder.Run(ctx, agents.CodeInput{
			UserTask:          state.UserTask,
			Plan:              state.Plan,
			PreviousFiles:     state.Files,
			ExecutionFeedback: executionFeedback,
			ReviewFeedback:    reviewFeedback,
			Attempt:           attempt,
		})
		if err != nil {
			return fmt.Errorf("Coder 失败: %w", err)
		}

		state.Files = codeResult.Files
		printJSON("Code notes", map[string]any{
			"files": tools.FilePaths(state.Files),
			"notes": codeResult.Notes,
		})

		fmt.Println("\n========== Step 3: File Tool ==========")

		if err := tools.ResetWorkspace(state.WorkspaceDir); err != nil {
			return fmt.Errorf("重置 workspace 失败: %w", err)
		}

		if err := tools.WriteCodeFiles(state.WorkspaceDir, state.Files); err != nil {
			return fmt.Errorf("写入代码文件失败: %w", err)
		}

		fmt.Println("文件已写入:", state.WorkspaceDir)

		fmt.Println("\n========== Step 4: Executor Tool ==========")

		execResult := tools.RunGoTests(ctx, state.WorkspaceDir)
		state.Execution = execResult
		printJSON("Execution", execResult)

		fmt.Println("\n========== Step 5: Reviewer Agent ==========")

		review, err := r.Reviewer.Run(ctx, agents.ReviewInput{
			UserTask:  state.UserTask,
			Plan:      state.Plan,
			Files:     state.Files,
			Execution: state.Execution,
		})
		if err != nil {
			return fmt.Errorf("Reviewer 失败: %w", err)
		}

		state.Review = review
		printJSON("Review", review)

		if state.Execution.Success && state.Review.Passed {
			fmt.Println("\n工作流判断：测试通过，审查通过，进入总结。")
			break
		}

		if attempt == state.MaxAttempts {
			fmt.Println("\n工作流判断：达到最大修复次数，进入总结。")
			break
		}

		executionFeedback = compactExecutionFeedback(state.Execution)
		reviewFeedback = compactReviewFeedback(state.Review)

		fmt.Println("\n工作流判断：测试或审查未通过，回到 Coder Agent 修复。")
	}

	fmt.Println("\n========== Step 6: Summarizer Agent ==========")

	finalReport, err := r.Summarizer.Run(ctx, agents.SummaryInput{
		UserTask:  state.UserTask,
		Plan:      state.Plan,
		Files:     state.Files,
		Execution: state.Execution,
		Review:    state.Review,
		Attempts:  state.Attempts,
	})
	if err != nil {
		return fmt.Errorf("Summarizer 失败: %w", err)
	}

	state.FinalReport = finalReport

	return nil
}

func compactExecutionFeedback(result tools.ExecutionResult) string {
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

func compactReviewFeedback(result agents.ReviewResult) string {
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data)
}

func printJSON(title string, value any) {
	data, _ := json.MarshalIndent(value, "", "  ")
	fmt.Println(title + ":")
	fmt.Println(string(data))
}