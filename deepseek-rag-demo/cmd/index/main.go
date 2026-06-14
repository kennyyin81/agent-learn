package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"deepseek-rag-demo/internal/chunker"
	"deepseek-rag-demo/internal/config"
	"deepseek-rag-demo/internal/document"
	"deepseek-rag-demo/internal/embedding"
	"deepseek-rag-demo/internal/vectorstore"
)

func main() {
	dataDir := flag.String("data", "data", "知识库文本目录")
	storePath := flag.String("store", "store/index.json", "向量库保存路径")
	chunkSize := flag.Int("chunk-size", 900, "每个 chunk 的字符数")
	overlap := flag.Int("overlap", 150, "chunk 重叠字符数")
	flag.Parse()

	embedURL := config.Env("OLLAMA_EMBED_URL", "http://localhost:11434/api/embeddings")
	embedModel := config.Env("OLLAMA_EMBED_MODEL", "bge-m3")

	fmt.Println("开始建立 RAG 索引")
	fmt.Println("dataDir:", *dataDir)
	fmt.Println("storePath:", *storePath)
	fmt.Println("embedding model:", embedModel)

	docs, err := document.LoadDocuments(*dataDir)
	if err != nil {
		log.Fatal(err)
	}

	if len(docs) == 0 {
		log.Fatalf("没有在 %s 中找到 .txt 或 .md 文件", *dataDir)
	}

	fmt.Printf("读取到 %d 个文档\n", len(docs))

	chunks := chunker.ChunkDocuments(docs, *chunkSize, *overlap)
	if len(chunks) == 0 {
		log.Fatal("切分后没有生成 chunks")
	}

	fmt.Printf("生成 %d 个 chunks\n", len(chunks))

	embedClient := embedding.NewOllamaClient(embedURL, embedModel)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Minute)
	defer cancel()

	for i := range chunks {
		fmt.Printf("Embedding chunk %d/%d: %s\n", i+1, len(chunks), chunks[i].ID)

		embeddingVector, err := embedClient.Embed(ctx, chunks[i].Text)
		if err != nil {
			log.Fatalf("生成 embedding 失败，chunk=%s, err=%v", chunks[i].ID, err)
		}

		chunks[i].Embedding = embeddingVector
	}

	store := vectorstore.Store{
		EmbeddingModel: embedModel,
		Chunks:         chunks,
	}

	if err := vectorstore.Save(*storePath, store); err != nil {
		log.Fatal(err)
	}

	fmt.Println("索引建立完成:", *storePath)
}