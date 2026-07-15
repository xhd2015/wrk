# Scenario

**Feature**: wrk --tag-next --push publishes new tags to origin

```
# Apply + git push origin <tag> for each created tag
wrk --tag-next --push -> local tag + tag ref on bare origin
```

## Preconditions

- Repo has `origin` remote pointing at a bare repository.

## Steps

- Descendants use `setupPushRepo` and `req.Args = []string{"--tag-next", "--push"}`.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	tagNextEnsureHelpersUsed()
	return nil
}
```