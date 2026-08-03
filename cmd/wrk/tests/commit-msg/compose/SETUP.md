# Scenario

**Feature**: flag-layer compose of manual --commit -m with primary partners

```
# same compose partners as gen-commit-msg (sample: --done)
wrk --commit -m "x" --done [--dry-run]
  -> must NOT stderr "mutually exclusive"
  -> may later fail: not a linked worktree / nothing staged
```

## Preconditions

- Flag-layer leaves only (done-compose pattern). Full multi-stage apply is out of scope.

## Steps

1. Leaves init main repo via `initGitRepoOnMain` and set Args.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	ensureCommitMsgHelpersUsed()
	return nil
}
```
