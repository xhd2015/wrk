# Scenario

**Feature**: wrk --all-deps errors when cwd is not a git repository

```
# cwd is a plain directory (no .git) -> wrk --all-deps -> non-zero, is not a git repository
plain dir (no .git) -> wrk --all-deps -> error (is not a git repository)
```

## Steps

1. Create a non-git temp dir as the cwd.
2. Run `wrk --all-deps` from the non-git cwd.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	req.RepoDir = t.TempDir()
	req.Args = []string{"--all-deps"}
	return nil
}
```