# Scenario

**Feature**: linked worktree fixture with modules diverged from main (prove scan root)

```
# mainrepo/ (branch main) — multi-module tree planned when useMain=true
mainrepo/
  go.mod example.com/cli-main-root + cmd/mainbin
  tools/go.mod example.com/cli-main-tools + cmd/toolbin
linked-wt/ (branch side, diverged) — planned when useMain=false
  go.mod example.com/cli-wt-root + cmd/wtbin
  (no tools/)
GOBIN/{mainbin,toolbin,wtbin} present
process cwd = linked-wt
```

## Preconditions

- Git available (same as other multi git leaves under this tree).
- Main and linked worktree **diverge** so dry-run stdout differs by scan root:
  - main → multi plan (`mainbin` + `toolbin`, `across 2 modules`)
  - linked WT → single-mod plan (`wtbin` only, K=1 format)

## Steps

1. Init `mainrepo` with root + `tools` modules; commit on `main`.
2. `git worktree add -b side linked-wt` (same content initially).
3. Rewrite linked worktree to a single different module (`wtbin`); commit on `side`.
4. Touch GOBIN stubs for `mainbin`, `toolbin`, and `wtbin`.
5. Default process cwd (`ModuleRoot`) to linked worktree.
6. Leaves set Args for compose / without-main / flag order.

## Context

- Group default Args is **compose** order: `--main --reinstall-local --dry-run`
  (MC1). Flag-order leaf overrides; without-main drops `--main`.
- RelDir in multi headers is relative to **main** scan root (`.` and `tools`).

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	mainRepo := filepath.Join(req.WorkRoot, "mainrepo")
	initGitRepoOnMain(t, mainRepo)

	writeGoMod(t, mainRepo, "example.com/cli-main-root")
	writePackageMain(t, filepath.Join(mainRepo, "cmd", "mainbin"))

	toolsMod := filepath.Join(mainRepo, "tools")
	writeGoMod(t, toolsMod, "example.com/cli-main-tools")
	writePackageMain(t, filepath.Join(toolsMod, "cmd", "toolbin"))

	gitCommitAll(t, mainRepo, "init main multi-module for main-compose")

	linkedWT := filepath.Join(req.WorkRoot, "linked-wt")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "side", linkedWT)

	// Diverge the linked worktree so useMain=false plans different modules.
	if err := os.RemoveAll(filepath.Join(linkedWT, "tools")); err != nil {
		t.Fatalf("remove linked tools: %v", err)
	}
	if err := os.RemoveAll(filepath.Join(linkedWT, "cmd")); err != nil {
		t.Fatalf("remove linked cmd: %v", err)
	}
	writeGoMod(t, linkedWT, "example.com/cli-wt-root")
	writePackageMain(t, filepath.Join(linkedWT, "cmd", "wtbin"))
	gitCommitAll(t, linkedWT, "diverge worktree to wt-only module")

	linkedWT = resolvePath(t, linkedWT)

	touchBin(t, req.BinDir, "mainbin")
	touchBin(t, req.BinDir, "toolbin")
	touchBin(t, req.BinDir, "wtbin")

	// Process cwd is the linked worktree (not main).
	req.ModuleRoot = linkedWT
	// Default compose Args (MC1); leaves may override.
	req.Args = []string{"--main", "--reinstall-local", "--dry-run"}
	return nil
}
```
