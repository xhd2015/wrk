# Scenario

**Feature**: --push publishes newly created tag to bare origin

```
# root bump + origin remote -> wrk --tag-next --push -> tag on origin
git repo + bare origin -> wrk --tag-next --push -> refs/tags/v0.0.2 on origin
```

## Steps

1. `setupPushRepo` (repo with origin bare remote).
2. Run `wrk --tag-next --push`.

```go
func Setup(t *testing.T, req *Request) error {
	setupPushRepo(t, req)
	req.Args = []string{"--tag-next", "--push"}
	return nil
}
```