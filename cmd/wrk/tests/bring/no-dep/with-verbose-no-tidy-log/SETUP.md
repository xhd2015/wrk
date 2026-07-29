# Scenario

**Feature**: --bring --no-dep -v may log git worktree but never go mod tidy

```
# matching consumer + dep -> wrk --bring <dep> --no-dep -v
#   -> stderr may have git worktree add pre-line / stream
#   -> stderr must NOT contain go … mod tidy pre-line
consumer (require dep) + mydep
  -> wrk --bring <dep> --no-dep -v -> no tidy logs
```

## Steps

1. Same matching fixtures as `skips-replace-tidy`.
2. Run with `--no-dep` and `-v`.

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
	req.Args = []string{"--bring", dep, "--no-dep", "-v"}
	return nil
}
```
