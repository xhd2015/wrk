# Scenario

**Feature**: successful create from non-home shell cwd leaves follow-up empty

```
# shell process cwd = main repo (not os.UserHomeDir)
myrepo (cwd) + WRK_FOLLOWUP_FILE=tmp
wrk -> stdout worktree path; follow-up empty (home gate closed)
```

## Steps

1. Init main repo `myrepo`.
2. Run bare `wrk` with process cwd = main repo and follow-up env set.
3. Expect create success but no follow-up `cd` (shell is not yanked).

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	// Shell cwd is main checkout — not FakeHome / user home.
	req.RepoDir = mainRepo
	req.UseFollowupEnv = true
	req.CLIArgs = nil // bare create from main repo cwd
	return nil
}
```
