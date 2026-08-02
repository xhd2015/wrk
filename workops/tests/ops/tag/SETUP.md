# Scenario

**Feature**: TagNext / TagNextAll plans or applies release tag(s) on main tip

```
# main with prior release tags (root and optional nested scopes)
Caller -> workops.TagNextAll(ctx, {Checkout, DryRun})
  -> Tags[] planned/created; Tag = primary first
```

## Steps

1. Grouping only: set Op to tag-next.
2. Dry-run leaves set DryRun and seed root-bump or multi-scope fixtures.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = OpTagNext
	return nil
}
```
