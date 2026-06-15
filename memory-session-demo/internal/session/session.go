package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"memory-session-demo/internal/llm"
)

type Session struct {
	ID        string        `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Messages  []llm.Message `json:"messages"`
}

func New(id string) *Session {
	now := time.Now()
	return &Session{
		ID:     id,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  []llm.Message{},
	}
}

func LoadOrNew(path string, id string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return New(id), nil
		}
		return nil, fmt.Errorf("读取 session 失败: %w", err)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("解析 session 失败: %w", err)
	}

	if s.ID == "" {
		s.ID = id
	}

	return &s, nil
}

func (s *Session) Add(role string, content string) {
	s.Messages = append(s.Messages, llm.Message{
		Role:    role,
		Content: content,
	})
	s.UpdatedAt = time.Now()
}

func (s *Session) Recent(limit int) []llm.Message {
	if limit <= 0 || len(s.Messages) <= limit {
		return append([]llm.Message(nil), s.Messages...)
	}

	return append([]llm.Message(nil), s.Messages[len(s.Messages)-limit:]...)
}

func (s *Session) Clear() {
	s.Messages = []llm.Message{}
	s.UpdatedAt = time.Now()
}

func (s *Session) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建 session 目录失败: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 session 失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入 session 失败: %w", err)
	}

	return nil
}