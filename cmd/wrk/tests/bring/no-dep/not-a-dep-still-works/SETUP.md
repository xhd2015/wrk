# Scenario

**Feature**: --bring --no-dep succeeds even when dep is not required (no analyse)

```
# consumer has go.mod but no require of dep -> wrk --bring <dep> --no-dep
#   -> exit 0; external worktree + gitignore
#   -> go.mod unchanged; no SKIP (module analyse skipped)
consumer (go.mod, no require) + mydep
  -> wrk --bring <dep> --no-dep
  -> stdout external path; no SKIP on stderr
```

## Steps

1. Create consumer **without** requiring dep.
2. Create valid dep repo.
3. Snapshot go.mod; run `wrk --bring <dep> --no-dep`.

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
	snapshotBringGoMod(t, req, consumer)
	req.Args = []string{"--bring", dep, "--no-dep"}
	return nil
}
```
