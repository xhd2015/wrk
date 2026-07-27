# Scenario

**Feature**: single saved project match resolves basename for --repos

```
saved/myrepo in projects.json
workspace/ (cwd, no ./myrepo) -> wrk myrepo --repos -> repo paths for saved root
```

## Steps

- Descendants record one saved project and run `wrk <basename> --repos` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureReposBasenameFallbackHelpersUsed()
	return nil
}
```