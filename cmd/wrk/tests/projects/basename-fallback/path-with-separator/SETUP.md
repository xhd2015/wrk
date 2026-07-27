# Scenario

**Feature**: path with separator is not a basename — no projects.json lookup

```
# <dir> contains '/' -> treat as normal path, no fallback
wrk sub/foo (missing) -> does not exist
```

## Steps

- Descendants pass a relative path with a separator as `<dir>`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureBasenameFallbackHelpersUsed()
	return nil
}
```