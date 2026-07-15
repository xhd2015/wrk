# Scenario

**Feature**: dry-run plans root patch bump when owned files changed

```
# v0.0.1 tagged, README changed at HEAD -> plan v0.0.2, no new tag ref
git repo + tags -> wrk --tag-next --dry-run -> stdout plan only
```

## Steps

1. `setupRootBumpRepo` (v0.0.1 at first commit, README changed at HEAD).
2. Run `wrk --tag-next --dry-run` from the repo.

```go
func Setup(t *testing.T, req *Request) error {
	setupRootBumpRepo(t, req)
	req.Args = []string{"--tag-next", "--dry-run"}
	return nil
}
```