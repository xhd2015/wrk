# Scenario

**Feature**: no saved project and no cwd entry yields normal missing-dir error for --status

```
neutral cwd, no ./basename, no projects.json match -> wrk <basename> --status -> does not exist
```

## Steps

- Descendants run `wrk <basename> --status` with no local entry and no matching saved project.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureStatusBasenameFallbackHelpersUsed()
	return nil
}
```
