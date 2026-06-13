package localtools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ollama-tool-demo/internal/llm"
)

type CurrentTimeArgs struct {
	Timezone string `json:"timezone"`
}

func CurrentTimeSchema() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.ToolFunction{
			Name:        "get_current_time",
			Description: "Get the current time for a timezone. Prefer IANA timezone names such as Asia/Singapore, Asia/Shanghai, Asia/Tokyo, or UTC.",
			Parameters: map[string]any{
				"type":     "object",
				"required": []string{"timezone"},
				"properties": map[string]any{
					"timezone": map[string]any{
						"type":        "string",
						"description": "IANA timezone name, for example Asia/Singapore or UTC",
					},
				},
			},
		},
	}
}

func RunCurrentTime(args json.RawMessage) (string, error) {
	var input CurrentTimeArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("get_current_time 参数解析失败: %w", err)
	}

	timezone := strings.TrimSpace(input.Timezone)
	if timezone == "" {
		timezone = "Asia/Singapore"
	}

	aliases := map[string]string{
		"Singapore": "Asia/Singapore",
		"新加坡":       "Asia/Singapore",
		"China":     "Asia/Shanghai",
		"中国":        "Asia/Shanghai",
		"Shanghai":  "Asia/Shanghai",
		"北京":        "Asia/Shanghai",
		"Beijing":   "Asia/Shanghai",
		"Tokyo":     "Asia/Tokyo",
		"东京":        "Asia/Tokyo",
	}

	if realTimezone, ok := aliases[timezone]; ok {
		timezone = realTimezone
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("不支持的 timezone: %s，请使用 IANA 时区名，例如 Asia/Singapore 或 UTC", timezone)
	}

	now := time.Now().In(loc)

	output := map[string]string{
		"timezone": timezone,
		"datetime": now.Format("2006-01-02 15:04:05 MST"),
		"rfc3339":  now.Format(time.RFC3339),
	}

	outputBytes, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("get_current_time 结果序列化失败: %w", err)
	}

	return string(outputBytes), nil
}