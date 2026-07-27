# Scenario

**Feature**: successful create without WRK_FOLLOWUP_FILE leaves no follow-up side effects

```
# shell cwd = FakeHome; channel not opened
FakeHome (cwd); WRK_FOLLOWUP_FILE unset
wrk <mainRepo> -> stdout worktree path; pre-created followup.txt stays empty
```

## Steps

1. Init main repo.
2. Prepare empty follow-up path but do **not** export WRK_FOLLOWUP_FILE.
3. Run `wrk <mainRepo>` from FakeHome (home would open the gate if env were set).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = req.FakeHome
	req.UseFollowupEnv = false
	req.FollowupFile = defaultFollowupPath(req)
	req.CLIArgs = []string{mainRepo}
	return nil
}
```
