# Scenario

**Feature**: wrapper --done on another worktree while shell cwd is a surviving sibling does not auto-cd

```
# shell cwd = linked wt A; wrapper removes sibling B
source bash.sh from wtA; wrk --done <wtB>
  -> B removed; no stderr cd; FinalPWD stays wtA
```

## Steps

1. Create main + two linked worktrees; commit/ff-merge B only (already-included).
2. Run wrapper `wrk --done <wtB>` with StartDir = wtA.

```go
func Setup(t *testing.T, req *Request) error {
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
	req.RepoDir = wtA
	req.StartDir = wtA
	req.CLIArgs = []string{"--done", wtB}
	return nil
}
```
