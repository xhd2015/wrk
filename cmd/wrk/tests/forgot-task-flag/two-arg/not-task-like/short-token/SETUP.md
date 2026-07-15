# Scenario

**Feature**: short token `out` without spaces stays target-dir (no promote)

```
wrk <myrepo> out
  -> worktree at {WorkRoot}/out (missing target, parent exists)
```

## Steps

1. Init myrepo; SpawnDir = `out`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initMyrepoForForgotTask(t, req)
	req.RepoDir = req.WorkRoot
	req.TargetDir = mainRepo
	req.SpawnDir = "out"
	return nil
}
```
