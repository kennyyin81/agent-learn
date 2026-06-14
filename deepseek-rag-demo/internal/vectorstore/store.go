package vectorstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"deepseek-rag-demo/internal/ragtypes"
)

type Store struct {
	EmbeddingModel string           `json:"embedding_model"`
	Chunks         []ragtypes.Chunk `json:"chunks"`
}

func Save(path string, store Store) error {
	// 取出文件所在目录，如果目录不存在就创建
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建 store 目录失败: %w", err)
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化向量库失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入向量库失败: %w", err)
	}

	return nil
}

func Load(path string) (Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Store{}, fmt.Errorf("读取向量库失败: %w", err)
	}

	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return Store{}, fmt.Errorf("解析向量库失败: %w", err)
	}

	if store.EmbeddingModel == "" {
		return Store{}, fmt.Errorf("向量库缺少 embedding_model")
	}

	if len(store.Chunks) == 0 {
		return Store{}, fmt.Errorf("向量库没有 chunks")
	}

	return store, nil
}