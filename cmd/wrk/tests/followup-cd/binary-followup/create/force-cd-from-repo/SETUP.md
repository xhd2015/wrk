# Scenario

**Feature**: --force-cd on successful create from non-home shell cwd writes follow-up (home gate bypass)

```
# shell process cwd = main repo (not os.UserHomeDir)
myrepo (cwd) + WRK_FOLLOWUP_FILE=tmp
wrk --force-cd -> stdout worktree path; follow-up: cd <abs-worktree>
# no interactive shell (channel open = Branch A)
```

## Steps

1. Init main repo `myrepo`.
2. Run `wrk --force-cd` with process cwd = main repo and follow-up env set.
3. Expect create success and ungated follow-up `cd` to the new worktree.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	// Shell cwd is main checkout — home gate would normally skip write.
	req.RepoDir = mainRepo
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--force-cd"}
	return nil
}
```
