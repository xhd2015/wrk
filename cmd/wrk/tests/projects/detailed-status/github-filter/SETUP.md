# Scenario

**Feature**: wrk --projects --github filters to github.com origin remotes

```
wrk --projects --github -> only mains whose origin URL host is github.com
```

## Preconditions

- Git must be available.
- Filter is runtime-only; projects.json is not modified.
- Non-matching and missing-origin paths are omitted silently (exit 0).

## Steps

- Descendants record a mix of remotes, then run `wrk --projects --github`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureDetailedStatusHelpersUsed()
	req.Args = []string{"--projects", "--github"}
	req.RepoDir = req.WorkRoot
	return nil
}

func setOriginURL(t *testing.T, repo, url string) {
	t.Helper()
	runGitIsolated(t, repo, "remote", "set-url", "origin", url)
}

func addOriginURL(t *testing.T, repo, url string) {
	t.Helper()
	runGitIsolated(t, repo, "remote", "add", "origin", url)
}

func ensureGitHubFilterHelpersUsed() {
	_ = setOriginURL
	_ = addOriginURL
}
```
