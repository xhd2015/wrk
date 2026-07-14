# Scenario

**Feature**: wrk create from LFS repo fails with stripped PATH even when git-lfs is in login-shell paths

```
fakehome/.local/bin/git-lfs
myrepo (LFS hooks) + PATH=/usr/bin:/bin -> wrk -> exit 1 (expected)
```

## Steps

1. Install fake `git-lfs` under `{WorkRoot}/fakehome/.local/bin/`.
2. Create git repo with LFS post-checkout hook at `{WorkRoot}/hooks/`.
3. Run `wrk` from repo root with `PATH=/usr/bin:/bin` and `HOME=fakehome`.

```go
func Setup(t *testing.T, req *Request) error {
	req.FakeHome = initFakeHomeWithGitLFS(t, req.WorkRoot)
	req.UseMinimalPath = true
	repoDir := filepath.Join(req.WorkRoot, "myrepo")
	hooksDir := filepath.Join(req.WorkRoot, "hooks")
	initGitRepoWithLFSHooks(t, repoDir, hooksDir)
	req.RepoDir = repoDir
	return nil
}
```