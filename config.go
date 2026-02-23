package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// config holds Telegram bot credentials.
type config struct {
	BotToken string
	ChatID   string
}

// isValid returns true if both token and chat ID are set.
func (c config) isValid() bool {
	return c.BotToken != "" && c.ChatID != ""
}

// loadConfig loads configuration from environment variables first,
// then falls back to ~/.tfn.env file.
func loadConfig() config {
	cfg := config{
		BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
	}

	if cfg.isValid() {
		return cfg
	}

	// Fallback: load from ~/.tfn.env
	fileCfg := loadEnvFile()

	if cfg.BotToken == "" {
		cfg.BotToken = fileCfg.BotToken
	}
	if cfg.ChatID == "" {
		cfg.ChatID = fileCfg.ChatID
	}

	return cfg
}

// loadEnvFile reads KEY=VALUE pairs from ~/.tfn.env.
func loadEnvFile() config {
	home, err := os.UserHomeDir()
	if err != nil {
		return config{}
	}

	path := filepath.Join(home, ".tfn.env")
	f, err := os.Open(path)
	if err != nil {
		return config{}
	}
	defer f.Close()

	cfg := config{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || line[0] == '#' {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		// Strip surrounding quotes if present
		value = strings.Trim(value, "\"'")

		switch key {
		case "TELEGRAM_BOT_TOKEN":
			cfg.BotToken = value
		case "TELEGRAM_CHAT_ID":
			cfg.ChatID = value
		}
	}

	return cfg
}
