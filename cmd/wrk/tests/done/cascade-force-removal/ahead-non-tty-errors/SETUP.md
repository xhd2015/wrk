# Scenario

**Bug**: ahead external dep on non-TTY `wrk --done` errors instead of force-removing

```
consumer wt + ahead external dep -> wrk --done (non-TTY) -> non-zero; dep wt + commits remain
```

## Steps

1. Build consumer wt with ahead external dep worktree.
2. Run `wrk --done` from consumer wt (non-TTY).

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done"}
	return nil
}
```
