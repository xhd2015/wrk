# Scenario

**Feature**: `wrk --done -y` succeeds when cascaded external dep is ahead (non-TTY auto-yes)

```
consumer wt + ahead external dep (replace dropped) -> wrk --done -y
  -> exit 0; external + consumer removed; dep fix merged
```

## Steps

1. Build consumer wt with ahead external dep (shared `setupConsumerWithAheadExternalDep`).
2. Drop replace/require and commit.
3. Run `wrk --done -y` on non-TTY.

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	prepareAheadExternalDepConsumerForDone(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done", "-y"}
	return nil
}
```
