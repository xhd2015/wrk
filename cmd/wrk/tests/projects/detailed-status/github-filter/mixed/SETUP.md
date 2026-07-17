# Scenario

**Feature**: mixed registry — only github.com origin prints

```
record GH + local bare origin + no-remote -> wrk --projects --github -> only GH block
```

## Steps

1. Create three repos: `aaa-gh` with github.com origin URL, `mmm-local` with bare origin, `zzz-noremote` with no remote.
2. Record all three via `wrk --add`.
3. Run `wrk --projects --github`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureGitHubFilterHelpersUsed()

	// GitHub: tracked bare then rewrite origin URL to github.com (local tracking refs remain).
	originGH := setupBareOrigin(t, req.WorkRoot, "origin-gh")
	repoGH := setupTrackedMainRepo(t, req.WorkRoot, "aaa-gh", originGH, "github project")
	setOriginURL(t, repoGH, "https://github.com/example/aaa-gh.git")

	originLocal := setupBareOrigin(t, req.WorkRoot, "origin-local")
	repoLocal := setupTrackedMainRepo(t, req.WorkRoot, "mmm-local", originLocal, "local project")

	repoNoRemote := filepath.Join(req.WorkRoot, "zzz-noremote")
	initDetailedStatusRepo(t, repoNoRemote, "no remote project")

	recordProject(t, req, repoGH)
	recordProject(t, req, repoLocal)
	recordProject(t, req, repoNoRemote)

	req.MainRepo = repoGH
	req.SecondRepo = repoLocal
	// Reuse DepPath for third path (no ThirdRepo on Request).
	req.DepPath = repoNoRemote
	return nil
}
```
