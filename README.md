# tfn

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Dependencies](https://img.shields.io/badge/dependencies-stdlib%20only-brightgreen)](#technical-decisions)

A transparent Terraform CLI wrapper that sends **Telegram notifications** on command completion.

```bash
# Before: terraform plan (no feedback when away from terminal)
terraform plan -chdir=terraform/environments/crm_dev

# After: same behavior + Telegram notification
tfn plan -chdir=terraform/environments/crm_dev
```

## Features

- **Transparent Proxy** — `tfn <args>` is identical to `terraform <args>` (stdin/stdout/stderr/exit code fully preserved)
- **Selective Notification** — Only notifies on `plan`, `apply`, `init`, `validate`; all other commands pass through silently
- **Zero Dependencies** — Built with Go standard library only; no external packages
- **Graceful Degradation** — Runs terraform normally when credentials are not configured

## Quick Start

### Prerequisites

- Go 1.21+
- Terraform CLI installed and available in `$PATH`
- Telegram Bot token and Chat ID ([how to create a bot](https://core.telegram.org/bots#how-do-i-create-a-bot))

### Build & Install

```bash
git clone https://github.com/dhy02014/custom_tools.git
cd custom_tools
go build -o tfn .
sudo ln -sf "$(pwd)/tfn" /usr/local/bin/tfn
```

### Configuration

Create `~/.tfn.env`:

```env
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
TELEGRAM_CHAT_ID=-100123456789
```

Alternatively, use environment variables:

```bash
export TELEGRAM_BOT_TOKEN="123456:ABC-DEF..."
export TELEGRAM_CHAT_ID="-100123456789"
```

> Config priority: Environment variables > `~/.tfn.env` > no config (notification skipped)

## Usage

```bash
# Commands that trigger notifications
tfn plan -chdir=terraform/environments/dev
tfn apply -chdir=terraform/environments/prd
tfn init -upgrade
tfn validate

# Commands that pass through without notification
tfn fmt
tfn state list
tfn version
tfn output
```

## How It Works

```
Terminal ──> tfn ──> terraform
               │
               ├── stdin   directly connected (interactive prompts supported)
               ├── stdout  real-time streaming
               ├── stderr  real-time streaming + buffer capture
               └── exit code preserved

After terraform exits:
  notify-worthy command? ──Yes──> Send Telegram message
                         ──No──> Exit
```

### Notification Format

**On success:**
```
✅ Terraform plan succeeded
━━━━━━━━━━━━━━━━━━━━
📁 Directory: crm_dev
⏱ Duration: 12.3s
💻 Command: terraform plan -chdir=...
```

**On failure** (includes last 10 lines of stderr):
```
❌ Terraform validate failed
━━━━━━━━━━━━━━━━━━━━
📁 Directory: crm_dev
⏱ Duration: 0.8s
💻 Command: terraform validate

📋 Error summary:
  Error: Missing required argument...
```

## Project Structure

```
.
├── main.go        # Entrypoint, subcommand detection, orchestration
├── runner.go      # Terraform process execution, I/O streaming
├── notifier.go    # Telegram Bot API client, message formatting
├── config.go      # Configuration loader (env vars / .env file)
├── go.mod
├── docs/          # Per-module detailed documentation
│   ├── main.md
│   ├── runner.md
│   ├── notifier.md
│   └── config.md
├── LICENSE
└── README.md
```

## Documentation

Detailed documentation for each module (package usage, diagrams, internals):

| Document | Description |
|----------|-------------|
| [docs/main.md](docs/main.md) | Entrypoint, subcommand routing, execution flow |
| [docs/runner.md](docs/runner.md) | Transparent proxy pattern, I/O streaming, exit code handling |
| [docs/notifier.md](docs/notifier.md) | Telegram API integration, JSON serialization, MarkdownV2 escaping |
| [docs/config.md](docs/config.md) | Config loading priority, `.env` file parsing |

## Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Single binary, no runtime dependencies, easy cross-compilation |
| Dependencies | stdlib only | Zero supply chain risk, minimal build complexity |
| Config format | `.env` (KEY=VALUE) | Simple line-by-line parsing without YAML library |
| Notification | Telegram Bot API | Single HTTP POST call, no SDK needed |
| stderr capture | Last 10 lines | Sufficient context within Telegram's 4096-char message limit |
| MarkdownV2 escape | Manual | Only 19 special characters; library would be overkill |

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -m 'feat: add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
