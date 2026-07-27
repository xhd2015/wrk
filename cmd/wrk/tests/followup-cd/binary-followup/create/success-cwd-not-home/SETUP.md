# Scenario

**Feature**: successful create from non-home shell cwd leaves follow-up empty

```
# shell process cwd = main repo (not os.UserHomeDir)
myrepo (cwd) + WRK_FOLLOWUP_FILE=tmp
wrk -> stdout worktree path; follow-up empty (home gate closed)
```

## Steps

1. Init main repo `myrepo`.
2. Run `wrk --new` with process cwd = main repo and follow-up env set.
3. Expect create success but no follow-up `cd` (shell is not yanked).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := setupMainRepo(t, req)
	// Shell cwd is main checkout — not FakeHome / user home.
	req.RepoDir = mainRepo
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--new"} // create entry (bare no-args is dashboard)
	return nil
}
```
