package main

import (
	"context"
	"fmt"
	"log"

	mcp "trpc.group/trpc-go/trpc-mcp-go"
)

func main() {
	server := mcp.NewStdioServer(
		"study-mcp-server",
		"1.0.0",
		mcp.WithStdioServerLogger(mcp.GetDefaultLogger()),
	)

	echoTool := mcp.NewTool(
		"echo",
		mcp.WithDescription("Echo back a message with an optional prefix."),
		mcp.WithString("message", mcp.Required(), mcp.Description("The message to echo.")),
		mcp.WithString("prefix", mcp.Description("Optional prefix. Default is 'Echo: '.")),
	)
	server.RegisterTool(echoTool, handleEcho)

	addTool := mcp.NewTool(
		"add",
		mcp.WithDescription("Add two numbers and return the result."),
		mcp.WithNumber("a", mcp.Required(), mcp.Description("First number.")),
		mcp.WithNumber("b", mcp.Required(), mcp.Description("Second number.")),
	)
	server.RegisterTool(addTool, handleAdd)

	readNoteTool := mcp.NewTool(
		"read_note",
		mcp.WithDescription("Read a built-in study note by topic. Supported topics: mcp, rag, agent."),
		mcp.WithString("topic", mcp.Required(), mcp.Description("Topic name: mcp, rag, or agent.")),
	)
	server.RegisterTool(readNoteTool, handleReadNote)

	log.Println("Starting study MCP STDIO server...")
	log.Println("Available tools: echo, add, read_note")

	if err := server.Start(); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func handleEcho(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message, _ := req.Params.Arguments["message"].(string)
	if message == "" {
		return nil, fmt.Errorf("missing required parameter: message")
	}

	prefix, _ := req.Params.Arguments["prefix"].(string)
	if prefix == "" {
		prefix = "Echo: "
	}

	return mcp.NewTextResult(prefix + message), nil
}

func handleAdd(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a, err := getNumber(req, "a")
	if err != nil {
		return nil, err
	}

	b, err := getNumber(req, "b")
	if err != nil {
		return nil, err
	}

	result := a + b
	return mcp.NewTextResult(fmt.Sprintf("%.2f + %.2f = %.2f", a, b, result)), nil
}

func handleReadNote(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	topic, _ := req.Params.Arguments["topic"].(string)
	if topic == "" {
		return nil, fmt.Errorf("missing required parameter: topic")
	}

	notes := map[string]string{
		"mcp":   "MCP 是 Model Context Protocol，用于把外部工具、资源和 prompt 以标准协议暴露给 AI 应用。",
		"rag":   "RAG 是 Retrieval-Augmented Generation，核心流程是先检索相关资料，再把资料拼入 prompt，让模型基于资料回答。",
		"agent": "Agent 的核心是 LLM + Tools + Loop + State。模型决定下一步，程序执行工具，再把结果交回模型。",
	}

	text, ok := notes[topic]
	if !ok {
		return nil, fmt.Errorf("unknown topic: %s, supported topics: mcp, rag, agent", topic)
	}

	return mcp.NewTextResult(text), nil
}

func getNumber(req *mcp.CallToolRequest, name string) (float64, error) {
	value, ok := req.Params.Arguments[name]
	if !ok {
		return 0, fmt.Errorf("missing required parameter: %s", name)
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("invalid parameter %s: must be a number", name)
	}
}
