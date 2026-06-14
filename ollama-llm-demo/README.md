# Ollama LLM Demo

使用 Go 标准库调用 Ollama `/api/chat` 接口的渐进式示例，覆盖普通对话、多轮上下文、流式输出和结构化 JSON 输出。

## 环境要求

- Go 1.24 或更高版本
- 已安装并启动 Ollama
- 本地已准备代码中配置的模型

当前示例默认使用 `qwen3.6:27b`，Ollama 服务地址为 `http://localhost:11434`。如果本机使用其他模型，请修改对应示例中的 `modelName`。

```powershell
ollama pull qwen3.6:27b
ollama list
```

## 示例说明

| 目录 | 内容 |
| --- | --- |
| 根目录 | 直接调用 `/api/chat` 并流式打印响应 |
| `demo/` | 将 Ollama HTTP 调用封装为可复用客户端 |
| `demo01/` | 最基础的非流式单次对话 |
| `demo02/` | 从终端循环读取问题的单轮问答 |
| `demo03/` | 使用 system message 约束回答风格 |
| `demo04/` | 保存 user/assistant 历史的多轮对话 |
| `demo05/` | 逐块读取并打印流式响应 |
| `demo06/` | 使用 `format: "json"` 获取 JSON 输出 |
| `demo07/` | 使用 JSON Schema 约束结构化输出 |

## 运行

进入项目目录：

```powershell
cd ollama-llm-demo
```

运行根目录流式示例：

```powershell
go run .
```

运行指定示例：

```powershell
go run ./demo01
go run ./demo04
go run ./demo07
```

`demo/` 目录包含多个 Go 文件，需要以目录方式运行：

```powershell
go run ./demo
```

交互式示例中输入 `exit` 可以退出。

## 学习顺序

建议依次阅读 `demo01` 到 `demo07`。这些示例从请求结构、HTTP 调用和响应解析开始，逐步加入 system message、对话历史、流式响应以及结构化输出。

