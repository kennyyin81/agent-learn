package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"deepseek-rag-demo/internal/ragtypes"
)

func LoadDocuments(dataDir string) ([]ragtypes.Document, error) {
	var docs []ragtypes.Document
	// 遍历一个目录下的所有文件和子目录，并对每个文件执行一个函数
	// path: 当前访问的文件或目录的路径
	// d: 当前访问的文件或目录的信息
	// walkErr: 遍历过程中发生的错误
	err := filepath.WalkDir(dataDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}
		// 只处理 .txt 和 .md 文件
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".txt" && ext != ".md" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取文件失败 %s: %w", path, err)
		}

		rel, err := filepath.Rel(dataDir, path)
		if err != nil {
			// 如果无法计算相对路径，就使用文件名
			rel = filepath.Base(path)
		}

		text := strings.TrimSpace(string(data))
		if text == "" {
			return nil
		}

		docs = append(docs, ragtypes.Document{
			Source: rel,
			Text:   text,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return docs, nil
}