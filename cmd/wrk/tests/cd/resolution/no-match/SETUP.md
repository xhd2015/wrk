# Scenario

**Feature**: wrk --cd basename with zero projects matches errors

```
no saved nosuch; no ./nosuch under cwd
wrk --cd nosuch -> non-zero; does not exist
```

## Steps

1. Neutral cwd; no projects.json entries for `nosuch`.
2. `wrk --cd nosuch`.

```go
func Setup(t *testing.T, req *Request) error {
	setCDFlagThenPath(req, "nosuch")
	return nil
}
```
