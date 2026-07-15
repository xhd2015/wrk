# Scenario

**Feature**: aborted `--done` does not run `--sync`

```
# user declines confirm → merge-back aborted; no sync stdout
myrepo + wtA (ahead)
  -> wrk --done --confirm-from-stdin --sync  (stdin: n)
  -> merge-back aborted
  -> no "synced:" line; wtA remains
```

## Steps

1. Create wrk-managed linked worktree and commit ahead.
2. Run `wrk --done --confirm-from-stdin --sync` with `n\n` on stdin.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	_, wtDir, _ := setupWrkWorktreeFromMain(t, req)
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
	req.Args = []string{"--done", "--confirm-from-stdin", "--sync"}
	req.StdinInput = "n\n"
	return nil
}
```
