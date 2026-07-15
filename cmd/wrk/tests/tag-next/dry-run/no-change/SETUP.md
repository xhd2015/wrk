# Scenario

**Feature**: dry-run skips when HEAD matches latest release commit

```
# v0.0.1 at HEAD, no post-tag commits -> skip, 0 tag planned
git repo + tag at HEAD -> wrk --tag-next --dry-run -> all skip
```

## Steps

1. `setupNoChangeRepo` (v0.0.1 tagged at HEAD).
2. Run `wrk --tag-next --dry-run`.

```go
func Setup(t *testing.T, req *Request) error {
	setupNoChangeRepo(t, req)
	req.Args = []string{"--tag-next", "--dry-run"}
	return nil
}
```