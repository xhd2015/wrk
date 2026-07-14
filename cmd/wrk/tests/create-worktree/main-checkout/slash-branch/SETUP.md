# Scenario

**Feature**: slash in branch name sanitized for both path token and branch name (P1 behavior-change)

```
# branch feature/foo → token and branch segment feature-foo (no slash in git branch)
myrepo (feature/foo) -> wrk -> path myrepo-feature-foo-2026-06-30, branch feature-foo-2026-06-30
```

## Steps

1. Initialize git repo `myrepo` on branch `main`.
2. Create and check out branch `feature/foo`.
3. Run `wrk` from `myrepo`.

```go
func Setup(t *testing.T, req *Request) error {
	initGitRepoOnMain(t, req.RepoDir)
	runGitIsolated(t, req.RepoDir, "checkout", "-b", "feature/foo")
	return nil
}
```
