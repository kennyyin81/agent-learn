package chunker

import (
	"fmt"
	"strings"

	"deepseek-rag-demo/internal/ragtypes"
)

func ChunkDocument(doc ragtypes.Document, chunkSize int, overlap int) []ragtypes.Chunk {
	text := normalizeText(doc.Text)
	runes := []rune(text)  // 处理 Unicode 字符，确保按字符切分而不是字节

	if chunkSize <= 0 {
		chunkSize = 900
	}

	if overlap < 0 {
		overlap = 0
	}

	if overlap >= chunkSize {
		overlap = chunkSize / 5
	}

	step := chunkSize - overlap

	var chunks []ragtypes.Chunk

	for start := 0; start < len(runes); start += step {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunkText := strings.TrimSpace(string(runes[start:end]))
		if chunkText == "" {
			continue
		}

		index := len(chunks)
		chunks = append(chunks, ragtypes.Chunk{
			ID:     fmt.Sprintf("%s#chunk-%d", doc.Source, index),
			Source: doc.Source,
			Index:  index,
			Text:   chunkText,
		})

		if end >= len(runes) {
			break
		}
	}

	return chunks
}

func ChunkDocuments(docs []ragtypes.Document, chunkSize int, overlap int) []ragtypes.Chunk {
	var all []ragtypes.Chunk

	for _, doc := range docs {
		chunks := ChunkDocument(doc, chunkSize, overlap)
		all = append(all, chunks...)
	}

	return all
}

func normalizeText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}