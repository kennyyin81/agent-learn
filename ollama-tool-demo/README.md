# Ollama Tool Demo

一个使用 Go 和 Ollama 实现的本地 Tool Calling Agent。模型可以在 Agent Loop 中选择工具、读取工具结果，并继续推理直到生成最终回答。

## 功能

- 支持普通和流式对话
- 支持多轮工具调用
- 工具参数使用 JSON Schema 描述
- 工具执行结果以 `tool` 消息返回给模型
- 最多执行 8 个 Agent Step，避免无限循环
- 写文件前要求用户输入 `yes` 确认
- 文件访问被限制在项目的 `workspace/` 目录内

## 内置工具

| 工具 | 功能 | 是否确认 |
| --- | --- | --- |
| `get_current_time` | 查询指定 IANA 时区的当前时间 | 否 |
| `calculator` | 加、减、乘、除双数运算 | 否 |
| `random_number` | 生成指定闭区间内的随机整数 | 否 |
| `read_text_file` | 读取 `workspace/` 下不超过 1 MiB 的文本文件 | 否 |
| `write_text_file` | 在 `workspace/` 下写入 UTF-8 文本文件 | 是 |

文件工具只接受相对路径，并会拒绝绝对路径和逃逸到 `workspace/` 之外的路径。

## 环境要求

- Go 1.24 或更高版本
- 已安装并启动 Ollama
- 使用支持 Tool Calling 的本地模型

当前代码默认使用 `qwen3.6:27b`，Ollama 服务地址为 `http://localhost:11434/api/chat`。如需更换模型，请修改 `main.go` 中的 `modelName`。

```powershell
ollama pull qwen3.6:27b
ollama list
```

## 运行

```powershell
cd ollama-tool-demo
go run .
```

开启流式输出：

```powershell
go run . -stream
```

输入 `exit` 或 `quit` 退出。

## 对话示例

```text
生成一个 1 到 100 的随机整数
查询 Asia/Shanghai 当前时间
计算 15 * 23 + 100
读取 note.txt 并总结
写入 hello.txt，内容是：你好，Ollama Agent
```

读写文件前，可以手动创建工作目录：

```powershell
New-Item -ItemType Directory -Force workspace
```

写文件工具会自动创建目标文件的父目录，但执行前必须在终端输入 `yes`。

## 项目结构

```text
.
├── main.go
├── internal/
│   ├── agent/          # Agent Loop 与写操作确认
│   ├── llm/            # Ollama 请求、响应及流式解析
│   └── localtools/     # 工具 Schema、安全检查和执行逻辑
└── workspace/          # 文件工具允许访问的本地目录
```

