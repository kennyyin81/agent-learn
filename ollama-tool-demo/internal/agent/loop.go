package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"ollama-tool-demo/internal/llm"
	"ollama-tool-demo/internal/localtools"
)

type ConfirmFunc func(toolName string, args json.RawMessage) bool

type Options struct {
	MaxSteps int
	Stream   bool
	Confirm  ConfirmFunc
}

func Run(
	ctx context.Context,
	client *llm.Client,
	registry *localtools.Registry,
	messages []llm.Message,
	tools []llm.Tool,
	opts Options,
) (string, []llm.Message, error) {
	if opts.MaxSteps <= 0 {
		opts.MaxSteps = 8
	}

	for step := 1; step <= opts.MaxSteps; step++ {
		fmt.Fprintf(os.Stderr, "\n========== Agent Step %d ==========\n", step)

		var assistantMsg llm.Message
		var err error

		if opts.Stream {
			assistantMsg, err = client.ChatStream(ctx, messages, tools, func(delta string) {
				fmt.Print(delta)
			})
		} else {
			var resp llm.ChatResponse
			resp, err = client.Chat(ctx, messages, tools)
			assistantMsg = resp.Message
		}

		if err != nil {
			return "", messages, err
		}

		messages = append(messages, assistantMsg)

		if len(assistantMsg.ToolCalls) == 0 {
			content := strings.TrimSpace(assistantMsg.Content)
			if content == "" {
				return "", messages, fmt.Errorf("模型没有返回 tool_calls，也没有返回最终 content")
			}
			return content, messages, nil
		}

		for _, toolCall := range assistantMsg.ToolCalls {
			toolName := toolCall.Function.Name
			args := toolCall.Function.Arguments

			fmt.Fprintf(os.Stderr, "\n[模型请求调用工具] %s\n", toolName)
			fmt.Fprintf(os.Stderr, "[工具参数] %s\n", string(args))

			var result string

			if registry.RequiresConfirmation(toolName) {
				if opts.Confirm == nil {
					result = toToolErrorJSON(fmt.Errorf("工具 %s 需要人工确认，但没有提供 Confirm 函数", toolName))
				} else if !opts.Confirm(toolName, args) {
					result = toToolErrorJSON(fmt.Errorf("用户拒绝执行工具: %s", toolName))
				} else {
					result, err = registry.Run(toolName, args)
					if err != nil {
						result = toToolErrorJSON(err)
					}
				}
			} else {
				result, err = registry.Run(toolName, args)
				if err != nil {
					result = toToolErrorJSON(err)
				}
			}

			fmt.Fprintf(os.Stderr, "[工具结果] %s\n", result)

			messages = append(messages, llm.Message{
				Role:     "tool",
				ToolName: toolName,
				Content:  result,
			})
		}
	}

	return "", messages, fmt.Errorf("超过最大工具调用轮数 maxSteps=%d，可能进入死循环", opts.MaxSteps)
}

func NewStdinConfirm(reader *bufio.Reader) ConfirmFunc {
	return func(toolName string, args json.RawMessage) bool {
		fmt.Fprintf(os.Stderr, "\n[安全确认]\n")
		fmt.Fprintf(os.Stderr, "工具名: %s\n", toolName)
		fmt.Fprintf(os.Stderr, "参数: %s\n", string(args))
		fmt.Fprintf(os.Stderr, "这个工具会修改本地状态。输入 yes 才允许执行: ")
		// 带缓冲的读取器
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取确认输入失败: %v\n", err)
			return false
		}

		return strings.EqualFold(strings.TrimSpace(line), "yes")
	}
}

func toToolErrorJSON(err error) string {
	output := map[string]string{
		"error": err.Error(),
	}

	data, marshalErr := json.Marshal(output)
	if marshalErr != nil {
		return `{"error":"unknown tool error"}`
	}

	return string(data)
}