# main.go — Entrypoint

CLI 인자를 파싱하고, 알림 대상 명령어를 판별하여 runner → notifier 순으로 오케스트레이션합니다.

## 사용 패키지

| 패키지 | 사용된 함수/변수 | 용도 |
|--------|-----------------|------|
| `os` | `os.Args` | CLI 인자 수신 (`os.Args[1:]`로 tfn 자신 제외) |
| | `os.Exit(code)` | terraform exit code를 프로세스에 그대로 반환 |
| | `os.Stderr` | 알림 실패 시 경고 메시지 출력 대상 |
| | `os.Getwd()` | `-chdir` 미지정 시 현재 디렉토리명 추출 |
| `fmt` | `fmt.Fprintf()` | stderr에 포맷팅된 에러 메시지 출력 |
| `time` | `time.Now()` | terraform 실행 직전 시각 기록 |
| | `time.Since(t)` | 실행 완료 후 경과 시간 계산 |

## 알림 대상 명령어 판별

```go
var notifyCommands = map[string]bool{
    "plan":     true,
    "apply":    true,
    "init":     true,
    "validate": true,
}
```

첫 번째 non-flag 인자를 subcommand로 인식합니다:

```
tfn -chdir=path plan    → subcommand = "plan"  (알림 O)
tfn fmt                 → subcommand = "fmt"   (알림 X)
tfn version             → subcommand = "version" (알림 X)
```

## 실행 흐름

```
main()
  │
  ├─ args 파싱 (os.Args[1:])
  │
  ├─ subcommand 감지 (첫 번째 non-flag arg)
  │
  ├─ time.Now() ─── 타이머 시작
  │
  ├─ runTerraform(args) ─── runner.go 호출
  │
  ├─ time.Since(start) ─── 소요시간 계산
  │
  ├─ notifyCommands[subcommand]?
  │   ├─ Yes → loadConfig() → sendNotification()
  │   └─ No  → 스킵
  │
  └─ os.Exit(result.ExitCode)
```

## extractWorkDir 함수

`-chdir=terraform/environments/crm_dev`에서 마지막 경로 컴포넌트(`crm_dev`)를 추출합니다.
Telegram 메시지의 "디렉토리" 필드에 사용됩니다.

```
-chdir=terraform/environments/crm_dev → "crm_dev"
-chdir=dev                            → "dev"
(미지정)                               → 현재 디렉토리명 (os.Getwd)
```
