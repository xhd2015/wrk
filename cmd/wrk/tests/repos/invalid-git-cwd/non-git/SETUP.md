# Scenario

**Feature**: wrk --repos reports a clear error for non-git cwd

```
plain cwd -> wrk --repos -> non-zero stderr
```

## Steps

1. Use `{WorkRoot}` without initializing git.
2. Run `wrk --repos` from that directory.

```go
func Setup(t *testing.T, req *Request) error {
	req.RepoDir = req.WorkRoot
	return nil
}
```
