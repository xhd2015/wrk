# Scenario

**Feature**: linked worktree fixture with modules diverged from main (prove scan root)

```
# mainrepo/ (branch main) — planned when useMain=true
mainrepo/
  go.mod example.com/cli-install-main-root + cmd/mainbin
  tools/go.mod example.com/cli-install-main-tools + cmd/toolbin
linked-wt/ (branch side, diverged) — planned when useMain=false
  go.mod example.com/cli-install-wt-root + cmd/wtbin
  (no tools/, no mainbin)
process cwd = linked-wt
```

## Preconditions

- Git available; main and linked worktree **diverge** so the dry-run plan differs
  by scan root (main → `mainbin` + `toolbin`, K=2; linked WT → `wtbin`, K=1).
- GOBIN stays **empty**: `--install` needs no stub to plan or install a name.

## Steps

1. Init `mainrepo` with root + `tools` modules; commit on `main`.
2. `git worktree add -b side linked-wt`.
3. Rewrite the linked worktree to a single different module (`wtbin`); commit.
4. Default process cwd (`ModuleRoot`) to the linked worktree.
5. Leaves set Args for the `--main` compose and the without-main contrast.

## Context

- Group default Args is `--main --install mainbin toolbin --dry-run` (I8);
  the contrast leaf drops `--main` and requests `wtbin` (I9).
- RelDir in multi headers is relative to the **main** scan root (`.` and `tools`).

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

	writeGoMod(t, mainRepo, "example.com/cli-install-main-root")
	writePackageMain(t, filepath.Join(mainRepo, "cmd", "mainbin"))

	toolsMod := filepath.Join(mainRepo, "tools")
	writeGoMod(t, toolsMod, "example.com/cli-install-main-tools")
	writePackageMain(t, filepath.Join(toolsMod, "cmd", "toolbin"))

	gitCommitAll(t, mainRepo, "init main multi-module for install main-compose")

	linkedWT := filepath.Join(req.WorkRoot, "linked-wt")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "side", linkedWT)

	// Diverge the linked worktree so useMain=false plans different modules.
	if err := os.RemoveAll(filepath.Join(linkedWT, "tools")); err != nil {
		t.Fatalf("remove linked tools: %v", err)
	}
	if err := os.RemoveAll(filepath.Join(linkedWT, "cmd")); err != nil {
		t.Fatalf("remove linked cmd: %v", err)
	}
	writeGoMod(t, linkedWT, "example.com/cli-install-wt-root")
	writePackageMain(t, filepath.Join(linkedWT, "cmd", "wtbin"))
	gitCommitAll(t, linkedWT, "diverge worktree to wt-only module")

	// Process cwd is the linked worktree (not main); L2 in-process capture.
	req.ModuleRoot = resolvePath(t, linkedWT)
	req.InProcess = true
	// Default compose Args (I8); leaves may override.
	req.Args = []string{"--main", "--install", "mainbin", "toolbin", "--dry-run"}
	return nil
}
```
