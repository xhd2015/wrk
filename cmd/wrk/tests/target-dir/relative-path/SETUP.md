# Scenario

**Feature**: relative <target-dir> is resolved against the shell (process) cwd, not <dir>

```
# process cwd = {WorkRoot}; source repo = {WorkRoot}/myrepo; <target-dir> = "wt" (relative)
myrepo (main) -> wrk myrepo wt -> {WorkRoot}/wt (resolved vs shell cwd, NOT vs {WorkRoot}/myrepo)
```

## Preconditions

- Proves the resolution basis: a relative `<target-dir>` joins the process cwd, not the
  `<dir>` source repo. If it were resolved against `<dir>`, the worktree would land at
  `{WorkRoot}/myrepo/wt` instead.

## Steps

1. Source repo `myrepo` on `main` is initialized by the parent setup.
2. Set `req.SpawnDir = "wt"` (relative — intentionally not absolute).
3. Run `wrk myrepo wt` from process cwd `{WorkRoot}`.

```go
func Setup(t *testing.T, req *Request) error {
	req.SpawnDir = "wt"
	return nil
}
```
