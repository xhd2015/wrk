# Scenario

**Feature**: `wrk --done` cascade preflight (D1–D4, D7) — hard nested-main error, all-or-nothing dirty/not-included gates, stricter cascade confirm

```
# D1 nested main under consumer → hard Error:, abort, no mutations
# D2 dirty cascade target or dirty own → preflight Error:, nothing removed
# D3/D4 cascade HEAD not contained in its main → must confirm; default auto-yes
#     does NOT skip; -y still auto-yes; one prompt per not-included target
# D7 preflight failures also apply under --dry-run (covered under done-output/)
```

## Preconditions

- Classic TDD: behaviors not fully implemented — leaves expect **RED** until implementer lands.
- Git + Go available for consumer/`--bring` fixtures.
- Prefer this grouping over rewriting sealed siblings that still document current
  warn+skip nested-main or cascade default auto-yes.

## Steps

- Grouping only: shared helpers build fixtures; leaves set `req.Args` / stdin.

```go
import (
	"os/exec"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

const cascadePreflightDepModule = "example.com/dep-preflight"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}

// setupCascadePreflightCleanContained builds:
//   consumer linked wt + clean contained external/* (no ahead commits on dep)
//   + dropreplace + clean gitignore commit so own --done can succeed after cascade.
// Sets MainRepo, WtDir, WtBranch, DepPath, ExternalWtDir, RepoDir.
func setupCascadePreflightCleanContained(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGoModInDir(t, mainRepo, "edit", "-require="+cascadePreflightDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+cascadePreflightDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--bring", depRepo)
	externalPath = compositionResolvePath(t, externalPath)
	req.ExternalWtDir = externalPath

	// Drop replace so local-replace guard does not block after cascade.
	runGoModInDir(t, wtDir, "edit", "-dropreplace="+cascadePreflightDepModule)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", ".gitignore", "go.mod")
	runGitIsolated(t, wtDir, "commit", "-m", "drop replace; ignore external")

	req.RepoDir = wtDir
}

// setupCascadePreflightWithNestedMain: clean contained cascade + nested full main
// under consumer tree (RepoTypeMain). Nested path stored in SecondRepo.
func setupCascadePreflightWithNestedMain(t *testing.T, req *Request) {
	t.Helper()
	setupCascadePreflightCleanContained(t, req)

	nestedMain := filepath.Join(req.WtDir, "vendor", "nested-main")
	initGitRepoOnMain(t, nestedMain)
	writeFile(t, filepath.Join(nestedMain, "README"), "nested main clone\n")
	runGitIsolated(t, nestedMain, "add", "README")
	runGitIsolated(t, nestedMain, "commit", "-m", "nested main content")
	req.SecondRepo = compositionResolvePath(t, nestedMain)

	// Keep consumer porcelain clean despite nested files.
	writeFile(t, filepath.Join(req.WtDir, ".gitignore"), "/external\n/vendor\n")
	runGitIsolated(t, req.WtDir, "add", ".gitignore")
	runGitIsolated(t, req.WtDir, "commit", "-m", "ignore vendor nested main")
}

// setupCascadePreflightAheadExternal: ahead external (not contained in dep main)
// + drop replace so consumer can finish if cascade is confirmed/-y.
func setupCascadePreflightAheadExternal(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	// Reuse root helper for ahead external, then clear replace + clean porcelain.
	setupConsumerWithAheadExternalDep(t, req)
	req.WtDir = compositionResolvePath(t, req.WtDir)
	req.ExternalWtDir = compositionResolvePath(t, req.ExternalWtDir)
	req.DepPath = compositionResolvePath(t, req.DepPath)
	req.MainRepo = compositionResolvePath(t, req.MainRepo)

	runGoModInDir(t, req.WtDir, "edit", "-dropreplace="+cascadeAheadExternalDepModule)
	writeFile(t, filepath.Join(req.WtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, req.WtDir, "add", "-A")
	runGitIsolated(t, req.WtDir, "commit", "-m", "drop replace; ignore external", "--allow-empty")

	req.RepoDir = req.WtDir
}

// assertCascadePreflightNoRemovals: external + consumer still on disk (preflight abort).
func assertCascadePreflightNoRemovals(t *testing.T, req *Request) {
	t.Helper()
	if req.ExternalWtDir != "" {
		assertFileExists(t, req.ExternalWtDir)
		assertGitFileIsWorktreeLink(t, req.ExternalWtDir)
	}
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
}

var (
	_ = setupCascadePreflightCleanContained
	_ = setupCascadePreflightWithNestedMain
	_ = setupCascadePreflightAheadExternal
	_ = assertCascadePreflightNoRemovals
)
```
