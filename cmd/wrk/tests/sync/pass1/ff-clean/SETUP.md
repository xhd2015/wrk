# Scenario

**Feature**: Pass 1 FF harvest when worktree is strictly ahead with clean subjects

```
# main at base; feature-login +2 clean commits
myrepo + wt/feature-login -> wrk --sync
  -> main ← feature-login  (+2 commits)
  -> synced: 1 into main, 0 into worktrees, 0 skipped
  -> main HEAD == feature-login tip
```

## Steps

1. Init main + linked worktree on branch `feature-login` at the same init tip.
2. On the worktree, add two non-WIP commits (`feat: one`, `feat: two`).
3. Record pre-run main/wt SHAs; run `wrk --sync` from main.

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
	req.Args = []string{"--sync"}
	return nil
}
```
