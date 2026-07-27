# Scenario

**Feature**: wrapper create from user home auto-cds into new worktree

```
# shell StartDir = FakeHome; create via positional main repo
source bash.sh from FakeHome
wrk <mainRepo> -> stderr "cd <worktree>"; FinalPWD = worktree; exit 0
```

## Steps

1. Init main repo; start shell cwd at FakeHome (test user home).
2. Invoke `wrk <mainRepo>` via installed wrapper (default auto-cd on).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	// Shell cwd is exact user home so create home gate opens.
	req.StartDir = req.FakeHome
	req.CLIArgs = []string{mainRepo}
	return nil
}
```
