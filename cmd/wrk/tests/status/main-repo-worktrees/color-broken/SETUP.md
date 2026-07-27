# Scenario

**Feature**: --color highlights red error value on appended broken block

```
# broken external wt + wrk --status --color (pipe-safe)
external wt (broken) + --color -> red error: value on appended block
```

## Steps

1. Create external wrk worktree from main repo.
2. Break git metadata on the external worktree.
3. Run `wrk --status --color` from the main repo root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtDir, _ := createExternalWrkWorktree(t, req)
	breakWorktreeGitMetadata(t, req, wtDir)
	req.RepoDir = mainRepo
	req.Args = []string{"--status", "--color"}
	return nil
}
```