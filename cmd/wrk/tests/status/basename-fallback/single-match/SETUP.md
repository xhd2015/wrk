# Scenario

**Feature**: single saved project match resolves basename for --status

```
# one projects.json entry matches basename; cwd has no local ./basename
neutral cwd -> wrk myrepo --status -> status blocks for saved project root
```

## Steps

- Descendants seed exactly one saved project whose basename matches `<dir>`.
- Run `wrk <basename> --status` from a cwd without a local `./<basename>` entry.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureStatusBasenameFallbackHelpersUsed()
	return nil
}
```
