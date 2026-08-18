# Scenario

**Feature**: dry-run prints would: dep-replace and does not mutate go.mod

```
consumer requires dep; no replace yet
  -> wrk --dep-replace <dep> --dry-run
  -> ==== dep-replace (dry-run) ====
  -> would: replace example.com/dep => <abs>
  -> go.mod identical to baseline
  -> exit 0
```

## Steps

1. Seed consumer with require + dep module.
2. Run with `--dry-run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupConsumerWithDep(t, req, true)
	req.Args = []string{"--dep-replace", req.DepDir, "--dry-run"}
	return nil
}
```
