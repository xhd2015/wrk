# Scenario

**Feature**: bare `wrk --pr` show mode refuses main-repo checkout (linked worktree only)

```
# cwd is main repo checkout, not a linked worktree
myrepo (main) + github origin
  -> wrk --pr
  -> non-zero
  -> stderr names linked worktree (and preferably --pr)
  -> no gh create/comment
```

## Steps

1. Seed main with github-shaped origin (no linked wt as cwd).
2. Install fake gh (should not matter; refuse before list/create).
3. Run bare `--pr` from main repo root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrMainWithGithubOrigin(t, req)
	installFakeGh(t, req)
	req.RepoDir = req.MainRepo
	req.Args = prShowArgs()
	return nil
}
```
