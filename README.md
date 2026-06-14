# Agent Learn

使用 Go 学习大模型应用开发的示例仓库，包含 Ollama 本地对话、Tool Calling Agent 和 DeepSeek RAG。

## 项目

| 项目 | 内容 |
| --- | --- |
| [`ollama-llm-demo`](./ollama-llm-demo/) | 从基础 HTTP 调用到多轮对话、流式输出和 JSON Schema |
| [`ollama-tool-demo`](./ollama-tool-demo/) | 带本地工具、安全目录和人工确认的 Ollama Agent |
| [`deepseek-rag-demo`](./deepseek-rag-demo/) | 使用 Ollama Embedding、混合检索和 DeepSeek 生成回答 |

## 建议学习顺序

1. 从 `ollama-llm-demo` 了解 Ollama `/api/chat` 的请求和响应格式。
2. 通过 `ollama-tool-demo` 学习工具 Schema、Tool Calling 和 Agent Loop。
3. 通过 `deepseek-rag-demo` 学习文档分块、向量索引、检索和上下文增强生成。

## 环境要求

- Go 1.24 或更高版本
- Ollama
- 各示例所需的本地模型
- DeepSeek RAG 示例额外需要 DeepSeek API Key

每个项目都是独立的 Go Module。请进入对应目录运行，并参考子项目 README 完成模型及环境配置。

```powershell
cd ollama-llm-demo
go run .
```

## 仓库说明

- 示例默认连接本机 `http://localhost:11434` 的 Ollama 服务。
- 模型名称直接配置在各示例代码中，可以替换为本机已有模型。
- RAG 使用的本地知识文档及生成索引不会提交到 Git。
