# Scenario

**Feature**: dry-run prints would: dep-update and would: go mod tidy; no write

```
consumer has replace + require v0.0.1; dep tags up to v0.0.2
  -> wrk --dep-update <dep> --dry-run
  -> ==== dep-update (dry-run) ====
  -> would: pin example.com/dep v0.0.1 -> v0.0.2
  -> would: go mod tidy
  -> go.mod identical to baseline
  -> exit 0
```

## Steps

1. Seed drop-replace-latest fixture.
2. Run with `--dry-run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDropReplaceLatest(t, req)
	req.Args = []string{"--dep-update", req.DepDir, "--dry-run"}
	return nil
}
```
