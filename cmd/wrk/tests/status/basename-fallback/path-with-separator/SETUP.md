# Scenario

**Feature**: path with separator is not a basename — no projects.json lookup for --status

```
# <dir> contains '/' -> treat as normal path, no fallback
wrk saved/myrepo --status (missing) -> does not exist
```

## Steps

- Descendants pass a relative path with a separator as `<dir>`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureStatusBasenameFallbackHelpersUsed()
	return nil
}
```
