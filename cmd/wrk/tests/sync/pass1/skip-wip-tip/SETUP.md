# Scenario

**Feature**: Pass 1 skips when the tip commit subject is WIP

```
# feature-login tip subject "wip: half done"
myrepo + wt/feature-login -> wrk --sync
  -> warning: skip feature-login: wip commit in range (<short7> wip: half done)
  -> synced: 0 into main, 0 into worktrees, 1 skipped
  -> main HEAD unchanged
```

## Steps

1. Init main + linked `feature-login` worktree.
2. On the worktree, commit with subject `wip: half done`.
3. Record short=7 hash and subject for the warning pin.
4. Run `wrk --sync` from main.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtPath := initMainWithLinkedBranch(t, req, "feature-login", "wt-feature-login")

	const subject = "wip: half done"
	tip := commitFile(t, wtPath, "wip.txt", "half\n", subject)

	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.WtPath = wtPath
	req.WtBranch = "feature-login"
	req.MainSHA = revParseHEAD(t, mainRepo)
	req.WtSHA = tip
	req.WipSubject = subject
	req.WipHashShort = shortSHA(t, wtPath, tip)
	req.Args = []string{"--sync"}
	return nil
}
```
