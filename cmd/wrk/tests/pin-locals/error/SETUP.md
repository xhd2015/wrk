# Scenario

**Feature**: wrk --pin-locals hard-fails when cwd is not a git repository

```
plain non-git dir -> wrk --pin-locals -> Error: not a git repository; non-zero
```

## Steps

- Descendants use non-git RepoDir under WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensurePinLocalsHelpersUsed()
	return nil
}
```
