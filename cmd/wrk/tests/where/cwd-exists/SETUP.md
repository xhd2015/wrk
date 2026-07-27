# Scenario

**Feature**: wrk --where ignores local cwd entries

```
# ./basename exists in cwd (even non-git) but saved project differs
wrk --where spl -> stdout = saved path only (no cwd stat fallback)
```

## Steps

- Descendants create a local non-git `./spl` in cwd and a different saved `spl` project.
- Run `wrk --where spl` and assert stdout is the saved path only.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureWhereHelpersUsed()
	return nil
}```
