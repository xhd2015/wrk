# Scenario

**Feature**: branch without upstream shows (no upstream)

```
main repo without tracking branch -> Remote: (no upstream)
```

## Steps

1. Create git repo `{WorkRoot}/noup` without `git push -u`.
2. Record and run `wrk --projects`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureDetailedStatusHelpersUsed()
	repo := initProjectsRepo(t, req.WorkRoot, "noup")
	recordProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```