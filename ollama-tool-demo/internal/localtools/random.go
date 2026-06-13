package localtools

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"ollama-tool-demo/internal/llm"
)


type RandomNumberArgs struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

func RandomNumberSchema() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.ToolFunction{
			Name:        "random_number",
			Description: "Generate a random integer between min and max, inclusive.",
			Parameters: map[string]any{
				"type":     "object",
				"required": []string{"min", "max"},
				"properties": map[string]any{
					"min": map[string]any{
						"type":        "integer",
						"description": "The minimum integer value, inclusive.",
					},
					"max": map[string]any{
						"type":        "integer",
						"description": "The maximum integer value, inclusive.",
					},
				},
			},
		},
	}
}

func RunRandomNumber(args json.RawMessage) (string, error) {
var input RandomNumberArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("random_number 参数解析失败: %w", err)
	}

	if input.Min > input.Max {
		return "", fmt.Errorf("min 不能大于 max")
	}

	rangeSize := input.Max - input.Min + 1
	if rangeSize <= 0 {
		return "", fmt.Errorf("随机数范围过大或非法")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(rangeSize))
	if err != nil {
		return "", fmt.Errorf("生成随机数失败: %w", err)
	}

	result := input.Min + n.Int64()

	output := map[string]any{
		"min":    input.Min,
		"max":    input.Max,
		"result": result,
	}

	outputBytes, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("random_number 结果序列化失败: %w", err)
	}

	return string(outputBytes), nil
}
