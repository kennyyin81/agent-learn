package localtools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const SafeRoot = "workspace"

// 路径安全检查
func safeJoin(userPath string) (string, error) {
	if strings.TrimSpace(userPath) == "" {
		return "", fmt.Errorf("path 不能为空")
	}

	if filepath.IsAbs(userPath) {
		return "", fmt.Errorf("只允许相对路径，不允许绝对路径: %s", userPath)
	}

	rootAbs, err := filepath.Abs(SafeRoot)  // 获取安全根目录的绝对路径
	if err != nil {
		return "", fmt.Errorf("解析安全根目录失败: %w", err)
	}

	fullAbs, err := filepath.Abs(filepath.Join(rootAbs, filepath.Clean(userPath)))
	if err != nil {
		return "", fmt.Errorf("解析目标路径失败: %w", err)
	}

	rel, err := filepath.Rel(rootAbs, fullAbs)
	if err != nil {
		return "", fmt.Errorf("计算相对路径失败: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("禁止访问 workspace 之外的路径: %s", userPath)
	}

	return fullAbs, nil
}