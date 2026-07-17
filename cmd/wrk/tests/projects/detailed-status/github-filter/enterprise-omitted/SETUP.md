# Scenario

**Feature**: github enterprise host is not github.com — omitted

```
origin https://github.mycorp.com/o/r -> wrk --projects --github -> empty
```

## Steps

1. Create tracked repo, set origin to enterprise host.
2. Record and run `wrk --projects --github`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureGitHubFilterHelpersUsed()
	origin := setupBareOrigin(t, req.WorkRoot, "origin-ent")
	repo := setupTrackedMainRepo(t, req.WorkRoot, "ent-repo", origin, "enterprise project")
	setOriginURL(t, repo, "https://github.mycorp.com/example/ent-repo.git")
	recordProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```
