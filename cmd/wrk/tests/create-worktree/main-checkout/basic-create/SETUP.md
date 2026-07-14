# Scenario

**Feature**: first wrk from fresh repo on main

```
# fresh repo on main, no prior worktrees under WRK_HOME
myrepo (main) -> wrk -> ~/.wrk/worktrees/myrepo-main-2026-06-30
```

## Steps

1. Initialize git repo `myrepo` on branch `main` with one commit.
2. Run `wrk` with cwd set to `myrepo`.

```go
func Setup(t *testing.T, req *Request) error {
	initGitRepoOnMain(t, req.RepoDir)
	return nil
}
```