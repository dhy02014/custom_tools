# notifier.go — Telegram Notification

Telegram Bot API를 호출하여 terraform 실행 결과를 MarkdownV2 메시지로 전송합니다.
외부 SDK 없이 Go stdlib만 사용합니다.

## 사용 패키지

| 패키지 | 사용된 함수/타입 | 용도 |
|--------|-----------------|------|
| `net/http` | `http.Post(url, contentType, body)` | Telegram API에 HTTPS POST 전송 (TLS 내장) |
| | `http.StatusOK` | HTTP 200 상수 (응답 성공 여부 판별) |
| | `resp.StatusCode` | 응답 HTTP 상태 코드 |
| | `resp.Body.Close()` | 응답 바디 리소스 해제 (defer 필수) |
| `encoding/json` | `json.Marshal(v)` | Go struct → JSON `[]byte` 변환 |
| | struct tag `` `json:"..."` `` | JSON 필드명 매핑 (PascalCase → snake_case) |
| `fmt` | `fmt.Sprintf()` | API URL 조합, 소요시간 포맷팅 |
| | `fmt.Errorf()` | `%w`로 에러 체이닝 (원본 에러 추적 가능) |
| `strings` | `strings.Builder` | 메시지를 효율적으로 조립 (O(n)) |
| | `strings.NewReader(s)` | JSON 문자열을 `io.Reader`로 변환 (HTTP body) |
| | `strings.ReplaceAll()` | MarkdownV2 특수문자 19종 이스케이프 |
| `time` | `time.Duration` | 소요시간 타입 (notifyPayload 필드) |
| | `Duration.Seconds()` | float64 초 단위 변환 → "12.3s" 포맷 |

## Telegram API 호출 흐름

```
  1. Go struct 구성
     telegramMessage{
       ChatID:    "-100xxx"
       Text:      "✅ Terraform plan 성공..."
       ParseMode: "MarkdownV2"
     }
          │
          ▼
  2. JSON 직렬화 (encoding/json)
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
  4. 응답 확인
     resp.StatusCode == 200 → 성공
     resp.Body.Close()      → 리소스 해제
```

### http.Post 내부 동작

```
http.Post(url, "application/json", body)
     │
     ├─ 1. URL 파싱 (scheme: https, host: api.telegram.org)
     ├─ 2. DNS 조회 (api.telegram.org → IP)
     ├─ 3. TCP 연결 (3-way handshake)
     ├─ 4. TLS 핸드셰이크 (자동 — stdlib 내장, 별도 설정 불필요)
     ├─ 5. HTTP POST 전송 (Content-Type: application/json)
     └─ 6. 응답 수신 (StatusCode + Body)
```

### defer resp.Body.Close()가 필요한 이유

HTTP 응답 Body는 네트워크 소켓에 연결된 스트림입니다.
Close() 하지 않으면 TCP 연결이 열린 채 남아 file descriptor가 누수됩니다.
`defer`는 함수 종료 시 자동 호출을 보장합니다.

## JSON 직렬화 — struct tag

```
Go struct                              JSON output
─────────                              ───────────
telegramMessage{                       {
  ChatID:    "-100xxx",       json       "chat_id": "-100xxx",
  Text:      "✅ ...",       ────>      "text": "✅ ...",
  ParseMode: "MarkdownV2",              "parse_mode": "MarkdownV2"
}                                      }

struct tag가 필드명을 변환:
  ChatID    `json:"chat_id"`     Go PascalCase → API snake_case
  ParseMode `json:"parse_mode"`
```

## MarkdownV2 이스케이프

Telegram MarkdownV2는 19개 특수문자를 `\`로 이스케이프해야 합니다.

```
일반 텍스트 (escape 함수):
  대상 문자: \ _ * [ ] ( ) ~ ` > # + - = | { } . !
  예: "📁 디렉토리: crm_dev" → "📁 디렉토리: crm\_dev"
  예: "⏱ 소요시간: 12.3s"   → "⏱ 소요시간: 12\.3s"

코드 블록 (escapeCodeBlock 함수):
  대상 문자: \ ` (2개만)
  나머지는 코드 블록 안에서 이스케이프 불필요
```

## 메시지 포맷

### 성공 시

```
✅ Terraform plan 성공
━━━━━━━━━━━━━━━━━━━━
📁 디렉토리: crm_dev
⏱ 소요시간: 12.3s
💻 명령어: terraform plan -chdir=...
```

### 실패 시

```
❌ Terraform validate 실패
━━━━━━━━━━━━━━━━━━━━
📁 디렉토리: crm_dev
⏱ 소요시간: 0.8s
💻 명령어: terraform validate

📋 에러 요약:
  Error: Missing required argument...
  (stderr 마지막 10줄)
```

## fmt.Errorf %w — 에러 체이닝

```go
fmt.Errorf("marshal message: %w", err)
//                            ~~
// %w는 원본 에러를 감싸서 errors.Is(), errors.As()로 추적 가능
// → "marshal message: json: unsupported type" 같은 체인 출력
```
