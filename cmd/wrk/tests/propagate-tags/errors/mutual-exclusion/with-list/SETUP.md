# Scenario

**Feature**: wrk --propagate-tags --list is mutually exclusive

```
workspace/ -> wrk --propagate-tags --list
  -> non-zero, mutually exclusive
```

## Steps

1. Use neutral cwd (git not required for flag-layer exclusion).
2. Run `--propagate-tags` with `--list`.

```go
func Setup(t *testing.T, req *Request) error {
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	req.Args = []string{"--propagate-tags", "--list"}
	return nil
}
```
