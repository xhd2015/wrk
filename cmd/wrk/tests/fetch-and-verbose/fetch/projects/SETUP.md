# Scenario

**Feature**: wrk --projects default no-fetch vs --fetch upstream refresh

```
origin advanced via clone push, main repo not fetched -> --projects Remote: identical (stale)
same fixture + --fetch -> Remote: needs pull
```

## Steps

- Descendants build tracked repos with optional stale `origin/main` tracking refs.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureFetchVerboseHelpersUsed()
	req.RepoDir = req.WorkRoot
	return nil
}
```