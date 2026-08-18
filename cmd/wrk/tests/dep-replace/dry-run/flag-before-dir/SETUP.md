# Scenario

**Feature**: `--dep-replace` is a Bool; `--dry-run` after it is a flag, not a path

```
consumer requires dep
  -> wrk --dep-replace --dry-run <dep>
  -> ==== dep-replace (dry-run) ====
  -> would: replace; go.mod unchanged
```

## Steps

1. Seed consumer with require + dep module.
2. Run `--dep-replace --dry-run <dep>` (dry-run between flag and dir).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupConsumerWithDep(t, req, true)
	req.Args = []string{"--dep-replace", "--dry-run", req.DepDir}
	return nil
}
```
