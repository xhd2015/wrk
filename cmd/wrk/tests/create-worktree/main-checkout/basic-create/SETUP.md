# Scenario

**Feature**: first `wrk --new` from fresh repo on main

```
# fresh repo on main, no prior worktrees under WRK_HOME
myrepo (main) -> wrk --new -> ~/.wrk/worktrees/myrepo-main-2026-06-30
```

## Steps

1. Initialize git repo `myrepo` on branch `main` with one commit.
2. Run `wrk --new` with cwd set to `myrepo` (Args set by create-worktree parent).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	initGitRepoOnMain(t, req.RepoDir)
	return nil
}
```
