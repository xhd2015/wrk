# Scenario

**Feature**: prefix alone exceeds name budget → clear error (no silent basename/token chop)

```
repo basename 240×'r' (basename-main-date already > 252 budget)
  wrk  # no task
  -> non-zero; clear error; no worktree; do not chop basename/token
```

## Steps

1. Create repo with 240-char basename.
2. Run bare `wrk` (no `--task`).
3. Expect non-zero with budget/name error (not a successful create with truncated basename).

```go
func Setup(t *testing.T, req *Request) error {
	_, _ = initLongBasenameRepo(t, req, overBudgetBasenameLen)
	// No TaskDesc — bare create.
	return nil
}
```
