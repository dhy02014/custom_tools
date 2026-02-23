# runner.go — Terraform Process Execution

terraform 바이너리를 자식 프로세스로 실행하고, I/O 스트리밍과 exit code 캡처를 담당합니다.

## 사용 패키지

| 패키지 | 사용된 함수/타입 | 용도 |
|--------|-----------------|------|
| `os/exec` | `exec.Command(name, args)` | terraform 명령 객체 생성 (아직 실행 안 함) |
| | `cmd.Stdin / Stdout / Stderr` | 자식 프로세스 I/O 파이프 연결 설정 |
| | `cmd.Run()` | 프로세스 실행 + 종료까지 blocking 대기 |
| | `*exec.ExitError` | 비정상 종료 에러 타입 (exit code 포함) |
| | `ExitError.ExitCode()` | 자식 프로세스의 exit code 추출 |
| `io` | `io.MultiWriter(writers...)` | 여러 Writer에 동시에 쓰는 Writer 생성 |
| `bytes` | `bytes.Buffer` | stderr 출력을 메모리에 누적 캡처 |
| | `Buffer.String()` | 캡처된 바이트를 문자열로 변환 |
| `os` | `os.Stdin / Stdout / Stderr` | 터미널 표준 I/O 파일 디스크립터 |
| `strings` | `strings.Split()` | stderr를 줄 단위로 분리 |
| | `strings.TrimSpace()` | 앞뒤 공백/개행 제거 |
| | `strings.Join()` | 분리된 줄을 다시 합침 |

## Transparent Proxy 패턴

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
  │                 │  os.Stderr ─────┼──> 터미널 실시간 출력   │
  │                 │  stderrBuf ─────┼──> 에러 캡처 (last 10) │
  │                 └─────────────────┘                         │
  │                                                             │
  └─────────────────────────────────────────────────────────────┘
```

| I/O | 처리 방식 | 목적 |
|-----|----------|------|
| stdin | `cmd.Stdin = os.Stdin` | `apply` 시 yes/no 프롬프트 지원 |
| stdout | `cmd.Stdout = os.Stdout` | 실시간 터미널 출력 (변조 없음) |
| stderr | `io.MultiWriter(os.Stderr, &buf)` | 실시간 출력 + 에러 내용 캡처 |
| exit code | `exec.ExitError.ExitCode()` | terraform exit code 그대로 반환 |

## io.MultiWriter — Unix tee와 동일한 개념

```
$ terraform plan 2>&1 | tee /tmp/stderr.log
                         ~~~
                  이것이 io.MultiWriter
```

`bytes.Buffer`가 `io.Writer` 인터페이스를 구현하기 때문에 `MultiWriter`의 대상으로 사용 가능합니다:

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
// bytes.Buffer는 Write() 메서드를 구현
// → terraform이 stderr에 쓸 때마다 Buffer에도 자동 축적
```

## Exit Code 처리

```
cmd.Run() 결과:
  │
  ├─ err == nil              → exit code 0 (성공)
  ├─ err.(*exec.ExitError)   → exit code N (terraform 에러)
  └─ 그 외 error             → exit code 1 (binary not found 등)
```

## lastNLines — 에러 요약 추출

stderr에서 마지막 N줄만 잘라냅니다. Telegram 메시지 크기 제한(4096자) 내에서 핵심 에러만 전달하기 위함입니다.

```
terraform stderr (수십~수백 줄)
     │
     ▼
lastNLines(stderr, 10)
     │
     ▼
마지막 10줄만 반환 → 알림 메시지의 에러 요약 블록에 사용
```
