# Scenario

**Feature**: Dry-run with an ahead worktree prints would: lines and leaves refs unchanged

```
# feature-login +2 clean commits ahead of main
myrepo + wt/feature-login -> wrk --sync --dry-run
  -> would: main ← feature-login  (+2 commits)
  -> would: synced: 1 into main, 0 into worktrees, 0 skipped
  -> main/wt HEAD unchanged
```

## Steps

1. Init main + linked `feature-login`.
2. Two clean commits on the worktree.
3. Record pre-run SHAs.
4. Run `wrk --sync --dry-run` from main.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, wtPath := initMainWithLinkedBranch(t, req, "feature-login", "wt-feature-login")

	commitFile(t, wtPath, "a.txt", "one\n", "feat: one")
	tip := commitFile(t, wtPath, "b.txt", "two\n", "feat: two")

	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.WtPath = wtPath
	req.WtBranch = "feature-login"
	req.MainSHA = revParseHEAD(t, mainRepo)
	req.WtSHA = tip
	req.Args = []string{"--sync", "--dry-run"}
	return nil
}
```
