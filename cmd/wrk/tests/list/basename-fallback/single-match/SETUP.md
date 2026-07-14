# Scenario

**Feature**: single saved project match resolves basename for --list

```
saved/myrepo in projects.json
workspace/ (cwd, no ./myrepo) -> wrk myrepo --list -> git worktree list for saved root
```

## Steps

- Descendants record one saved project and run `wrk <basename> --list` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	ensureListBasenameFallbackHelpersUsed()
	return nil
}
```