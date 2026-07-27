# Scenario

**Feature**: wrk --done --confirm with piped confirmation (opt-in prompt)

```
# wt branch ahead of main; --confirm restores Y/n; --confirm-from-stdin + Enter
myrepo + wt -> commit on wt -> wrk --done --confirm --confirm-from-stdin (\n) -> ff-merge + remove
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Commit on worktree so branch is ahead of main.
3. Run `wrk --done --confirm --confirm-from-stdin` with `\n` on stdin.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	req.RepoDir = wtDir
	req.Args = []string{"--done", "--confirm", "--confirm-from-stdin"}
	req.StdinInput = "\n"
	return nil
}
```
