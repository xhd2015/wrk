# Scenario

**Feature**: path with separator is not a basename — no projects.json lookup

```
# <dir> contains '/' -> treat as normal path, no fallback
wrk sub/foo (missing) -> does not exist
```

## Steps

- Descendants pass a relative path with a separator as `<dir>`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureBasenameFallbackHelpersUsed()
	return nil
}
```