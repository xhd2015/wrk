# Scenario

**Feature**: compose with --push publishes new tag to origin then propagates consumers

```
# root-bump lib + bare origin; app requires older lib
cwd=lib -> wrk --tag-next --propagate-tags --push
  -> (1) tag-next apply: create v1.0.1 locally
  -> push v1.0.1 to origin (existing --tag-next --push behavior)
  -> blank line
  -> (2) propagate apply: bump app to v1.0.1 + build + commit
  -> origin has refs/tags/v1.0.1; app require at v1.0.1
```

## Steps

1. `setupComposeRootBump` with origin (bare remote + push main).
2. Args: `--tag-next --propagate-tags --push` (order free; push is tag-next stage).

```go
func Setup(t *testing.T, req *Request) error {
	setupComposeRootBump(t, req, true)
	req.Args = []string{"--tag-next", "--propagate-tags", "--push"}
	return nil
}
```
