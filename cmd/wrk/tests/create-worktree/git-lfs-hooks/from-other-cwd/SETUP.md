# Scenario

**Feature**: wrk <repoDir> from a different cwd fails with stripped PATH (dot-pkgs -> agent-pro pattern)

```
fakehome/.local/bin/git-lfs
agent-pro repo elsewhere; cwd=dot-pkgs; PATH=/usr/bin:/bin
wrk <agent-pro> -> exit 1 (expected)
```

## Steps

1. Install fake `git-lfs` under `{WorkRoot}/fakehome/.local/bin/`.
2. Create LFS-hook repo at `{WorkRoot}/agent-pro`.
3. Set cwd to `{WorkRoot}/dot-pkgs` (not inside the source repo).
4. Run `wrk {WorkRoot}/agent-pro` with stripped PATH.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.FakeHome = initFakeHomeWithGitLFS(t, req.WorkRoot)
	req.UseMinimalPath = true
	repoDir := filepath.Join(req.WorkRoot, "agent-pro")
	hooksDir := filepath.Join(req.WorkRoot, "hooks")
	initGitRepoWithLFSHooks(t, repoDir, hooksDir)
	req.MainRepo = repoDir
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "dot-pkgs")
	req.TargetDir = repoDir
	return nil
}
```