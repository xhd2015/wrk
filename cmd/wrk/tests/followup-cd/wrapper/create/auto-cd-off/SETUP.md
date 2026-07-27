# Scenario

**Feature**: WRK_AUTO_CD=0 disables wrapper follow-up and cd even from home

```
# StartDir = FakeHome; home gate would open if channel were exported
WRK_AUTO_CD=0; source bash.sh; wrk <mainRepo>
  -> worktree created; no stderr cd; FinalPWD stays FakeHome
```

## Steps

1. Init main repo; start shell at FakeHome.
2. Run wrapper create with `WRK_AUTO_CD=0`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.StartDir = req.FakeHome
	req.AutoCD = "0"
	req.CLIArgs = []string{mainRepo}
	return nil
}
```
