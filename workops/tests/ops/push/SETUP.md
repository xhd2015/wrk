# Scenario

**Feature**: Push publishes branch and/or tags from a checkout

```
# checkout with optional origin
Caller -> workops.Push(ctx, {Checkout, DryRun, Tags}) -> err
```

## Steps

1. Grouping only: set Op to push.
2. Dry-run leaves set DryRun and seed origin fixtures.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = OpPush
	return nil
}
```
