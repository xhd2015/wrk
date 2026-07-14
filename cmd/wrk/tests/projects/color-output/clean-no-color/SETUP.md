# Scenario

**Feature**: --color on an all-clean project does not apply red or orange highlights

```
clean main, up-to-date remote, zero dirty worktrees -> wrk --projects --color -> plain values
```

## Steps

1. Create tracked clean repo `{WorkRoot}/allclean`.
2. Record and run `wrk --projects --color`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureColorOutputHelpersUsed()
	withProjectsColor(req)
	origin := setupColorBareOrigin(t, req.WorkRoot, "origin")
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "allclean", origin, "all clean")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```