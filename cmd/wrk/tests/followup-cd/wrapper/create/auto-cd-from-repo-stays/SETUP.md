# Scenario

**Feature**: wrapper create from main-repo shell cwd does not auto-cd

```
# shell StartDir = main checkout (not user home)
source bash.sh from main repo
wrk -> worktree created; no stderr "cd …"; FinalPWD stays main
```

## Steps

1. Init main repo; start shell cwd there (not FakeHome).
2. Invoke bare `wrk` via installed wrapper (auto-cd on, home gate closed).

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.StartDir = mainRepo
	req.CLIArgs = nil
	return nil
}
```
