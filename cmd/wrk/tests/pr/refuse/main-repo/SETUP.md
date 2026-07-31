# Scenario

**Feature**: `--pr` rejects main repo checkout (linked worktree only)

```
# cwd is main repo checkout, not a linked worktree
myrepo (main) + github origin
  -> wrk --pr --title T --comment C
  -> non-zero
  -> stderr names --pr and linked worktree (same pattern as --done)
```

## Steps

1. Seed main with github-shaped origin (no linked wt as cwd).
2. Install fake gh (should not matter; refuse before create).
3. Run `--pr` from main repo root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrMainWithGithubOrigin(t, req)
	installFakeGh(t, req)
	req.RepoDir = req.MainRepo
	req.Args = prDefaultArgs()
	return nil
}
```
