# Scenario

**Feature**: `-y` on non-TTY cascade ahead is a no-op synonym of default auto-yes (success)

```
consumer wt + ahead external dep (replace dropped) -> wrk --done -y (non-TTY)
  -> exit 0; dep merged + both wts removed
```

## Steps

1. Build consumer wt with ahead external dep worktree.
2. Drop consumer replace/require and commit.
3. Run `wrk --done -y` from consumer wt (non-TTY).

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	prepareAheadExternalDepConsumerForDone(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done", "-y"}
	return nil
}
```
