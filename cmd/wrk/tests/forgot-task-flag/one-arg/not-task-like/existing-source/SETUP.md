# Scenario

**Feature**: one-arg existing repo path creates normally (no task-like prompt)

```
wrk <abs-myrepo> from WorkRoot
  -> default WRK_HOME worktree (no task slug); no treat-as-task
```

## Steps

1. Init myrepo.
2. First positional = absolute path to myrepo (existing source; path-like / resolvable).
3. Shell cwd = WorkRoot.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initMyrepoForForgotTask(t, req)
	req.RepoDir = req.WorkRoot
	req.TargetDir = mainRepo
	return nil
}
```
