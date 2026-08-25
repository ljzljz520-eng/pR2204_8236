# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
ok  	examvault/cmd/examvault	0.001s
ok  	examvault/internal/api	0.010s
ok  	examvault/internal/audit	0.010s
ok  	examvault/internal/crypto	0.005s
ok  	examvault/internal/domain	0.001s
ok  	examvault/internal/flow	0.010s
--- FAIL: Test2204BusinessRegression (0.00s)
    regression_test.go:33: paper-b permission = staff
FAIL
FAIL	examvault/internal/flow007	0.004s
ok  	examvault/internal/report	0.001s
ok  	examvault/internal/store	0.004s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/examvault): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/examvault): exit `0`
