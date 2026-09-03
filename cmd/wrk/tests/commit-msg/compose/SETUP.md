# Scenario

**Feature**: compose of manual --commit -m with primary/post partners

```
# flag-layer (sample: --done)
wrk --commit -m "x" --done [--dry-run]
  -> must NOT stderr "mutually exclusive"
  -> may later fail: not a linked worktree / nothing staged

# clean-tree soft-skip only when -m already matches HEAD
wrk --commit -m "initial" --exec true   # HEAD is "initial" → notice skip
wrk --commit -m "feat: other" --exec true  # differs → still nothing-to-commit error
```

## Preconditions

- Flag-layer leaves use `initGitRepoOnMain`. Soft-skip leaves use `initCleanGitRepo`.

## Steps

1. Leaves init a git fixture and set Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	ensureCommitMsgHelpersUsed()
	return nil
}
```
