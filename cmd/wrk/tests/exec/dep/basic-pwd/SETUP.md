# Scenario

**Feature**: `--dep` + `--exec pwd` runs pwd in the external dependency worktree

```
consumer + mydep -> wrk --dep mydep --exec pwd
  -> external = {consumer}/external/mydep-main-2026-06-30
  -> stdout: external\nexternal\n
```

## Steps

1. Create consumer with require `example.com/dep`.
2. Create dep repo `mydep` with that module.
3. Run `wrk --dep <dep> --exec pwd` from consumer.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initExecConsumer(t, req.WorkRoot)
	dep := initExecDepRepo(t, req.WorkRoot, "mydep")

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.Args = []string{"--dep", dep, "--exec", "pwd"}
	return nil
}
```
