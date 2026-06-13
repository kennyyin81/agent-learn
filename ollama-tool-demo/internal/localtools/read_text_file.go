package localtools

import (
	"encoding/json"
	"fmt"
	"os"

	"ollama-tool-demo/internal/llm"
)

const MaxReadBytes = 1024 * 1024

type ReadTextFileArgs struct {
	Path string `json:"path"`
}

func ReadTextFileSchema() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.ToolFunction{
			Name:        "read_text_file",
			Description: "Read a UTF-8 text file from the safe workspace directory. Only relative paths under workspace are allowed.",
			Parameters: map[string]any{
				"type":     "object",
				"required": []string{"path"},
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Relative file path under workspace, for example note.txt",
					},
				},
			},
		},
	}
}

func RunReadTextFile(args json.RawMessage) (string, error) {
	var input ReadTextFileArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("read_text_file 参数解析失败: %w", err)
	}

	fullPath, err := safeJoin(input.Path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return "", fmt.Errorf("读取文件信息失败: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("目标是目录，不是文件: %s", input.Path)
	}

	if info.Size() > MaxReadBytes {
		return "", fmt.Errorf("文件过大，超过 %d bytes", MaxReadBytes)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	output := map[string]any{
		"path":    input.Path,
		"size":    info.Size(),
		"content": string(data),
	}

	outputBytes, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("read_text_file 结果序列化失败: %w", err)
	}

	return string(outputBytes), nil
}