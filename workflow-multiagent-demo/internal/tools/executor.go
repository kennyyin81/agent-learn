package tools

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"workflow-multiagent-demo/internal/shared"
)

/**
* 保存命令执行结果
* Command: 执行的命令 go test ./...
* Dir: 执行命令的目录
* Success: 命令是否成功执行
* ExitCode: 命令的退出代码
* Stdout: 命令的标准输出
* Stderr: 命令的标准错误输出
* DurationMS: 命令执行的持续时间（毫秒）
* TimedOut: 命令是否因为超时而被终止
**/
type ExecutionResult = shared.ExecutionResult

func RunGoTests(parentCtx context.Context, dir string) ExecutionResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
	defer cancel()
	// 创建一个命令来执行 go test ./...，并设置工作目录为 dir
	cmd := exec.CommandContext(ctx, "go", "test", "./...")
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	duration := time.Since(start)

	result := ExecutionResult{
		Command:    "go test ./...",
		Dir:        dir,
		DurationMS: duration.Milliseconds(),
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
	}

	if ctx.Err() == context.DeadlineExceeded { // 超时
		result.TimedOut = true
		result.Success = false
		result.ExitCode = -1
		return result
	}

	if err == nil { // 正常执行成功
		result.Success = true
		result.ExitCode = 0
		return result
	}

	result.Success = false

	if exitErr, ok := err.(*exec.ExitError); ok { // go test 失败返回失败码
		result.ExitCode = exitErr.ExitCode()
	} else { // 系统层面错误
		result.ExitCode = -1
		result.Stderr += "\n" + err.Error()
	}

	return result
}
