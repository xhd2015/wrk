# Scenario

**Feature**: wrk --version prints embedded build version on stdout

```
embedded VERSION.txt in wrk binary
workspace/ -> wrk --version -> stdout v0.0.1\n
```

## Steps

- Descendants run `wrk --version` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ensureVersionHelpersUsed()
	return nil
}
```