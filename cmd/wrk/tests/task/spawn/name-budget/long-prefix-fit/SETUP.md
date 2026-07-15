# Scenario

**Feature**: long basename + long task → create succeeds with fitted names ≤255 bytes

```
repo basename 180×'r' + --task <long>
  -> soft-cap slug 64 would push path base over 255
  -> fit shortens slug; Base/branch ≤255; Base == basename + "-" + branch
```

## Steps

1. Create repo with 180-char basename.
2. `wrk --task` with long description (soft-cap slug 64).
3. Assert byte budgets and invariant.

```go
func Setup(t *testing.T, req *Request) error {
	_, _ = initLongBasenameRepo(t, req, longRepoBasenameLen)
	req.TaskDesc = longTaskDesc()
	return nil
}
```
