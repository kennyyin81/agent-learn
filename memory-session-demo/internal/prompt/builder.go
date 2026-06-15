package prompt

import (
	"fmt"
	"strings"

	"memory-session-demo/internal/llm"
	"memory-session-demo/internal/memory"
)

type BuildInput struct {
	UserInput      string
	RecentHistory  []llm.Message
	RelevantMemory []memory.Item
}

func BuildMessages(input BuildInput) []llm.Message {
	system := buildSystemPrompt(input.RelevantMemory)

	messages := []llm.Message{
		{
			Role:    "system",
			Content: system,
		},
	}

	messages = append(messages, input.RecentHistory...)

	messages = append(messages, llm.Message{
		Role:    "user",
		Content: input.UserInput,
	})

	return messages
}

func buildSystemPrompt(memories []memory.Item) string {
	var builder strings.Builder

	builder.WriteString(`
		你是一个严谨、耐心的中文编程学习助手。

		你需要遵守：
		1. 优先根据当前用户问题回答。
		2. 可以利用“长期记忆”理解用户背景和偏好。
		3. 不要把长期记忆当作绝对事实；如果和当前用户输入冲突，以当前用户输入为准。
		4. 不要主动暴露所有记忆，除非用户询问。
		5. 如果用户要求删除、修改或查看记忆，应提示他使用命令完成。
		6. 回答要适合正在学习 Go、LLM、Agent、RAG 的用户。
	`)

	builder.WriteString("\n\n长期记忆：\n")

	if len(memories) == 0 {
		builder.WriteString("无相关长期记忆。\n")
		return strings.TrimSpace(builder.String())
	}

	for i, item := range memories {
		builder.WriteString(fmt.Sprintf("[%d] type=%s, id=%s, text=%s\n",
			i+1,
			item.Type,
			item.ID,
			item.Text,
		))
	}

	return strings.TrimSpace(builder.String())
}