# Scenario

**Feature**: saved create.* config applies without CLI UX flags

```
config create.* + bare wrk [-t] -> same effective UX as matching flags
flags --no-* override config for this run
```

## Steps

- Seed config; leaves run create.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	return nil
}
```
