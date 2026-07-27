# Scenario

**Feature**: wrk --bring soft-skips analyse/replace when dep is git but not a Go module

```
# dep git without go.mod -> wrk --bring
#   -> exit 0; external worktree + /external gitignore
#   -> no replace; SKIP is not a go module on stderr
consumer (requires example.com/dep) + mydep (git, no go.mod)
  -> wrk --bring <dep>
  -> stdout external path; stderr SKIP … is not a go module
```

## Steps

1. Create consumer with dep require.
2. Create git repo without `go.mod`.
3. Run `wrk --bring <dep>` from consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, true)
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", false)

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.Args = []string{"--bring", dep}
	return nil
}
```
