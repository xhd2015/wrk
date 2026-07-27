# Scenario

**Feature**: no saved project and no cwd entry yields normal missing-dir error

```
neutral cwd, no ./basename, no projects.json match -> wrk <basename> -> does not exist
```

## Steps

- Descendants run `wrk <basename>` with no local entry and no matching saved project.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureBasenameFallbackHelpersUsed()
	return nil
}
```