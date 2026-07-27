# Scenario

**Feature**: single saved project match resolves basename to saved path

```
# one projects.json entry matches basename; cwd has no local ./basename
neutral cwd -> wrk myrepo -> create from saved project path
```

## Steps

- Descendants seed exactly one saved project whose basename matches `<dir>`.
- Run `wrk <basename>` from a cwd without a local `./<basename>` entry.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureBasenameFallbackHelpersUsed()
	return nil
}
```