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
?   	firmwarehub/cmd/firmwarehub	[no test files]
ok  	firmwarehub/internal/archive	0.040s
ok  	firmwarehub/internal/catalog	0.048s
ok  	firmwarehub/internal/domain	0.007s
--- FAIL: Test2203BusinessRegression (0.00s)
    workflows_test.go:80: imported score for second-row: got 15 want 92
FAIL
FAIL	firmwarehub/internal/flow020	0.060s
ok  	firmwarehub/internal/importer	0.026s
ok  	firmwarehub/internal/integration	0.027s
ok  	firmwarehub/internal/review	0.033s
ok  	firmwarehub/internal/store	0.039s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/firmwarehub): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/firmwarehub): exit `0`
