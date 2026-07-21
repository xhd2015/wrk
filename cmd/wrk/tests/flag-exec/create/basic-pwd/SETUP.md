# Scenario

**Feature**: create + `--exec pwd` runs pwd in the new worktree

```
myrepo (main) -> wrk --exec pwd
  -> wt = {WRK_HOME}/worktrees/myrepo-main-2026-06-30
  -> stdout: wt\nwt\n
```

## Steps

1. Initialize git repo `myrepo` on `main`.
2. Run `wrk --exec pwd` from the main checkout.

```go
func Setup(t *testing.T, req *Request) error {
	initGitRepoOnMain(t, req.RepoDir)
	req.Args = []string{"--exec", "pwd"}
	return nil
}
```
