# main.go — Entrypoint

Parses CLI arguments, detects the terraform subcommand, and orchestrates the runner → notifier pipeline.

## Packages Used

| Package | Function / Variable | Purpose |
|---------|-------------------|---------|
| `os` | `os.Args` | Receive CLI arguments (`os.Args[1:]` excludes the tfn binary itself) |
| | `os.Exit(code)` | Forward terraform's exit code to the calling process |
| | `os.Stderr` | Output warning messages when notification fails |
| | `os.Getwd()` | Extract current directory name when `-chdir` is not specified |
| `fmt` | `fmt.Fprintf()` | Write formatted error messages to stderr |
| `time` | `time.Now()` | Record timestamp just before terraform execution |
| | `time.Since(t)` | Calculate elapsed time after execution completes |

## Subcommand Detection

```go
var notifyCommands = map[string]bool{
    "plan":     true,
    "apply":    true,
    "init":     true,
    "validate": true,
}
```

The first non-flag argument is identified as the subcommand:

```
tfn -chdir=path plan    → subcommand = "plan"     (notify)
tfn fmt                 → subcommand = "fmt"      (passthrough)
tfn version             → subcommand = "version"  (passthrough)
```

## Execution Flow

```
main()
  │
  ├─ Parse args (os.Args[1:])
  │
  ├─ Detect subcommand (first non-flag arg)
  │
  ├─ time.Now() ─── start timer
  │
  ├─ runTerraform(args) ─── calls runner.go
  │
  ├─ time.Since(start) ─── calculate duration
  │
  ├─ notifyCommands[subcommand]?
  │   ├─ Yes → loadConfig() → sendNotification()
  │   └─ No  → skip
  │
  └─ os.Exit(result.ExitCode)
```

## extractWorkDir

Extracts the last path component from the `-chdir=<path>` argument for use in the Telegram message "Directory" field.

```
-chdir=terraform/environments/crm_dev → "crm_dev"
-chdir=dev                            → "dev"
(not specified)                        → current directory name via os.Getwd()
```
