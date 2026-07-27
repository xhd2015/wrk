# Scenario

**Feature**: wrk --done on a nonexistent directory should fail

```
# wrk --done on a path that doesn't exist
wrk --done /nonexistent/path -> non-zero, "does not exist" or "not a git repository"
```

## Steps

1. Run `wrk --done /nonexistent/path` from WorkRoot.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--done", "/nonexistent/path"}
	return nil
}
```