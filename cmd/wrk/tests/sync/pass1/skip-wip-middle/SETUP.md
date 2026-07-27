# Scenario

**Feature**: Pass 1 skips when an older unique commit in the range is WIP (tip clean)

```
# range main..feature-login has middle "wip: half done", tip "feat: clean tip"
myrepo + wt/feature-login -> wrk --sync
  -> warning names first wip in range (not the clean tip)
  -> skip pass1; main unchanged
```

## Steps

1. Init main + linked `feature-login` worktree.
2. On the worktree: commit `wip: half done`, then `feat: clean tip`.
3. Record short hash of the **first** WIP commit (the older one) for the warning.
4. Run `wrk --sync` from main.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtPath := initMainWithLinkedBranch(t, req, "feature-login", "wt-feature-login")

	const wipSubject = "wip: half done"
	wipSHA := commitFile(t, wtPath, "wip.txt", "half\n", wipSubject)
	tip := commitFile(t, wtPath, "clean.txt", "ok\n", "feat: clean tip")

	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.WtPath = wtPath
	req.WtBranch = "feature-login"
	req.MainSHA = revParseHEAD(t, mainRepo)
	req.WtSHA = tip
	req.WipSubject = wipSubject
	req.WipHashShort = shortSHA(t, wtPath, wipSHA)
	req.Args = []string{"--sync"}
	return nil
}
```
