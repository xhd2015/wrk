# Scenario

**Feature**: --force-cd create with channel closed launches interactive shell at new worktree (Branch B)

```
# shell cwd = main repo; WRK_FOLLOWUP_FILE unset; fake bash on PATH
myrepo (cwd)
wrk --force-cd
  -> stdout worktree path (create contract)
  -> stderr mentions wrk --bash-integration --install
  -> fake shell cwd = new worktree
```

## Steps

1. Init main repo; install fake bash (exit 0).
2. Run `wrk --force-cd` with process cwd = main repo and **no** follow-up env.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.UseFollowupEnv = false
	req.CLIArgs = []string{"--force-cd"}
	installFakeBash(t, req, 0)
	return nil
}
```
