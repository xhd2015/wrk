# Scenario

**Feature**: broken nested checkout is warned + skipped; dep-update still plans

```
# primary requires dep; sandbox/broken-wt/.git -> nonexistent gitdir
cwd=primary -> wrk --dep-update <dep> --dry-run
  -> stderr: warning: skipping nested checkout sandbox/broken-wt: …
  -> stdout: ==== dep-update (dry-run) ====; would: pin; exit 0
  -> go.mod unchanged
```

## Steps

1. Seed soft-skip broken nested fixture.
2. Run dry-run.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSoftSkipBrokenNested(t, req)
	req.Args = []string{"--dep-update", req.DepDir, "--dry-run"}
	return nil
}
```
