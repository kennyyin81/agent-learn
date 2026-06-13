package localtools

import (
	"encoding/json"
	"fmt"
	"math"

	"ollama-tool-demo/internal/llm"
)

type CalculatorArgs struct {
	A  float64 `json:"a"`
	B  float64 `json:"b"`
	Op string  `json:"op"`
}

func CalculatorSchema() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.ToolFunction{
			Name:        "calculator",
			Description: "Calculate exactly two numbers. Use this tool for arithmetic. The op must be one of add, sub, mul, div.",
			Parameters: map[string]any{
				"type":     "object",
				"required": []string{"a", "b", "op"},
				"properties": map[string]any{
					"a": map[string]any{
						"type":        "number",
						"description": "The first number",
					},
					"b": map[string]any{
						"type":        "number",
						"description": "The second number",
					},
					"op": map[string]any{
						"type":        "string",
						"description": "The operation to perform",
						"enum":        []string{"add", "sub", "mul", "div"},
					},
				},
			},
		},
	}
}

func RunCalculator(args json.RawMessage) (string, error) {
	var input CalculatorArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("calculator 参数解析失败: %w", err)
	}

	var result float64

	switch input.Op {
	case "add":
		result = input.A + input.B
	case "sub":
		result = input.A - input.B
	case "mul":
		result = input.A * input.B
	case "div":
		if input.B == 0 {
			return "", fmt.Errorf("除数不能为 0")
		}
		result = input.A / input.B
	default:
		return "", fmt.Errorf("不支持的 op: %s", input.Op)
	}

	if math.IsNaN(result) || math.IsInf(result, 0) {
		return "", fmt.Errorf("计算结果非法")
	}

	output := map[string]any{
		"a":      input.A,
		"b":      input.B,
		"op":     input.Op,
		"result": result,
	}

	outputBytes, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("calculator 结果序列化失败: %w", err)
	}

	return string(outputBytes), nil
}