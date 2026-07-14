# Scenario

**Feature**: --done with --force-cd from a surviving sibling writes follow-up to main (cwd gate bypass)

```
# shell cwd = linked wt A (still valid after); operate on sibling B
wtA (cwd) + wtB already-included + WRK_FOLLOWUP_FILE
wrk --done <wtB> --force-cd -> B removed; follow-up: cd <main-abs>
```

## Steps

1. Create main repo and two linked worktrees (A then B).
2. Commit on B and ff-merge B into main (already-included remove path).
3. Run `wrk --done <wtB> --force-cd` with process cwd = wtA and follow-up env set.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	wtA := runWrkWithArgs(t, req, mainRepo)
	wtB := runWrkWithArgs(t, req, mainRepo)
	branchB := branchName("main", wrkDate, 1)
	commitAheadOnWorktree(t, wtB, "b-work", "sibling b")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branchB)

	req.MainRepo = mainRepo
	req.WtDir = wtB
	// Process/shell cwd remains sibling A (cwd-missing gate would normally skip).
	req.RepoDir = wtA
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--done", wtB, "--force-cd"}
	return nil
}
```
