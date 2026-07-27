# Scenario

**Feature**: path-like `./real-target` second positional never treated as task

```
wrk <myrepo> ./real-target
  -> create worktree at {WorkRoot}/real-target (relative to shell cwd)
  -> no treat-as-task prompt / error
```

## Steps

1. Init myrepo; shell cwd = WorkRoot.
2. SpawnDir = `./real-target` (path-like by `./` prefix).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := initMyrepoForForgotTask(t, req)
	req.RepoDir = req.WorkRoot
	req.TargetDir = mainRepo
	req.SpawnDir = "./real-target"
	return nil
}
```
