# Scenario

**Feature**: wrk --dep errors when dep path is git but not a Go module

```
# git repo without go.mod -> wrk --dep -> non-zero
```

## Steps

1. Create consumer with dep require.
2. Create git repo without `go.mod`.
3. Run `wrk --dep`.

```go
func Setup(t *testing.T, req *Request) error {
	consumer := initConsumerRepo(t, req.WorkRoot, true)
	dep := initDepRepo(t, req.WorkRoot, "mydep", false)

	req.RepoDir = consumer
	req.DepPath = dep
	req.Args = []string{"--dep", dep}
	return nil
}
```