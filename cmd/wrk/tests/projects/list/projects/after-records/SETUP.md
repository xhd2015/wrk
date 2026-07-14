# Scenario

**Feature**: wrk --projects after multiple auto-records

```
wrk --list (repoA) + wrk --list (repoB) -> wrk --projects -> sorted detailed blocks
```

## Steps

1. Create two git repos `aaa` and `zzz` under `{WorkRoot}`, each with its own bare `origin` remote.
2. Run `wrk --list` from each to auto-record.
3. Run `wrk --projects`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureProjectListHelpersUsed()
	originA := setupBareOriginForList(t, req.WorkRoot, "origin-aaa")
	originZ := setupBareOriginForList(t, req.WorkRoot, "origin-zzz")

	repoA := initProjectsRepo(t, req.WorkRoot, "aaa")
	repoZ := initProjectsRepo(t, req.WorkRoot, "zzz")
	runGitIsolated(t, repoA, "remote", "add", "origin", originA)
	runGitIsolated(t, repoA, "push", "-u", "origin", "main")
	runGitIsolated(t, repoZ, "remote", "add", "origin", originZ)
	runGitIsolated(t, repoZ, "push", "-u", "origin", "main")

	runWrkWithArgs(t, req, repoA, "--list")
	runWrkWithArgs(t, req, repoZ, "--list")
	req.SecondRepo = repoZ
	req.MainRepo = repoA
	return nil
}
```