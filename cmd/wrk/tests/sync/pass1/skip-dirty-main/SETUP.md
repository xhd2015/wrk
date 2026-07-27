# Scenario

**Feature**: Pass 1 skips harvest when main worktree is dirty

```
# feature-login clean + ahead; main has uncommitted file
myrepo (dirty) + wt/feature-login -> wrk --sync
  -> warning: skip feature-login: dirty main
  -> no merge into main
```

## Steps

1. Init main + linked `feature-login`.
2. On worktree: one clean ahead commit (`feat: ahead`).
3. On main: write an uncommitted dirty file (do not commit).
4. Run `wrk --sync` from main.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtPath := initMainWithLinkedBranch(t, req, "feature-login", "wt-feature-login")

	wtTip := commitFile(t, wtPath, "ahead.txt", "a\n", "feat: ahead")
	makeDirty(t, mainRepo, "dirty-main.txt", "uncommitted on main\n")

	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.WtPath = wtPath
	req.WtBranch = "feature-login"
	req.MainSHA = revParseHEAD(t, mainRepo)
	req.WtSHA = wtTip
	req.Args = []string{"--sync"}
	return nil
}
```
