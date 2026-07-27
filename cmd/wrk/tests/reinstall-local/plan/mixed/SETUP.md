# Scenario

**Feature**: multi-candidate plans combine filter and stable sort

```
# several cmd/script candidates → sorted by BinName; install vs skip per binDir
mixed discovery + partial binDir
  -> Items sorted lexicographically; Action install|skip each
```

## Preconditions

- Leaves create multiple distinct bin names with mixed presence in binDir.

## Steps

1. Leaves write multi-bin fixtures.
2. Assert full ordered Items list.

## Context

- Sort is by BinName only (not by method or path).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.WantItems == nil {
		req.WantItems = []WantPlanItem{}
	}
	return nil
}
```
