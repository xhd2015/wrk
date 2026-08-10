# Scenario

**Feature**: dry-run prints would: dep-update and does not mutate go.mod

```
consumer has replace + require v0.0.1; dep tags up to v0.0.2
  -> wrk --dep-update <dep> --dry-run
  -> would: dep-update example.com/dep -> v0.0.2
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
