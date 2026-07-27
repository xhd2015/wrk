# Scenario

**Feature**: wrk --main from a linked worktree root launches shell at main repo root

```
# cwd is linked worktree root (not main)
linked-wt (linked of myrepo) -> wrk --main
  -> fake shell cwd = myrepo (main root)
  -> exit 0; minimal UX
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Add linked worktree at `{WorkRoot}/linked-wt` on branch `linked-side`.
3. Install fake bash (exit 0); set cwd to the linked worktree root.
4. Run `wrk --main`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, linkedWT := initLinkedWorktree(t, req)
	req.MainRepo = mainRepo
	req.RepoDir = linkedWT
	installFakeBash(t, req, 0)
	setMainArgs(req)
	return nil
}
```
