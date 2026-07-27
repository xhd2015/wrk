# Scenario

**Feature**: --done on another worktree while shell cwd is a surviving sibling writes no follow-up

```
# shell cwd = linked wt A (still valid after); operate on sibling B
wtA (cwd) + wtB already-included + WRK_FOLLOWUP_FILE
wrk --done <wtB> -> B removed; follow-up empty (A still exists)
```

## Steps

1. Create main repo and two linked worktrees (A then B).
2. Commit on B and ff-merge B into main (already-included remove path).
3. Run `wrk --done <wtB>` with process cwd = wtA and follow-up env set.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := setupMainRepo(t, req)
	// A is only the surviving shell cwd; B is the operated tree.
	wtA := runWrkWithArgs(t, req, mainRepo)
	wtB := runWrkWithArgs(t, req, mainRepo)
	branchB := branchName("main", wrkDate, 1)
	// Make B already-included so --done removes without confirm prompts.
	commitAheadOnWorktree(t, wtB, "b-work", "sibling b")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branchB)

	req.MainRepo = mainRepo
	req.WtDir = wtB
	// Process/shell cwd remains sibling A (not the operated tree).
	req.RepoDir = wtA
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--done", wtB}
	return nil
}
```
