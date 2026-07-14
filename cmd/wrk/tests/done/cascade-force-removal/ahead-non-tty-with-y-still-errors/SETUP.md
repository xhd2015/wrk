# Scenario

**Bug**: `-y` does not bypass cascade non-TTY guard for ahead external deps

```
consumer wt + ahead external dep -> wrk --done -y (non-TTY) -> non-zero; dep wt preserved
```

## Steps

1. Build consumer wt with ahead external dep worktree.
2. Run `wrk --done -y` from consumer wt (non-TTY).

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done", "-y"}
	return nil
}
```
