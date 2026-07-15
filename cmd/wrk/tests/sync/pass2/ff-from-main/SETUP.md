# Scenario

**Feature**: Pass 2 FF distribute when main is strictly ahead of a clean worktree

```
# feature-login at base; main +1 clean commit
myrepo + wt/feature-login -> wrk --sync
  -> feature-login ← main  (+1 commit)
  -> synced: 0 into main, 1 into worktrees, 0 skipped
  -> wt HEAD == main tip
```

## Steps

1. Init main + linked `feature-login` at shared base.
2. Commit on main only (`feat: main ahead`) so main is ahead by 1.
3. Worktree stays clean and behind.
4. Run `wrk --sync` from main.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, wtPath := initMainWithLinkedBranch(t, req, "feature-login", "wt-feature-login")

	// Snapshot wt tip before main advances.
	wtBefore := revParseHEAD(t, wtPath)
	mainTip := commitFile(t, mainRepo, "main-ahead.txt", "m\n", "feat: main ahead")

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
