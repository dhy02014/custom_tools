# config.go — Configuration Loader

환경변수와 `~/.tfn.env` 파일에서 Telegram 자격증명을 로드합니다.

## 사용 패키지

| 패키지 | 사용된 함수/타입 | 용도 |
|--------|-----------------|------|
| `os` | `os.Getenv(key)` | 환경변수에서 TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID 읽기 |
| | `os.UserHomeDir()` | `~/.tfn.env` 경로 구성을 위한 홈 디렉토리 조회 |
| | `os.Open(path)` | `~/.tfn.env` 파일 열기 (읽기 전용) |
| `bufio` | `bufio.NewScanner(r)` | `io.Reader`를 줄 단위 Scanner로 래핑 |
| | `scanner.Scan()` | 다음 줄 읽기 (EOF면 false 반환) |
| | `scanner.Text()` | 현재 줄의 문자열 반환 (개행 제외) |
| `path/filepath` | `filepath.Join(parts...)` | OS별 경로 구분자로 경로 조합 |
| `strings` | `strings.TrimSpace()` | 줄 앞뒤 공백 제거 |
| | `strings.Cut(s, sep)` | "KEY=VALUE"를 KEY와 VALUE로 분리 (Go 1.18+) |
| | `strings.Trim(s, cutset)` | config 값에서 따옴표(`"'`) 제거 |

## 설정 로드 순서

```
  ┌────────────────────────────────────┐
  │ 1. Environment Variables (최우선)   │
  │    TELEGRAM_BOT_TOKEN              │
  │    TELEGRAM_CHAT_ID                │
  └──────────────┬─────────────────────┘
                 │ missing?
                 ▼
  ┌────────────────────────────────────┐
  │ 2. ~/.tfn.env (tfn 전용 설정 파일)  │
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

## ~/.tfn.env 파싱 흐름

```
파일 내용:                      Scanner 동작:
┌──────────────────────────┐
│ # Telegram credentials   │ → Scan() → Text() → '#'으로 시작 → 스킵
│ TELEGRAM_BOT_TOKEN=abc   │ → Scan() → Text() → Cut("=") → key/value 저장
│ TELEGRAM_CHAT_ID=-100    │ → Scan() → Text() → Cut("=") → key/value 저장
│                          │ → Scan() → 빈 줄 → 스킵
└──────────────────────────┘ → Scan() → false (EOF) → 루프 종료
```

## filepath.Join — OS 호환 경로 조합

```go
// Bad - OS 의존적
path := home + "/.tfn.env"          // Linux/macOS에서만 동작
path := home + "\\.tfn.env"         // Windows에서만 동작

// Good - OS 자동 감지
path := filepath.Join(home, ".tfn.env")
// macOS/Linux → /Users/khko/.tfn.env
// Windows    → C:\Users\khko\.tfn.env
```

## strings.Cut — config 파싱의 핵심

```
입력: "TELEGRAM_BOT_TOKEN=123:ABC"
strings.Cut(line, "=")
→ key="TELEGRAM_BOT_TOKEN", value="123:ABC", ok=true

입력: "# 주석 라인"
strings.Cut(line, "=")
→ ok=false → 스킵
```

## bufio.Scanner — 줄 단위 스트리밍

```go
// Bad - 파일 전체를 메모리에 로드
data, _ := os.ReadFile(path)
lines := strings.Split(string(data), "\n")

// Good - 줄 단위로 스트리밍 (메모리 효율적)
scanner := bufio.NewScanner(f)
for scanner.Scan() {
    line := scanner.Text()
}
// → 파일 크기와 무관하게 한 줄 분량의 메모리만 사용
```
