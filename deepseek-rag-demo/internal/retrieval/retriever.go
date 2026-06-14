package retrieval

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"deepseek-rag-demo/internal/ragtypes"
)

func Retrieve(query string, queryEmbedding []float64, chunks []ragtypes.Chunk, topK int) []ragtypes.ScoredChunk {
	if topK <= 0 {
		topK = 5
	}

	scored := make([]ragtypes.ScoredChunk, 0, len(chunks))

	for _, chunk := range chunks {
		vectorScore := CosineSimilarity(queryEmbedding, chunk.Embedding)
		lexicalScore := LexicalScore(query, chunk.Text)

		finalScore := 0.85*vectorScore + 0.15*lexicalScore

		scored = append(scored, ragtypes.ScoredChunk{
			Chunk:        chunk,
			VectorScore:  vectorScore,
			LexicalScore: lexicalScore,
			FinalScore:   finalScore,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].FinalScore > scored[j].FinalScore
	})

	if len(scored) > topK {
		scored = scored[:topK]
	}

	return scored
}

func CosineSimilarity(a []float64, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}

	var dot float64
	var normA float64
	var normB float64

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func LexicalScore(query string, text string) float64 {
	query = normalizeForLexical(query)
	text = normalizeForLexical(text)

	if query == "" || text == "" {
		return 0
	}

	grams := charNgrams(query, 2)
	if len(grams) == 0 {
		if strings.Contains(text, query) {
			return 1
		}
		return 0
	}

	hit := 0
	for _, gram := range grams {
		if strings.Contains(text, gram) {
			hit++
		}
	}

	return float64(hit) / float64(len(grams))
}

func normalizeForLexical(s string) string {
	var builder strings.Builder

	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		if unicode.IsPunct(r) {
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