# Scenario

**Feature**: multiple recorded projects render lexicographic blocks

```
wrk --add aaa + wrk --add zzz -> wrk --projects -> two blocks, blank line separator
```

## Steps

1. Create repos `{WorkRoot}/aaa` and `{WorkRoot}/zzz`, each with its own bare `origin` remote.
2. Record both via `wrk --add`.
3. Run `wrk --projects`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDetailedStatusHelpersUsed()
	originA := setupBareOrigin(t, req.WorkRoot, "origin-aaa")
	originZ := setupBareOrigin(t, req.WorkRoot, "origin-zzz")
	repoA := setupTrackedMainRepo(t, req.WorkRoot, "aaa", originA, "repo aaa")
	repoZ := setupTrackedMainRepo(t, req.WorkRoot, "zzz", originZ, "repo zzz")
	recordProject(t, req, repoA)
	recordProject(t, req, repoZ)
	req.MainRepo = repoA
	req.SecondRepo = repoZ
	return nil
}
```