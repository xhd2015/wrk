# Scenario

**Feature**: scp-like git@github.com origin is included

```
origin git@github.com:owner/repo.git -> wrk --projects --github -> shown
```

## Steps

1. Create tracked repo, set origin to `git@github.com:example/ssh-repo.git`.
2. Record and run `wrk --projects --github`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureGitHubFilterHelpersUsed()
	origin := setupBareOrigin(t, req.WorkRoot, "origin-ssh")
	repo := setupTrackedMainRepo(t, req.WorkRoot, "ssh-repo", origin, "ssh github project")
	setOriginURL(t, repo, "git@github.com:example/ssh-repo.git")
	recordProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```
