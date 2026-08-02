# Scenario

**Feature**: TagNext plans or applies the next release tag on main tip

```
# main with prior release tags
Caller -> workops.TagNext(ctx, {Checkout, DryRun}) -> tag string
```

## Steps

1. Grouping only: set Op to tag-next.
2. Dry-run leaves set DryRun and seed root-bump fixtures.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = OpTagNext
	return nil
}
```
