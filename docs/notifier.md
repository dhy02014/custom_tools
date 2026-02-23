# notifier.go — Telegram Notification

Sends terraform execution results to Telegram using the Bot API with MarkdownV2 formatting. No external SDK required — built entirely with Go standard library.

## Packages Used

| Package | Function / Type | Purpose |
|---------|----------------|---------|
| `net/http` | `http.Post(url, contentType, body)` | Send HTTPS POST to Telegram API (TLS built-in) |
| | `http.StatusOK` | HTTP 200 constant for response validation |
| | `resp.StatusCode` | Response HTTP status code |
| | `resp.Body.Close()` | Release response body resources (must defer) |
| `encoding/json` | `json.Marshal(v)` | Convert Go struct to JSON `[]byte` |
| | struct tag `` `json:"..."` `` | Map Go field names to JSON keys (PascalCase → snake_case) |
| `fmt` | `fmt.Sprintf()` | Compose API URL and format duration string |
| | `fmt.Errorf()` | Create wrapped errors with `%w` for error chaining |
| `strings` | `strings.Builder` | Efficiently assemble message text (O(n) vs O(n²) concatenation) |
| | `strings.NewReader(s)` | Convert JSON string to `io.Reader` for HTTP body |
| | `strings.ReplaceAll()` | Escape 19 MarkdownV2 special characters |
| `time` | `time.Duration` | Duration type (notifyPayload field) |
| | `Duration.Seconds()` | Convert to float64 seconds → "12.3s" format |

## Telegram API Call Flow

```
1. Compose Go struct
   telegramMessage{
     ChatID:    "-100xxx"
     Text:      "✅ Terraform plan succeeded..."
     ParseMode: "MarkdownV2"
   }
        │
        ▼
2. JSON serialization (encoding/json)
   json.Marshal(msg)
   → {"chat_id":"-100xxx","text":"...","parse_mode":"MarkdownV2"}
        │
        ▼
3. HTTPS POST (net/http)
   http.Post(
     "https://api.telegram.org/bot<TOKEN>/sendMessage",
     "application/json",
     body,
   )
        │
        ▼
4. Validate response
   resp.StatusCode == 200 → success
   resp.Body.Close()      → release resources
```

### What http.Post does internally

```
http.Post(url, "application/json", body)
     │
     ├─ 1. Parse URL (scheme: https, host: api.telegram.org)
     ├─ 2. DNS lookup (api.telegram.org → IP)
     ├─ 3. TCP connection (3-way handshake)
     ├─ 4. TLS handshake (automatic — built into stdlib)
     ├─ 5. Send HTTP POST (Content-Type: application/json)
     └─ 6. Receive response (StatusCode + Body)
```

### Why `defer resp.Body.Close()` is required

The response Body is a stream connected to a network socket. Without closing it, TCP connections remain open, causing file descriptor leaks. `defer` guarantees the call executes when the function returns.

## JSON Serialization — Struct Tags

```
Go struct                              JSON output
─────────                              ───────────
telegramMessage{                       {
  ChatID:    "-100xxx",       json       "chat_id": "-100xxx",
  Text:      "✅ ...",       ────>      "text": "✅ ...",
  ParseMode: "MarkdownV2",              "parse_mode": "MarkdownV2"
}                                      }

Struct tags map Go PascalCase to API snake_case:
  ChatID    `json:"chat_id"`
  ParseMode `json:"parse_mode"`
```

## MarkdownV2 Escaping

Telegram MarkdownV2 requires escaping 19 special characters with `\`.

```
Normal text (escape function):
  Characters: \ _ * [ ] ( ) ~ ` > # + - = | { } . !
  Example: "Directory: crm_dev" → "Directory: crm\_dev"
  Example: "Duration: 12.3s"   → "Duration: 12\.3s"

Code blocks (escapeCodeBlock function):
  Characters: \ ` (only 2)
  All other characters are safe inside ``` blocks
```

## Error Chaining with fmt.Errorf %w

```go
fmt.Errorf("marshal message: %w", err)
//                            ~~
// %w wraps the original error, enabling errors.Is() / errors.As()
// → produces: "marshal message: json: unsupported type"
```
