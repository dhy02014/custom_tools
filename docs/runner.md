# runner.go — Terraform Process Execution

Spawns the terraform binary as a child process, manages I/O streaming, and captures the exit code.

## Packages Used

| Package | Function / Type | Purpose |
|---------|----------------|---------|
| `os/exec` | `exec.Command(name, args)` | Create terraform command object (not yet executed) |
| | `cmd.Stdin / Stdout / Stderr` | Configure child process I/O pipes |
| | `cmd.Run()` | Execute process and block until completion |
| | `*exec.ExitError` | Error type for non-zero exits (contains exit code) |
| | `ExitError.ExitCode()` | Extract child process exit code |
| `io` | `io.MultiWriter(writers...)` | Create a writer that duplicates writes to multiple targets |
| `bytes` | `bytes.Buffer` | In-memory buffer to accumulate stderr output |
| | `Buffer.String()` | Convert captured bytes to string |
| `os` | `os.Stdin / Stdout / Stderr` | Terminal standard I/O file descriptors |
| `strings` | `strings.Split()` | Split stderr into lines |
| | `strings.TrimSpace()` | Trim leading/trailing whitespace |
| | `strings.Join()` | Rejoin lines into a single string |

## Transparent Proxy Pattern

```
  ┌─────────────────────────────────────────────────────────────┐
  │                    Transparent Proxy                         │
  │                                                             │
  │   Terminal                                    terraform     │
  │   ┌───────┐                                   ┌───────┐    │
  │   │       │──── stdin (os.Stdin) ────────────>│       │    │
  │   │       │                                   │       │    │
  │   │       │<─── stdout (os.Stdout) ──────────│       │    │
  │   │       │                                   │       │    │
  │   │       │<─── stderr ──┬───────────────────│       │    │
  │   └───────┘              │                    └───────┘    │
  │                          │                                  │
  │                 ┌────────▼────────┐                         │
  │                 │  io.MultiWriter │                         │
  │                 │                 │                         │
  │                 │  os.Stderr ─────┼──> real-time terminal   │
  │                 │  stderrBuf ─────┼──> capture (last 10)   │
  │                 └─────────────────┘                         │
  │                                                             │
  └─────────────────────────────────────────────────────────────┘
```

| I/O | Implementation | Purpose |
|-----|---------------|---------|
| stdin | `cmd.Stdin = os.Stdin` | Support interactive prompts (e.g., `apply` yes/no) |
| stdout | `cmd.Stdout = os.Stdout` | Real-time terminal output (unmodified) |
| stderr | `io.MultiWriter(os.Stderr, &buf)` | Real-time output + error capture for notifications |
| exit code | `exec.ExitError.ExitCode()` | Forward terraform's exit code as-is |

## io.MultiWriter — Equivalent to Unix `tee`

```bash
# Unix equivalent
terraform plan 2>&1 | tee /tmp/stderr.log
#                      ~~~
#               this is io.MultiWriter
```

`bytes.Buffer` implements the `io.Writer` interface, making it usable as a `MultiWriter` target:

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
// bytes.Buffer implements Write()
// → every stderr write from terraform is automatically accumulated
```

## Exit Code Handling

```
cmd.Run() result:
  │
  ├─ err == nil              → exit code 0 (success)
  ├─ err.(*exec.ExitError)   → exit code N (terraform error)
  └─ other error             → exit code 1 (binary not found, etc.)
```

## lastNLines — Error Summary Extraction

Extracts the last N lines from stderr. Keeps notification messages within Telegram's 4096-character limit while preserving the most relevant error context.

```
terraform stderr (tens to hundreds of lines)
     │
     ▼
lastNLines(stderr, 10)
     │
     ▼
last 10 lines → used in notification error summary block
```
