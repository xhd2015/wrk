# Scenario

**Feature**: dry-run from git repo subdirectory still plans all modules (C4)

```
# C4: git repo with root + tools modules; cwd = pkg/sub (no go.mod)
repo/ (git)
  go.mod example.com/cli-scan-root + cmd/rootbin
  tools/go.mod example.com/cli-scan-tools + cmd/toolbin
  pkg/sub/   <- process cwd
  GOBIN/{rootbin,toolbin}
  -> wrk --reinstall-local --dry-run
  -> same multi plan as from repo root (both modules; across 2)
```

## Steps

1. Init git repo at `{WorkRoot}/repo` on branch `main`.
2. Write root module `example.com/cli-scan-root` with `./cmd/rootbin`.
3. Write nested `tools/` module `example.com/cli-scan-tools` with `./cmd/toolbin`.
4. Create empty `{repo}/pkg/sub` (no go.mod).
5. Commit tree so ShowToplevel is valid.
6. Touch `$GOBIN/rootbin` and `$GOBIN/toolbin`.
7. Set process cwd (`ModuleRoot`) to `pkg/sub`.
8. Expect multi dry-run covering both modules (scan root = git toplevel).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "repo")
	initGitRepoOnMain(t, repo)

	writeGoMod(t, repo, "example.com/cli-scan-root")
	writePackageMain(t, filepath.Join(repo, "cmd", "rootbin"))

	toolsMod := filepath.Join(repo, "tools")
	writeGoMod(t, toolsMod, "example.com/cli-scan-tools")
	writePackageMain(t, filepath.Join(toolsMod, "cmd", "toolbin"))

	sub := filepath.Join(repo, "pkg", "sub")
	mkdirAll(t, sub)

	gitCommitAll(t, repo, "init multi-module checkout for cli dry-run")

	repo = resolvePath(t, repo)
	sub = resolvePath(t, sub)

	touchBin(t, req.BinDir, "rootbin")
	touchBin(t, req.BinDir, "toolbin")

	// Process cwd is a subdir without go.mod; multi plan still uses git toplevel.
	req.ModuleRoot = sub
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
