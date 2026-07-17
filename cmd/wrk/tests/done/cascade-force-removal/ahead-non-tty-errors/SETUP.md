# Scenario

**Feature**: default auto-yes cascades ahead external dep on non-TTY (then replace guard)

```
consumer wt + ahead external dep -> wrk --done (non-TTY)
  -> cascade merges dep; external gone; parent blocked by replace
```

## Steps

1. Build consumer wt with ahead external dep worktree.
2. Run bare `wrk --done` from consumer wt (non-TTY).

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done"}
	return nil
}
```
