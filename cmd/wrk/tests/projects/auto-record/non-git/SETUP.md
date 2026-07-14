# Scenario

**Feature**: no auto-record when cwd is not a git repository

```
non-git cwd -> wrk --list -> error; no projects.json created
```

## Steps

1. Use `{WorkRoot}` as a plain directory (not git).
2. Run `wrk --list` from `{WorkRoot}`.

```go
func Setup(t *testing.T, req *Request) error {
	req.RepoDir = req.WorkRoot
	return nil
}
```