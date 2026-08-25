# Scenario

**Feature**: `--dep-replace --undo --dry-run` plans drop; no go.mod write

```
WT introduced replace
  -> wrk --dep-replace --undo --dry-run
  -> would: drop + would: go mod tidy; go.mod unchanged
```

## Steps

1. Seed introduced replace.
2. Run undo dry-run.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupUndoIntroduced(t, req)
	req.Args = []string{"--dep-replace", "--undo", "--dry-run"}
	return nil
}
```
