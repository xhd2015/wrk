# Scenario

**Feature**: wrapper create from main-repo shell cwd with --force-cd auto-cds into worktree

```
# shell StartDir = main checkout (not user home)
source bash.sh from main repo
wrk --force-cd -> worktree created; stderr "cd <wt>"; FinalPWD = worktree
```

## Steps

1. Init main repo; start shell cwd there (not FakeHome).
2. Invoke `wrk --force-cd` via installed wrapper (auto-cd on; binary bypasses home gate).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.StartDir = mainRepo
	req.CLIArgs = []string{"--force-cd"}
	return nil
}
```
