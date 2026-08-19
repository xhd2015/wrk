# Scenario

**Feature**: `--merge-back --push` lease-updates existing `origin/<WtBranch>` after FF land

```
# origin already has the worktree branch at local tip
  -> wrk --merge-back -y --push
  -> merged; pushed main and <WtBranch>
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAheadWithOriginBranch(t, req)
	req.Args = []string{"--merge-back", "-y", "--push"}
	return nil
}
```
