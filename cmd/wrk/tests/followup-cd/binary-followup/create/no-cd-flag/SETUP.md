# Scenario

**Feature**: create with --no-cd suppresses follow-up write even when env set and cwd is home

```
# shell cwd = FakeHome (home gate would otherwise open)
FakeHome (cwd) + WRK_FOLLOWUP_FILE=tmp
wrk --no-cd <mainRepo> -> worktree created; follow-up file empty
```

## Steps

1. Init main repo.
2. Set process cwd to FakeHome.
3. Run `wrk --no-cd <mainRepo>` with follow-up env set (`--no-cd` independent of home gate).

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = req.FakeHome
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--no-cd", mainRepo}
	return nil
}
```
