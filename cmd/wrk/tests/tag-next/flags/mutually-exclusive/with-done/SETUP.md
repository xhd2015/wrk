# Scenario

**Feature**: --tag-next and --done are mutually exclusive

```
# wrk --tag-next --done -> non-zero, stderr mutually exclusive
wrk --tag-next --done -> error before tagscope
```

## Steps

1. `setupRootBumpRepo`.
2. Run `wrk --tag-next --done` from the repo (not a linked worktree — mode clash fires first).

```go
func Setup(t *testing.T, req *Request) error {
	setupRootBumpRepo(t, req)
	req.Args = []string{"--tag-next", "--done"}
	return nil
}
```