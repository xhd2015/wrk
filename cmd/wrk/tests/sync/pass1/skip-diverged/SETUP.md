# Scenario

**Feature**: Pass 1 skips when main and worktree branch have diverged

```
# main and feature-login each have unique commits (diverged)
myrepo + wt/feature-login -> wrk --sync
  -> warning: skip feature-login: diverged from main
  -> no merge; HEADs unchanged
```

## Steps

1. Init main + linked `feature-login` at shared base.
2. Commit on main (`main-only`).
3. Commit on the worktree (`wt-only`) so histories diverge.
4. Run `wrk --sync` from main.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtPath := initMainWithLinkedBranch(t, req, "feature-login", "wt-feature-login")

	mainTip := commitFile(t, mainRepo, "main-only.txt", "m\n", "main-only")
	wtTip := commitFile(t, wtPath, "wt-only.txt", "w\n", "wt-only")

	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.WtPath = wtPath
	req.WtBranch = "feature-login"
	req.MainSHA = mainTip
	req.WtSHA = wtTip
	req.Args = []string{"--sync"}
	return nil
}
```
