# Scenario

**Feature**: wrk --propagate-tags requires a git work tree

```
# non-git cwd -> error before planning
neutral dir -> wrk --propagate-tags --dry-run -> not a git repository
```

## Steps

1. Set `req.RepoDir` to a neutral directory without `.git`.
2. Run `wrk --propagate-tags --dry-run`.

```go
func Setup(t *testing.T, req *Request) error {
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "not-git")
	req.Args = []string{"--propagate-tags", "--dry-run"}
	return nil
}
```
