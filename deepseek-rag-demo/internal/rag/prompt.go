package rag

import (
	"fmt"
	"strings"

	"deepseek-rag-demo/internal/ragtypes"
)

func BuildMessages(question string, contexts []ragtypes.ScoredChunk) []map[string]string {
	system := `
		你是一个严谨的研究生论文知识库问答助手。

		你必须遵守以下规则：
		1. 只能根据用户提供的“论文片段”回答。
		2. 如果论文片段中没有答案，必须明确说：“提供的论文片段中没有找到相关信息。”
		3. 不要使用外部知识补充、猜测或编造。
		4. 回答中必须使用 [S1]、[S2] 这样的格式标注来源。
		5. 如果多个片段支持同一结论，可以同时引用多个来源，例如 [S1][S3]。
		6. 回答要用中文，尽量清晰、具体、适合学术讨论。
		`

	user := buildUserPrompt(question, contexts)

	return []map[string]string{
		{
			"role":    "system",
			"content": strings.TrimSpace(system),
		},
		{
			"role":    "user",
			"content": user,
		},
	}
}

func buildUserPrompt(question string, contexts []ragtypes.ScoredChunk) string {
	var builder strings.Builder

	builder.WriteString("用户问题：\n")
	builder.WriteString(question)
	builder.WriteString("\n\n")

	builder.WriteString("论文片段：\n")

	for i, item := range contexts {
		sourceID := fmt.Sprintf("S%d", i+1)

		builder.WriteString(fmt.Sprintf("\n[%s]\n", sourceID))
		builder.WriteString(fmt.Sprintf("来源：%s，chunk=%d，final_score=%.4f，vector_score=%.4f，lexical_score=%.4f\n",
			item.Chunk.Source,
			item.Chunk.Index,
			item.FinalScore,
			item.VectorScore,
			item.LexicalScore,
		))
		builder.WriteString("内容：\n")
		builder.WriteString(item.Chunk.Text)
		builder.WriteString("\n")
	}

	builder.WriteString(`
		请基于上面的论文片段回答用户问题。
		要求：
		- 不能脱离论文片段。
		- 必须带来源引用。
		- 如果信息不足，要直接说明不足。
	`)

	return builder.String()
}