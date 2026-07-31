# Scenario

**Feature**: `--pr` errors on detached HEAD in a linked worktree

```
# linked worktree checked out detached
linked wt (detached HEAD) + github origin + fake gh
  -> wrk --pr --title T --comment C
  -> non-zero
  -> stderr mentions detached HEAD
```

## Steps

1. Seed linked feature with remote present; detach HEAD in the worktree.
2. Install fake gh; run default `--pr`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupPrDetachedLinked(t, req)
	installFakeGh(t, req)
	req.Args = prDefaultArgs()
	return nil
}
```
