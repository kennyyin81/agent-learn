package localtools

import (
	"encoding/json"
	"fmt"

	"ollama-tool-demo/internal/llm"
)

type Runner func(args json.RawMessage) (string, error)

type ToolSpec struct {
	Schema          llm.Tool
	Runner          Runner
	RequiresConfirm bool
}

type Registry struct {
	specs map[string]ToolSpec
}

func NewRegistry() *Registry {
	specs := map[string]ToolSpec{
		"calculator": {
			Schema:          CalculatorSchema(),
			Runner:          RunCalculator,
			RequiresConfirm: false,
		},
		"get_current_time": {
			Schema:          CurrentTimeSchema(),
			Runner:          RunCurrentTime,
			RequiresConfirm: false,
		},
		"random_number": {
			Schema:          RandomNumberSchema(),
			Runner:          RunRandomNumber,
			RequiresConfirm: false,
		},
		"read_text_file": {
			Schema:          ReadTextFileSchema(),
			Runner:          RunReadTextFile,
			RequiresConfirm: false,
		},
		"write_text_file": {
			Schema:          WriteTextFileSchema(),
			Runner:          RunWriteTextFile,
			RequiresConfirm: true,
		},
	}

	return &Registry{
		specs: specs,
	}
}

func (r *Registry) Schemas() []llm.Tool {
	tools := make([]llm.Tool, 0, len(r.specs))
	for _, spec := range r.specs {
		tools = append(tools, spec.Schema)
	}
	return tools
}

func (r *Registry) RequiresConfirmation(name string) bool {
	spec, ok := r.specs[name]
	if !ok {
		return false
	}
	return spec.RequiresConfirm
}

func (r *Registry) Run(name string, args json.RawMessage) (string, error) {
	spec, ok := r.specs[name]
	if !ok {
		return "", fmt.Errorf("未知工具: %s", name)
	}

	return spec.Runner(args)
}