# Scenario

**Feature**: successful create from user home with WRK_FOLLOWUP_FILE writes cd to worktree

```
# shell process cwd = FakeHome (os.UserHomeDir in tests); source via dir arg
FakeHome (cwd) + WRK_FOLLOWUP_FILE=tmp
wrk <mainRepo> -> stdout worktree path; follow-up: cd <abs-worktree>
```

## Steps

1. Init main repo `myrepo`.
2. Set process cwd to FakeHome (test user home).
3. Run `wrk <mainRepo>` with follow-up env set (positional create so home gate sees shell cwd, not source workDir).

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	// Shell cwd must be exact user home for the create home gate.
	// Binary Run uses cmd.Dir = req.RepoDir; HOME=FakeHome so UserHomeDir → FakeHome.
	req.RepoDir = req.FakeHome
	req.UseFollowupEnv = true
	req.CLIArgs = []string{mainRepo}
	return nil
}
```
