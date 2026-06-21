
---

## 3. 文件：`skills/go-agent-tutor/docs/roadmap.md`

```markdown
# Go Agent 学习路线

## 第 1 阶段：LLM API 基础

目标：

- 理解 system prompt / user prompt / assistant message
- 理解 token、上下文窗口、temperature、top_p
- 学会用 Go 调用模型 API
- 学会 streaming 和 JSON schema 输出

完成标准：

- 能写出用户输入 -> 调模型 -> 打印回复的程序
- 能让模型输出结构化 JSON

## 第 2 阶段：Tool Calling

目标：

- 理解工具是 Go 函数 + 参数 schema + 描述
- 理解模型只决定调用工具，不真正执行工具
- 理解 tool_calls 和 tool result

完成标准：

- 能写 calculator、current_time 等工具
- 能根据 tool name dispatch 到本地函数

## 第 3 阶段：Agent Loop

目标：

- 理解 LLM -> Tool -> Observation -> LLM 的循环
- 理解 maxSteps 防止死循环
- 理解 assistant tool_call 消息和 tool result 消息都要进入上下文

完成标准：

- 能手写一个最小 Agent Loop

## 第 4 阶段：RAG / Knowledge

目标：

- 理解 chunk、embedding、向量检索、TopK、rerank、引用来源
- 能把私有论文或文档做成知识库

完成标准：

- 能完成 用户问题 -> 检索文档 -> 拼 prompt -> 模型回答

## 第 5 阶段：Memory / Session

目标：

- 区分短期 Session 和长期 Memory
- 理解 Memory 不能随便写入
- 支持 remember / forget / session clear

完成标准：

- 能用 JSON 文件保存 Session 和 Memory

## 第 6 阶段：Workflow / Multi-Agent

目标：

- 理解 Planner / Coder / Reviewer / Executor
- 理解 chain / parallel / cycle / graph
- 理解 State 在工作流节点之间流动

完成标准：

- 能写一个 Planner -> Coder -> Test -> Reviewer 的最小工作流

## 第 7 阶段：MCP

目标：

- 理解 MCP 是 Agent 连接外部工具服务的协议
- 区分 MCP 和 Tool Calling
- 能写 STDIO MCP Server 并接入 trpc-agent-go

完成标准：

- 能让 DeepSeek 通过 trpc-agent-go 调 MCP 工具

## 第 8 阶段：Agent Skills

目标：

- 理解 Skill 是可复用工作流能力包
- 能写 SKILL.md、docs、scripts
- 能通过 skill_load / skill_run 使用 Skill

完成标准：

- 能创建一个自己的 Skill，并让 Agent 按 Skill 指南完成任务