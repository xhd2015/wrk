# Scenario

**Feature**: wrapper respects --no-cd (no auto-cd) even from home

```
# StartDir = FakeHome; home gate would open without --no-cd
source bash.sh; wrk --no-cd <mainRepo>
  -> worktree created; no stderr cd; FinalPWD stays FakeHome
```

## Steps

1. Init main repo; start shell at FakeHome.
2. Run `wrk --no-cd <mainRepo>` via wrapper.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.StartDir = req.FakeHome
	req.CLIArgs = []string{"--no-cd", mainRepo}
	return nil
}
```
