# config.go — Configuration Loader

Loads Telegram credentials from environment variables or the `~/.tfn.env` file.

## Packages Used

| Package | Function / Type | Purpose |
|---------|----------------|---------|
| `os` | `os.Getenv(key)` | Read TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID from environment |
| | `os.UserHomeDir()` | Resolve home directory for `~/.tfn.env` path |
| | `os.Open(path)` | Open `~/.tfn.env` for reading |
| `bufio` | `bufio.NewScanner(r)` | Wrap `io.Reader` as a line-by-line Scanner |
| | `scanner.Scan()` | Read next line (returns false at EOF) |
| | `scanner.Text()` | Get current line as string (newline excluded) |
| `path/filepath` | `filepath.Join(parts...)` | Join path components with OS-appropriate separator |
| `strings` | `strings.TrimSpace()` | Trim leading/trailing whitespace from each line |
| | `strings.Cut(s, sep)` | Split "KEY=VALUE" into key and value (Go 1.18+) |
| | `strings.Trim(s, cutset)` | Strip surrounding quotes (`"'`) from config values |

## Config Loading Priority

```
  ┌────────────────────────────────────┐
  │ 1. Environment Variables (highest) │
  │    TELEGRAM_BOT_TOKEN              │
  │    TELEGRAM_CHAT_ID                │
  └──────────────┬─────────────────────┘
                 │ missing?
                 ▼
  ┌────────────────────────────────────┐
  │ 2. ~/.tfn.env (config file)        │
  │    Format: KEY=VALUE (per line)    │
  └──────────────┬─────────────────────┘
                 │ missing?
                 ▼
  ┌────────────────────────────────────┐
  │ 3. No config → skip notification   │
  │    terraform runs normally         │
  │    (graceful degradation)          │
  └────────────────────────────────────┘
```

## .env File Parsing Flow

```
File contents:                    Scanner behavior:
┌──────────────────────────┐
│ # Telegram credentials   │ → Scan() → Text() → starts with '#' → skip
│ TELEGRAM_BOT_TOKEN=abc   │ → Scan() → Text() → Cut("=") → store key/value
│ TELEGRAM_CHAT_ID=-100    │ → Scan() → Text() → Cut("=") → store key/value
│                          │ → Scan() → empty line → skip
└──────────────────────────┘ → Scan() → false (EOF) → exit loop
```

## filepath.Join — Cross-Platform Path Construction

```go
// Bad — OS-dependent
path := home + "/.tfn.env"          // only works on Linux/macOS
path := home + "\\.tfn.env"         // only works on Windows

// Good — OS-aware
path := filepath.Join(home, ".tfn.env")
// macOS/Linux → /Users/khko/.tfn.env
// Windows    → C:\Users\khko\.tfn.env
```

## strings.Cut — Core of Config Parsing

```
Input: "TELEGRAM_BOT_TOKEN=123:ABC"
strings.Cut(line, "=")
→ key="TELEGRAM_BOT_TOKEN", value="123:ABC", ok=true

Input: "# comment line"
strings.Cut(line, "=")
→ ok=false → skip
```

## bufio.Scanner — Streaming Line Reader

```go
// Bad — loads entire file into memory
data, _ := os.ReadFile(path)
lines := strings.Split(string(data), "\n")

// Good — streams line by line (memory-efficient)
scanner := bufio.NewScanner(f)
for scanner.Scan() {
    line := scanner.Text()
}
// → uses only one line's worth of memory regardless of file size
```
