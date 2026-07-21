# Scenario

**Feature**: ahead external dep on non-TTY `wrk --done` auto-yes cascade merges (no hard guard)

```
consumer wt + ahead external dep (replace dropped) -> wrk --done (non-TTY)
  -> exit 0; dep merged + external removed; consumer removed
```

## Steps

1. Build consumer wt with ahead external dep worktree.
2. Drop consumer replace/require and commit so consumer `--done` can finish.
3. Run `wrk --done` from consumer wt (non-TTY).

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	prepareAheadExternalDepConsumerForDone(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done", "-y"}  // D3: cascade not-included needs -y
	return nil
}
```
