# Scenario

**Feature**: prefix alone exceeds name budget → clear error (no silent basename/token chop)

```
repo basename 240×'r' (basename-main-date already > 252 budget)
  wrk --new  # no task
  -> non-zero; clear error; no worktree; do not chop basename/token
```

## Steps

1. Create repo with 240-char basename.
2. Run `wrk --new` (no `--task`).
3. Expect non-zero with budget/name error (not a successful create with truncated basename).

```go
func Setup(t *testing.T, req *Request) error {
	_, _ = initLongBasenameRepo(t, req, overBudgetBasenameLen)
	// No TaskDesc — create via --new (bare no-args is dashboard).
	req.Args = []string{"--new"}
	return nil
}
```
