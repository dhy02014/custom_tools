# tfn - Terraform Notify

Terraform CLI wrapper that sends **Telegram notifications** on command completion.

```
tfn plan -chdir=terraform/environments/crm_dev

= terraform plan -chdir=terraform/environments/crm_dev
+ Telegram notification on completion
```

## Features

- **Transparent Proxy** — `tfn <args>` = `terraform <args>` (stdin/stdout/stderr/exit code 100% 보존)
- **Selective Notification** — `plan`, `apply`, `init`, `validate`만 알림, 나머지는 패스스루
- **Zero Dependencies** — Go stdlib만 사용, 외부 라이브러리 없음
- **Graceful Degradation** — 설정 없으면 알림 스킵, terraform만 실행

## Installation

```bash
# Build
cd ~/khko_tools/tfn
go build -o tfn .

# Install
sudo ln -sf ~/khko_tools/tfn/tfn /usr/local/bin/tfn
```

## Configuration

**Option A: Config File (`~/.tfn.env`) — 권장**

```env
# tfn - Terraform Notify configuration
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
TELEGRAM_CHAT_ID=-100123456789
```

**Option B: Environment Variables**

```bash
export TELEGRAM_BOT_TOKEN="123456:ABC-DEF..."
export TELEGRAM_CHAT_ID="-100123456789"
```

## Usage

```bash
# 알림 대상 명령어
tfn plan -chdir=terraform/environments/crm_dev
tfn apply -chdir=terraform/environments/crm_prd
tfn init -upgrade
tfn validate

# 알림 없이 패스스루
tfn fmt
tfn state list
tfn version
```

## How It Works

```
  Terminal ──> tfn ──> terraform
                │
                ├─ stdin  직접 연결 (interactive 지원)
                ├─ stdout 실시간 스트리밍
                ├─ stderr 실시간 스트리밍 + 버퍼 캡처
                └─ exit code 그대로 반환

  terraform 종료 후:
    알림 대상 명령? ──Yes──> Telegram 전송
                    ──No──> 종료
```

## Telegram Message

**Success:**
```
✅ Terraform plan 성공
━━━━━━━━━━━━━━━━━━━━
📁 디렉토리: crm_dev
⏱ 소요시간: 12.3s
💻 명령어: terraform plan -chdir=...
```

**Failure** (stderr 마지막 10줄 포함):
```
❌ Terraform validate 실패
━━━━━━━━━━━━━━━━━━━━
📁 디렉토리: crm_dev
⏱ 소요시간: 0.8s
💻 명령어: terraform validate

📋 에러 요약:
  Error: Missing required argument...
```

## Project Structure

```
~/khko_tools/tfn/
├── main.go        # Entrypoint, 알림 대상 판별
├── runner.go      # terraform 실행, I/O 스트리밍
├── notifier.go    # Telegram Bot API 호출
├── config.go      # 설정 로드 (env vars → ~/.tfn.env)
├── go.mod         # Go module (stdlib only)
├── docs/          # 모듈별 상세 문서
│   ├── main.md
│   ├── runner.md
│   ├── notifier.md
│   └── config.md
└── README.md
```

## Documentation

각 Go 파일의 패키지 사용 상세, 동작 원리, 다이어그램은 `docs/`를 참고하세요:

| 문서 | 내용 |
|------|------|
| [docs/main.md](docs/main.md) | 엔트리포인트, 알림 판별 로직, 실행 흐름 |
| [docs/runner.md](docs/runner.md) | Transparent Proxy 패턴, I/O 스트리밍, exit code 처리 |
| [docs/notifier.md](docs/notifier.md) | Telegram API 호출, JSON 직렬화, MarkdownV2 이스케이프 |
| [docs/config.md](docs/config.md) | 설정 로드 순서, .env 파싱, 경로 조합 |

## Technical Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Language | Go | Single binary, no runtime deps, cross-compile |
| Dependencies | stdlib only | No supply chain risk, minimal build complexity |
| Config format | `.env` (KEY=VALUE) | Simple parsing without YAML library |
| Notification | Telegram Bot API | Simple HTTP POST, no SDK needed |
| stderr capture | Last 10 lines | Enough context within Telegram message limit |
