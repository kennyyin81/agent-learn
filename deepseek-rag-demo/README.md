# DeepSeek RAG Demo

这是一个使用 Go 实现的本地 RAG 示例项目：

- 使用 Ollama 和 `bge-m3` 生成文档及问题向量
- 将向量索引保存到本地 JSON 文件
- 结合向量检索与关键词检索查找相关片段
- 调用 DeepSeek API，根据检索上下文生成回答

## 项目结构

```text
.
├── cmd/
│   ├── index/          # 构建知识库索引
│   └── ask/            # 检索并调用 DeepSeek 回答问题
├── data/               # 本地知识文档及相关代码
├── internal/           # 文档加载、分块、向量化、检索和 RAG 逻辑
├── store/              # 本地生成的向量索引
├── .env.example        # 环境变量示例
└── go.mod
```

## 环境要求

- Go 1.24 或更高版本
- 已安装并启动 Ollama
- Ollama 中已下载 `bge-m3` 模型
- DeepSeek API Key

```powershell
ollama pull bge-m3
```

## 配置

参考 `.env.example` 设置环境变量。程序不会自动加载 `.env`，运行前需要在终端中设置：

```powershell
$env:DEEPSEEK_API_KEY="你的 DeepSeek API Key"
$env:DEEPSEEK_BASE_URL="https://api.deepseek.com"
$env:DEEPSEEK_MODEL="deepseek-v4-flash"
$env:OLLAMA_EMBED_URL="http://localhost:11434/api/embeddings"
$env:OLLAMA_EMBED_MODEL="bge-m3"
```

## 准备知识库

将知识文档放入 `data/` 目录。当前文档加载器支持：

- `.txt`
- `.md`

`data/` 中的知识文档仅供本地使用，已通过 `.gitignore` 排除，不会提交到 Git。目录内的源码文件可以正常提交。当前版本不会解析 `.docx` 文件。

生成的 `store/index.json` 可能包含原始文档片段，因此同样不会提交。

## 构建索引

在项目根目录运行：

```powershell
go run ./cmd/index
```

也可以自定义参数：

```powershell
go run ./cmd/index -data data -store store/index.json -chunk-size 900 -overlap 150
```

## 提问

```powershell
go run ./cmd/ask -q "这篇文档主要讲了什么？"
```

常用参数：

- `-store`：索引文件路径，默认为 `store/index.json`
- `-q`：需要回答的问题
- `-k`：参与回答的相关片段数量，默认为 `5`

如果 `-q` 为空，程序会提示在终端中输入问题。
