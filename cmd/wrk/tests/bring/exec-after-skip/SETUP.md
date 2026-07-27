# Scenario

**Feature**: wrk --bring --exec runs in external worktree even after soft SKIP

```
# not-a-dependency + --exec pwd
#   -> exit 0; SKIP on stderr; worktree + gitignore
#   -> stdout: <external-abs>\n<external-abs>\n  (mode path then child pwd)
consumer (go.mod, no require) + mydep (module example.com/dep)
  -> wrk --bring <dep> --exec pwd
  -> child cmd.Dir = external abs
```

## Steps

1. Create consumer **without** requiring dep (soft SKIP path).
2. Create valid dep repo.
3. Run `wrk --bring <dep> --exec pwd` from consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, false)
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", dep, "--exec", "pwd"}
	return nil
}
```
