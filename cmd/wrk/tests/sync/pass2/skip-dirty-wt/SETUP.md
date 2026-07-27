# Scenario

**Feature**: Pass 2 skips distribute when the worktree is dirty

```
# main ahead; feature-login dirty
myrepo + wt/feature-login (dirty) -> wrk --sync
  -> warning: skip feature-login: dirty worktree
  -> no merge into wt
```

## Steps

1. Init main + linked `feature-login`.
2. Commit on main only so main is ahead by 1.
3. Write an uncommitted file in the worktree.
4. Run `wrk --sync` from main.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtPath := initMainWithLinkedBranch(t, req, "feature-login", "wt-feature-login")

	wtBefore := revParseHEAD(t, wtPath)
	mainTip := commitFile(t, mainRepo, "main-ahead.txt", "m\n", "feat: main ahead")
	makeDirty(t, wtPath, "dirty-wt.txt", "uncommitted on wt\n")

	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.WtPath = wtPath
	req.WtBranch = "feature-login"
	req.MainSHA = mainTip
	req.WtSHA = wtBefore
	req.Args = []string{"--sync"}
	return nil
}
```
