# Scenario

**Feature**: bare `--exec` without a command errors at parse

```
workspace/ -> wrk --exec -> non-zero; requires command; no worktree created
```

## Steps

1. Run `wrk --exec` with no trailing tokens from WorkRoot.

```go
func Setup(t *testing.T, req *Request) error {
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--exec"}
	return nil
}
```
