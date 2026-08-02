# Scenario

**Feature**: Push DryRun plans without network or remote ref mutation

```
# DryRun=true
Caller -> Push(..., DryRun) -> err nil; origin tips unchanged
```

## Steps

1. Force DryRun true for this subtree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	return nil
}
```
