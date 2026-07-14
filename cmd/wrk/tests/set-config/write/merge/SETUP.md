# Scenario

**Feature**: set-config merges only keys implied by flags; preserves siblings

```
sequential --set-config --create writes -> union of keys; no wipe of other create.*
```

## Steps

- Leaves seed config and/or run multiple writes.

```go
func Setup(t *testing.T, req *Request) error {
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	return nil
}
```
