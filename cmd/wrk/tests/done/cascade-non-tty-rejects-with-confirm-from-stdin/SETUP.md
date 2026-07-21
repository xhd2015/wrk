# Scenario

**Feature**: non-TTY cascade ahead auto-yes succeeds (no pre-flight hard guard)

```
consumer wt + ahead external dep (replace dropped) -> wrk --done (non-TTY)
  -> exit 0; cascade merges dep; both wts removed
```

## Steps

1. Build consumer wt with ahead external dep worktree.
2. Drop consumer replace/require and commit.
3. Run `wrk --done` (default auto-yes; no `--confirm`).

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	prepareAheadExternalDepConsumerForDone(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done"}
	return nil
}
```
