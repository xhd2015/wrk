# Scenario

**Feature**: apply skips when newest tag is a prerelease at HEAD

```
# v0.0.2 latest release, v0.0.3-alpha at HEAD -> prerelease-head skip
git repo + prerelease head -> wrk --tag-next -> 0 tag created
```

## Steps

1. `setupPrereleaseSkipRepo`.
2. Run `wrk --tag-next`.

```go
func Setup(t *testing.T, req *Request) error {
	setupPrereleaseSkipRepo(t, req)
	req.Args = []string{"--tag-next"}
	return nil
}
```