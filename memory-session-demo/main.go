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

	"memory-session-demo/internal/llm"
	"memory-session-demo/internal/memory"
	"memory-session-demo/internal/prompt"
	"memory-session-demo/internal/session"
)

func main() {
	ollamaURL := flag.String("ollama-url", env("OLLAMA_CHAT_URL", "http://localhost:11434/api/chat"), "Ollama chat API 地址")
	model := flag.String("model", env("OLLAMA_MODEL", "qwen3.6:27b"), "Ollama 模型名")
	sessionPath := flag.String("session", "runtime/sessions/default.json", "Session 保存路径")
	memoryPath := flag.String("memory", "runtime/memory/memories.json", "Memory 保存路径")
	historyLimit := flag.Int("history", 12, "每次发送给模型的最近消息数量")
	memoryLimit := flag.Int("memory-k", 5, "每次注入 prompt 的相关 memory 数量")
	flag.Parse()

	client := llm.NewOllamaClient(*ollamaURL, *model)

	sess, err := session.LoadOrNew(*sessionPath, "default")
	if err != nil {
		log.Fatal(err)
	}

	memStore, err := memory.LoadOrNew(*memoryPath)
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)

	printWelcome(*model, *sessionPath, *memoryPath, len(sess.Messages), len(memStore.Items))

	for {
		fmt.Print("\n你> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("已退出。")
			return
		}

		handled, err := handleCommand(input, sess, memStore, *sessionPath, *memoryPath)
		if err != nil {
			fmt.Println("命令执行失败:", err)
			continue
		}
		if handled {
			continue
		}

		relevantMemories := memStore.Search(input, *memoryLimit)
		recentHistory := sess.Recent(*historyLimit)

		messages := prompt.BuildMessages(prompt.BuildInput{
			UserInput:      input,
			RecentHistory:  recentHistory,
			RelevantMemory: relevantMemories,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)

		fmt.Println("正在调用模型...")

		answer, err := client.Chat(ctx, messages)
		cancel()

		if err != nil {
			fmt.Println("调用模型失败:", err)
			continue
		}

		fmt.Println("\n助手>", answer)

		sess.Add("user", input)
		sess.Add("assistant", answer)

		if err := sess.Save(*sessionPath); err != nil {
			fmt.Println("保存 session 失败:", err)
		}
	}
}

func handleCommand(
	input string,
	sess *session.Session,
	memStore *memory.Store,
	sessionPath string,
	memoryPath string,
) (bool, error) {
	if input == "/help" {
		printHelp()
		return true, nil
	}

	if input == "/memories" {
		items := memStore.List()
		if len(items) == 0 {
			fmt.Println("当前没有长期记忆。")
			return true, nil
		}

		fmt.Println("长期记忆列表：")
		for _, item := range items {
			fmt.Printf("- id=%s type=%s source=%s text=%s\n",
				item.ID,
				item.Type,
				item.Source,
				item.Text,
			)
		}
		return true, nil
	}

	if strings.HasPrefix(input, "/remember ") {
		text := strings.TrimSpace(strings.TrimPrefix(input, "/remember "))
		item, err := memStore.Add(text, "user_preference_or_background", "manual_command")
		if err != nil {
			return true, err
		}

		if err := memStore.Save(memoryPath); err != nil {
			return true, err
		}

		fmt.Println("已写入长期记忆：")
		fmt.Printf("id=%s text=%s\n", item.ID, item.Text)
		return true, nil
	}

	if strings.HasPrefix(input, "/forget ") {
		id := strings.TrimSpace(strings.TrimPrefix(input, "/forget "))
		if id == "" {
			fmt.Println("用法：/forget mem_xxx")
			return true, nil
		}

		ok := memStore.Delete(id)
		if !ok {
			fmt.Println("没有找到该 memory id:", id)
			return true, nil
		}

		if err := memStore.Save(memoryPath); err != nil {
			return true, err
		}

		fmt.Println("已删除 memory:", id)
		return true, nil
	}

	if input == "/clear-session" {
		sess.Clear()
		if err := sess.Save(sessionPath); err != nil {
			return true, err
		}
		fmt.Println("当前 Session 已清空。长期 Memory 不受影响。")
		return true, nil
	}

	if input == "/session" {
		fmt.Printf("Session ID: %s\n", sess.ID)
		fmt.Printf("消息数量: %d\n", len(sess.Messages))
		fmt.Printf("创建时间: %s\n", sess.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("更新时间: %s\n", sess.UpdatedAt.Format("2006-01-02 15:04:05"))
		return true, nil
	}

	return false, nil
}

func printWelcome(model string, sessionPath string, memoryPath string, sessionMessages int, memoryCount int) {
	fmt.Println("Memory + Session Demo 已启动")
	fmt.Println("模型:", model)
	fmt.Println("Session 文件:", sessionPath)
	fmt.Println("Memory 文件:", memoryPath)
	fmt.Println("当前 Session 消息数:", sessionMessages)
	fmt.Println("当前 Memory 数量:", memoryCount)
	fmt.Println()
	printHelp()
}

func printHelp() {
	fmt.Println("可用命令：")
	fmt.Println("  /remember 内容       写入长期记忆")
	fmt.Println("  /memories            查看长期记忆")
	fmt.Println("  /forget mem_xxx      删除指定长期记忆")
	fmt.Println("  /session             查看当前 Session 状态")
	fmt.Println("  /clear-session       清空当前 Session")
	fmt.Println("  /help                查看帮助")
	fmt.Println("  exit / quit          退出")
	fmt.Println()
	fmt.Println("普通输入会进入聊天，并自动写入当前 Session。")
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}