package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"workflow-multiagent-demo/internal/shared"
)

func ResetWorkspace(root string) error {
	if strings.TrimSpace(root) == "" {
		return fmt.Errorf("workspace root 不能为空")
	}

	if err := os.RemoveAll(root); err != nil {
		return fmt.Errorf("清理 workspace 失败: %w", err)
	}

	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("创建 workspace 失败: %w", err)
	}

	return nil
}

func WriteCodeFiles(root string, files []shared.CodeFile) error {
	for _, file := range files {
		fullPath, err := safeJoin(root, file.Path)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}

		if err := os.WriteFile(fullPath, []byte(file.Content), 0644); err != nil {
			return fmt.Errorf("写入文件失败 %s: %w", file.Path, err)
		}
	}

	return nil
}

func safeJoin(root string, userPath string) (string, error) {
	userPath = strings.TrimSpace(userPath)
	if userPath == "" {
		return "", fmt.Errorf("文件路径不能为空")
	}

	if filepath.IsAbs(userPath) {
		return "", fmt.Errorf("禁止绝对路径: %s", userPath)
	}

	clean := filepath.Clean(userPath)

	if clean == "." {
		return "", fmt.Errorf("非法路径: %s", userPath)
	}

	if strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("禁止访问 workspace 外部路径: %s", userPath)
	}

	base := filepath.Base(clean)
	if strings.HasPrefix(base, ".") {
		return "", fmt.Errorf("禁止写入隐藏文件: %s", userPath)
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("解析 workspace 失败: %w", err)
	}

	fullAbs, err := filepath.Abs(filepath.Join(rootAbs, clean))
	if err != nil {
		return "", fmt.Errorf("解析目标路径失败: %w", err)
	}

	rel, err := filepath.Rel(rootAbs, fullAbs)
	if err != nil {
		return "", fmt.Errorf("计算相对路径失败: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("禁止访问 workspace 外部路径: %s", userPath)
	}

	return fullAbs, nil
}

func FilePaths(files []shared.CodeFile) []string {
	result := make([]string, 0, len(files))
	for _, file := range files {
		result = append(result, file.Path)
	}
	return result
}
