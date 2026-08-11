# Scenario

**Feature**: --all uses git toplevel of cwd (linked worktree), not main repo

```
# app main + linked worktree both start with lib@v1.0.0
cwd=linked-wt -> wrk --dep-update --all
  -> pin + tidy under worktree modules only
  -> linked go.mod require@v1.2.3
  -> main go.mod still v1.0.0
```

## Steps

1. Seed owner + app main; `git worktree add` linked checkout.
2. Run apply from linked worktree with modproxy.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAllWorktreeOutdated(t, req)
	req.Args = []string{"--dep-update", "--all"}
	return nil
}
```
