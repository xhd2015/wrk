# Scenario

**Feature**: no auto-record when cwd is not a git repository

```
non-git cwd -> wrk --list -> error; no projects.json created
```

## Steps

1. Use `{WorkRoot}` as a plain directory (not git).
2. Run `wrk --list` from `{WorkRoot}`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.RepoDir = req.WorkRoot
	return nil
}
```