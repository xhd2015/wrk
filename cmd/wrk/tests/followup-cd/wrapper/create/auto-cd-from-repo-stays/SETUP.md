# Scenario

**Feature**: wrapper create from main-repo shell cwd does not auto-cd

```
# shell StartDir = main checkout (not user home)
source bash.sh from main repo
wrk -> worktree created; no stderr "cd …"; FinalPWD stays main
```

## Steps

1. Init main repo; start shell cwd there (not FakeHome).
2. Invoke `wrk --new` via installed wrapper (auto-cd on, home gate closed).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.StartDir = mainRepo
	req.CLIArgs = []string{"--new"}
	return nil
}
```
