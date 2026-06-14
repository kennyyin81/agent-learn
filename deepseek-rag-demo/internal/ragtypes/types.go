package ragtypes

// 原始文档
type Document struct {
	Source string  // 路径
	Text   string  // 完整文本内容
}

// 切分后的文本块，用于计算相似度
type Chunk struct {
	ID        string    `json:"id"`  // 唯一标识
	Source    string    `json:"source"`  // 来源路径
	Index     int       `json:"index"`  // 第几个chunk
	Text      string    `json:"text"`  // 对应chunk的文本内容
	Embedding []float64 `json:"embedding"`  // 语意向量
}

// 检索后的排序结果
type ScoredChunk struct {
	Chunk        Chunk
	VectorScore  float64  // 语意相似度得分
	LexicalScore float64  // 关键词分数
	FinalScore   float64  // 最终得分
}