package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// notifyPayload contains all info needed to compose a notification.
type notifyPayload struct {
	Subcommand string
	Args       []string
	WorkDir    string
	ExitCode   int
	Duration   time.Duration
	Stderr     string
}

// telegramMessage is the JSON body for Telegram sendMessage API.
type telegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// sendNotification sends a Telegram message about the terraform result.
func sendNotification(cfg config, p notifyPayload) error {
	text := formatMessage(p)

	msg := telegramMessage{
		ChatID:    cfg.ChatID,
		Text:      text,
		ParseMode: "MarkdownV2",
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.BotToken)

	resp, err := http.Post(url, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}

// formatMessage composes the MarkdownV2 notification text.
func formatMessage(p notifyPayload) string {
	var b strings.Builder

	// Status emoji and header
	if p.ExitCode == 0 {
		b.WriteString(escape("✅ Terraform " + p.Subcommand + " 성공"))
	} else {
		b.WriteString(escape("❌ Terraform " + p.Subcommand + " 실패"))
	}
	b.WriteString("\n")
	b.WriteString(escape("━━━━━━━━━━━━━━━━━━━━"))
	b.WriteString("\n")

	// Directory
	b.WriteString(escape("📁 디렉토리: " + p.WorkDir))
	b.WriteString("\n")

	// Duration
	b.WriteString(escape(fmt.Sprintf("⏱ 소요시간: %.1fs", p.Duration.Seconds())))
	b.WriteString("\n")

	// Full command
	fullCmd := "terraform " + strings.Join(p.Args, " ")
	b.WriteString(escape("💻 명령어: " + fullCmd))

	// Error summary (only on failure)
	if p.ExitCode != 0 && p.Stderr != "" {
		b.WriteString("\n\n")
		b.WriteString(escape("📋 에러 요약:"))
		b.WriteString("\n")
		b.WriteString("```\n")
		b.WriteString(escapeCodeBlock(p.Stderr))
		b.WriteString("\n```")
	}

	return b.String()
}

// escape escapes special characters for Telegram MarkdownV2.
// See: https://core.telegram.org/bots/api#markdownv2-style
func escape(s string) string {
	specialChars := []string{
		"\\", "_", "*", "[", "]", "(", ")", "~", "`", ">",
		"#", "+", "-", "=", "|", "{", "}", ".", "!",
	}
	result := s
	for _, ch := range specialChars {
		result = strings.ReplaceAll(result, ch, "\\"+ch)
	}
	return result
}

// escapeCodeBlock escapes only characters that need escaping inside ``` blocks.
// Inside pre/code blocks, only \, ` need escaping.
func escapeCodeBlock(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}
