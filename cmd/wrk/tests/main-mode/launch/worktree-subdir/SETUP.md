# Scenario

**Feature**: wrk --main from a subdirectory of a linked worktree launches shell at main root

```
# cwd is nested inside linked worktree
linked-wt/pkg/nested -> wrk --main
  -> ResolveMainRepo(ShowToplevel(cwd)) = myrepo
  -> fake shell cwd = myrepo
  -> exit 0; minimal UX
```

## Steps

1. Initialize main repo + linked worktree.
2. Create nested subdir under the linked worktree.
3. Install fake bash (exit 0); set cwd to that subdir.
4. Run `wrk --main`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, linkedWT, sub := initLinkedWorktreeSubdir(t, req, "pkg", "nested")
	req.MainRepo = mainRepo
	req.WtDir = linkedWT
	req.RepoDir = sub
	installFakeBash(t, req, 0)
	setMainArgs(req)
	return nil
}
```
