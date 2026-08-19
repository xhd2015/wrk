# Scenario

**Feature**: `--done --push` lease-updates `origin/<WtBranch>` after rebase+remove

```
# diverged land rewrites SHAs; --done still force-updates the standing PR branch
  -> wrk --done -y --push
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDivergedWithOriginBranch(t, req)
	req.Args = []string{"--done", "-y", "--push"}
	return nil
}
```
