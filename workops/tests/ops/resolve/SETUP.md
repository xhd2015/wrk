# Scenario

**Feature**: WhereMain resolves main repository absolute path from a checkout

```
# checkout is main or linked worktree
Caller -> workops.WhereMain(checkout) -> mainAbs
```

## Steps

1. Grouping only: set Op to where-main.
2. Leaves seed main and/or linked worktree and set Checkout.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = OpWhereMain
	return nil
}
```
