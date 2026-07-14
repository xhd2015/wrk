# Scenario

**Feature**: wrk --done with explicit directory argument

```
# wrk --done <dir> merges and removes the worktree at <dir>, same as cd <dir> && wrk --done
wrk --done <wtDir> -> merge-back --rm via worktree.MergeBack
wrk --done <wtDir> --confirm-from-stdin -> ff-merge + remove (ahead confirm)
wrk --done <nonexistent> -> non-zero, directory does not exist
wrk --done <mainRepo> -> non-zero, not a linked worktree
```

## Preconditions

- Git must be available.
- `wrk` binary must be built (shared with parent tree).

## Steps

- Process cwd is `{WorkRoot}` — not the worktree dir — to verify the dir arg sets effective cwd.
- `req.Args` includes `"--done"` plus the worktree path as a positional argument (e.g., `["--done", req.WtDir]`).
- For confirmation tests, `req.StdinInput` and `"--confirm-from-stdin"` follow the same pattern as existing done tests.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```