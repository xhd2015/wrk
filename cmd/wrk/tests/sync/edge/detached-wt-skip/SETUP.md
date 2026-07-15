# Scenario

**Feature**: Linked worktree with detached HEAD is skipped with a warning

```
# linked wt checked out --detach
myrepo + detached wt -> wrk --sync
  -> warning: skip <path>: detached HEAD
  -> exit 0; main unchanged; skipped=1
```

## Steps

1. Init main + linked `feature-login` worktree.
2. In the worktree: `git checkout --detach`.
3. Run `wrk --sync` from main.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, wtPath := initMainWithLinkedBranch(t, req, "feature-login", "wt-feature-login")
	runGitIsolated(t, wtPath, "checkout", "--detach")

	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.WtPath = wtPath
	req.WtBranch = "feature-login"
	req.MainSHA = revParseHEAD(t, mainRepo)
	req.WtSHA = revParseHEAD(t, wtPath)
	req.Args = []string{"--sync"}
	return nil
}
```
