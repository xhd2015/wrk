# Scenario

**Feature**: `wrk -t <task>` still creates without `--new`

```
myrepo -> wrk -t "ship feature"
  -> exit 0; worktree path includes task slug
  -> no --new required
```

## Steps

1. Init `myrepo` on main.
2. Run `wrk -t "ship feature"` (TaskDesc + TaskFlag; no `--new`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupDashboardMainRepo(t, req)
	req.TaskDesc = "ship feature"
	req.TaskFlag = "-t"
	req.Args = nil
	return nil
}
```
