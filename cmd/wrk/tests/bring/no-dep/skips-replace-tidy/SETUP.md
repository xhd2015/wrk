# Scenario

**Feature**: --bring --no-dep creates external worktree without replace or tidy

```
# consumer requires dep -> wrk --bring <dep> --no-dep
#   -> external/mydep + /external gitignore
#   -> go.mod unchanged (no replace); no tidy side effects
consumer (require example.com/dep) + mydep
  -> wrk --bring <dep> --no-dep -> stdout abs path; go.mod byte-identical
```

## Steps

1. Create consumer requiring `example.com/dep`.
2. Create dep repo `mydep` with that module.
3. Snapshot consumer `go.mod`.
4. Run `wrk --bring <dep> --no-dep` from consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, true)
	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePath
	snapshotBringGoMod(t, req, consumer)
	req.Args = []string{"--bring", dep, "--no-dep"}
	return nil
}
```
