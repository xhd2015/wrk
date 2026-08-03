# Scenario

**Feature**: manual --commit -m refuses when branch is checked out in multiple worktrees

```
# two live checkouts of same branch + staged
  -> wrk --commit -m "x"
  -> non-zero; Error: refuse … commit
  -> HEAD unchanged
```

## Steps

1. Grouping: leaves use `setupSharedTwoLinkedStaged`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	ensureCommitMsgHelpersUsed()
	return nil
}
```
