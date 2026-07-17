# Scenario

**Feature**: bare non-TTY `--done` cascades ahead dep with default auto-yes (then replace guard)

```
consumer wt + ahead external dep -> wrk --done (non-TTY)
  -> cascade merges dep; external gone; parent blocked by replace
```

## Steps

1. Build consumer wt with ahead external dep worktree.
2. Run bare `wrk --done` (non-TTY).

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done"}
	return nil
}
```
