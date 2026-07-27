# Scenario

**Feature**: --done with --force-cd from sibling, channel closed, launches shell at main (Branch B)

```
wtA (cwd) + wtB already-included; WRK_FOLLOWUP_FILE unset; fake bash
wrk --done <wtB> --force-cd
  -> B removed; stdout still done message
  -> stderr install hint
  -> fake shell cwd = main repo
```

## Steps

1. Create two worktrees; already-include B; install fake bash.
2. Run `--done <wtB> --force-cd` with cwd = wtA and no follow-up env.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := setupMainRepo(t, req)
	wtA := runWrkWithArgs(t, req, mainRepo)
	wtB := runWrkWithArgs(t, req, mainRepo)
	branchB := branchName("main", wrkDate, 1)
	commitAheadOnWorktree(t, wtB, "b-work", "sibling b")
	runGitIsolated(t, mainRepo, "merge", "--ff-only", branchB)

	req.MainRepo = mainRepo
	req.WtDir = wtB
	req.RepoDir = wtA
	req.UseFollowupEnv = false
	req.CLIArgs = []string{"--done", wtB, "--force-cd"}
	installFakeBash(t, req, 0)
	return nil
}
```
