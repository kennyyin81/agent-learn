package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

type Item struct {
	ID         string    `json:"id"`
	Text       string    `json:"text"`
	Type       string    `json:"type"`  // 记忆类型 user_note/perference/project/skill 等
	Source     string    `json:"source"`  // 记忆来源 manual/conversation/system 等
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Confidence float64   `json:"confidence"`  // 置信度
}

type Store struct {
	Items []Item `json:"items"`
}

func LoadOrNew(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{
				Items: []Item{},
			}, nil
		}
		return nil, fmt.Errorf("读取 memory 失败: %w", err)
	}

	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("解析 memory 失败: %w", err)
	}

	if s.Items == nil {
		s.Items = []Item{}
	}

	return &s, nil
}

func (s *Store) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建 memory 目录失败: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 memory 失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入 memory 失败: %w", err)
	}

	return nil
}

func (s *Store) Add(text string, memoryType string, source string) (Item, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Item{}, fmt.Errorf("memory 内容不能为空")
	}

	if len([]rune(text)) > 300 {
		return Item{}, fmt.Errorf("memory 内容过长，请压缩到 300 字以内")
	}

	if looksLikeSecret(text) {
		return Item{}, fmt.Errorf("疑似包含密钥、token 或密码，不建议写入长期记忆")
	}

	now := time.Now()

	item := Item{
		ID:         newID(),
		Text:       text,
		Type:       defaultString(memoryType, "user_note"),
		Source:     defaultString(source, "manual"),
		CreatedAt:  now,
		UpdatedAt:  now,
		Confidence: 1.0,
	}

	s.Items = append(s.Items, item)

	return item, nil
}

func (s *Store) Delete(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}

	for i, item := range s.Items {
		if item.ID == id {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			return true
		}
	}

	return false
}

func (s *Store) List() []Item {
	items := append([]Item(nil), s.Items...)

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	return items
}

type ScoredItem struct {
	Item  Item
	Score float64
}

func (s *Store) Search(query string, limit int) []Item {
	if limit <= 0 {
		limit = 5
	}

	query = strings.TrimSpace(query)
	if query == "" || len(s.Items) == 0 {
		return nil
	}

	scored := make([]ScoredItem, 0, len(s.Items))

	for _, item := range s.Items {
		score := lexicalScore(query, item.Text)
		if score > 0 {
			scored = append(scored, ScoredItem{
				Item:  item,
				Score: score,
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if len(scored) > limit {
		scored = scored[:limit]
	}

	result := make([]Item, 0, len(scored))
	for _, x := range scored {
		result = append(result, x.Item)
	}

	return result
}

func newID() string {
	return fmt.Sprintf("mem_%d", time.Now().UnixNano())
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func looksLikeSecret(text string) bool {
	lower := strings.ToLower(text)

	keywords := []string{
		"api_key",
		"apikey",
		"secret",
		"token",
		"password",
		"passwd",
		"bearer",
		"sk-",
	}

	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}

	return false
}

func lexicalScore(query string, text string) float64 {
	q := normalize(query)
	t := normalize(text)

	if q == "" || t == "" {
		return 0
	}

	qGrams := charNgrams(q, 2)
	if len(qGrams) == 0 {
		if strings.Contains(t, q) {
			return 1
		}
		return 0
	}

	hit := 0
	for _, gram := range qGrams {
		if strings.Contains(t, gram) {
			hit++
		}
	}

	return float64(hit) / float64(len(qGrams))
}

func normalize(s string) string {
	var builder strings.Builder

	for _, r := range s {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			continue
		}
		builder.WriteRune(unicode.ToLower(r))
	}

	return builder.String()
}

func charNgrams(s string, n int) []string {
	runes := []rune(s)
	if len(runes) < n {
		return nil
	}

	grams := make([]string, 0, len(runes)-n+1)
	for i := 0; i <= len(runes)-n; i++ {
		grams = append(grams, string(runes[i:i+n]))
	}

	return grams
}