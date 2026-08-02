# Scenario

**Feature**: TagNext DryRun returns planned next tag without creating refs

```
# DryRun=true
Caller -> TagNext(..., DryRun) -> planned tag; no git tag ref
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
