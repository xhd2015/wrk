# Scenario

**Feature**: `wrk --done -y` rejected when cascaded external dep is ahead (non-TTY)

```
consumer wt + ahead external dep -> wrk --done -y -> non-zero; external wt preserved
```

## Steps

1. Build consumer wt with ahead external dep (shared `setupConsumerWithAheadExternalDep`).
2. Run `wrk --done -y` on non-TTY.

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done", "-y"}
	return nil
}
```
