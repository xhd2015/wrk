# Scenario

**Feature**: `wrk --done -y` cascades ahead external dep on non-TTY (then replace guard)

```
consumer wt + ahead external dep -> wrk --done -y
  -> cascade merges dep; external gone; parent blocked by replace
```

## Steps

1. Build consumer wt with ahead external dep.
2. Run `wrk --done -y` on non-TTY.

```go
func Setup(t *testing.T, req *Request) error {
	setupConsumerWithAheadExternalDep(t, req)
	req.RepoDir = req.WtDir
	req.Args = []string{"--done", "-y"}
	return nil
}
```
