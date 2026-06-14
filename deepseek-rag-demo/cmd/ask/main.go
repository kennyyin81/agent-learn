package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"deepseek-rag-demo/internal/config"
	"deepseek-rag-demo/internal/embedding"
	"deepseek-rag-demo/internal/llm"
	"deepseek-rag-demo/internal/rag"
	"deepseek-rag-demo/internal/retrieval"
	"deepseek-rag-demo/internal/vectorstore"
)

func main() {
	storePath := flag.String("store", "store/index.json", "向量库路径")
	question := flag.String("q", "这篇论文有没有研究火星移民?", "用户问题")
	topK := flag.Int("k", 5, "检索返回的 chunk 数量")
	flag.Parse()

	q := strings.TrimSpace(*question)
	if q == "" {
		fmt.Print("请输入问题：")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		q = strings.TrimSpace(line)
	}

	if q == "" {
		log.Fatal("问题不能为空")
	}

	deepseekAPIKey := os.Getenv("DEEPSEEK_API_KEY")
	deepseekBaseURL := config.Env("DEEPSEEK_BASE_URL", "https://api.deepseek.com")
	deepseekModel := config.Env("DEEPSEEK_MODEL", "deepseek-v4-flash")
	embedURL := config.Env("OLLAMA_EMBED_URL", "http://localhost:11434/api/embeddings")

	store, err := vectorstore.Load(*storePath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("已加载向量库:", *storePath)
	fmt.Println("chunks:", len(store.Chunks))
	fmt.Println("embedding model:", store.EmbeddingModel)

	embedClient := embedding.NewOllamaClient(embedURL, store.EmbeddingModel)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fmt.Println("正在向量化问题...")
	queryEmbedding, err := embedClient.Embed(ctx, q)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("正在检索相关论文片段...")
	contexts := retrieval.Retrieve(q, queryEmbedding, store.Chunks, *topK)

	fmt.Println("\n========== 检索结果 ==========")
	for i, item := range contexts {
		fmt.Printf("[S%d] source=%s chunk=%d final=%.4f vector=%.4f lexical=%.4f\n",
			i+1,
			item.Chunk.Source,
			item.Chunk.Index,
			item.FinalScore,
			item.VectorScore,
			item.LexicalScore,
		)

		preview := item.Chunk.Text
		runes := []rune(preview)
		if len(runes) > 120 {
			preview = string(runes[:120]) + "..."
		}
		fmt.Println(preview)
		fmt.Println()
	}

	messages := rag.BuildMessages(q, contexts)

	deepseekClient := llm.NewDeepSeekClient(
		deepseekBaseURL,
		deepseekAPIKey,
		deepseekModel,
	)

	fmt.Println("正在调用 DeepSeek 生成回答...")

	answer, err := deepseekClient.Chat(ctx, messages)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n========== RAG 回答 ==========")
	fmt.Println(answer)
}