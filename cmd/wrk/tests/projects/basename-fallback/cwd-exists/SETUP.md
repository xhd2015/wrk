# Scenario

**Feature**: local cwd entry blocks basename fallback

```
# ./<basename> exists in cwd (even non-git) -> resolve to cwd path, no projects.json lookup
cwd/myrepo (non-git) + saved myrepo elsewhere -> wrk myrepo -> git error, no fallback
```

## Steps

- Descendants create a non-git `./<basename>` in cwd and a saved project with the same basename elsewhere.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureBasenameFallbackHelpersUsed()
	return nil
}
```