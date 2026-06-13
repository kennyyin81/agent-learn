package localtools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ollama-tool-demo/internal/llm"
)

type WriteTextFileArgs struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Overwrite bool   `json:"overwrite"`
}

func WriteTextFileSchema() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.ToolFunction{
			Name:        "write_text_file",
			Description: "Write UTF-8 text content to a file under the safe workspace directory. This tool changes local files and requires user confirmation.",
			Parameters: map[string]any{
				"type":     "object",
				"required": []string{"path", "content", "overwrite"},
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Relative file path under workspace, for example hello.txt",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "UTF-8 text content to write.",
					},
					"overwrite": map[string]any{
						"type":        "boolean",
						"description": "Whether to overwrite the file if it already exists.",
					},
				},
			},
		},
	}
}

func RunWriteTextFile(args json.RawMessage) (string, error) {
	var input WriteTextFileArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("write_text_file 参数解析失败: %w", err)
	}

	fullPath, err := safeJoin(input.Path)
	if err != nil {
		return "", err
	}
	// 递归创建目录（如果不存在） 0755 = rwxr-xr-x
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}
	// 是否覆盖文件
	if _, err := os.Stat(fullPath); err == nil && !input.Overwrite {
		return "", fmt.Errorf("文件已存在，且 overwrite=false: %s", input.Path)
	}
	// 0644 = rw-r--r--
	if err := os.WriteFile(fullPath, []byte(input.Content), 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	output := map[string]any{
		"path":      input.Path,
		"bytes":     len([]byte(input.Content)),
		"overwrite": input.Overwrite,
		"status":    "written",
	}

	outputBytes, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("write_text_file 结果序列化失败: %w", err)
	}

	return string(outputBytes), nil
}