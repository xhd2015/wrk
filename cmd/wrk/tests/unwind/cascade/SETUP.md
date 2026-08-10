# Scenario

**Feature**: global free-module cascade under `--unwind` + `--tag-next`

```
# stack module DAG (same inventory as show-graph) + free-first Kahn
# dry-run (P1): after peel lines; would: tag-next / would: pin; zero mutations
# apply (P2): land prelude then tag one scope → pin keep-replace → selective commit
# apply (P3): partial edit when dirty go.mod without --add-all
# apply (P4): nested skip pin + --reinstall-local tail (no tidy/unknown revision)
stack modules (pending) -> PlanUnwindCascade
  -> dry-run: would: tag-next / would: pin [+ would: reinstall local binaries]
  -> apply: tags + pin commits + push (no TagNextAll-on-peel)
  -> apply + reinstall-local: pin nested skip consumers then soft reinstall
  (testdata / forever-skip scopes never get tag-next)
```

## Preconditions

- Inherits root `cmd/wrk/tests/unwind` Request/Response/Run and fixture helpers.
- **P1 dry-run** leaves under `cascade/dry-run/` are **sealed** (6 leaves) — do
  not rewrite their ASSERT meaning. Plan printer is GREEN.
- **P2 apply** clean / `--add-all` leaves under `cascade/apply/clean/` and
  `dirty-gomod/with-add-all/` are **sealed GREEN**.
- **P3 partial edit** leaves under `cascade/apply/partial-edit/` and C-AP5
  (`dirty-gomod/without-add-all`) — do not break dry-run or clean apply leaves.
- **P4 reinstall** leaves under `cascade/apply/reinstall-local/` — nested skip
  consumer pin + reinstall-local regression (and multi-repo tail). Do not break
  the sealed 13 cascade leaves above.
- Leaves set `req.InProcess = true` and full `req.Args` including `--unwind`.
- Dry-run cascade vocabulary is **top-level** (no indent), distinct from under-peel
  `  would: pin stack consumers` / `  would: create release tag`.

## Steps

1. Grouping scopes the cascade family; descendants branch on dry-run vs apply.
2. Leaves seed multi-module / multi-repo fixtures and assert plan lines or
   apply side effects (P4 also asserts reinstall-local tail).

## Context

- Module identity = Go **module path** (e.g. `example.com/dot-pkgs`).
- Free-first on module edges: **From depends on To** → free deps before consumers.
- Pin dry-run still shown even when apply would keep local replace.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	// Intra-repo shared leaf (C-DR1 / C-DR3 / C-DR4).
	cascadeSharedModule   = "example.com/root/shared"
	cascadeSharedDir      = "pkgs/shared" // rel under main
	cascadeSharedOldTag   = "pkgs/shared/v0.0.1"
	cascadeSharedNextTag  = "pkgs/shared/v0.0.2"
	// Testdata path-scope tags must never appear as cascade tag-next (C-DR4).
	cascadeTestdataModule  = "example.com/root/testdata-x"
	cascadeTestdataRelDir  = "testdata/x"
	cascadeTestdataOldTag  = "testdata/x/v0.0.1"
	cascadeTestdataNextTag = "testdata/x/v0.0.2"
)

// cascadeTagNextLine is the locked dry-run vocabulary for a module tag step.
func cascadeTagNextLine(modulePath, nextTag string) string {
	return "would: tag-next " + modulePath + " @ " + nextTag
}

// cascadePinLine is the locked dry-run vocabulary for a module pin step.
func cascadePinLine(consumerMod, depMod, ver string) string {
	return "would: pin " + consumerMod + " <- " + depMod + " @ " + ver
}

// hasCascadeTagNext reports whether stdout has a cascade tag-next line for modulePath
// (prefix match so next-tag may vary slightly while module identity is locked).
func hasCascadeTagNext(stdout, modulePath string) bool {
	return strings.Contains(stdout, "would: tag-next "+modulePath+" @")
}

// hasCascadePin reports a top-level cascade pin (consumer <- dep), not under-peel
// "would: pin stack consumers".
func hasCascadePin(stdout, consumerMod, depMod string) bool {
	needle := "would: pin " + consumerMod + " <- " + depMod
	return strings.Contains(stdout, needle)
}

// assertNoCascadeModuleLines fails if dry-run stdout contains cascade tag-next
// or cascade pin (… <- …) lines.
func assertNoCascadeModuleLines(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, "would: tag-next ") {
		t.Fatalf("cascade tag-next lines must be absent\nstdout:\n%s", stdout)
	}
	// Cascade pin form uses " <- "; under-peel "  would: pin stack consumers" does not.
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "would: pin ") && strings.Contains(trim, " <- ") {
			t.Fatalf("cascade pin line must be absent; found %q\nstdout:\n%s", trim, stdout)
		}
	}
}

// assertNoSuccessfulCascadeBody fails if stdout looks like a completed multi-step
// free-module cascade plan (used on cycle / hard reject).
func assertNoSuccessfulCascadeBody(t *testing.T, stdout string) {
	t.Helper()
	nTag := strings.Count(stdout, "would: tag-next ")
	if nTag >= 2 {
		t.Fatalf("cycle/error path must not print multi-step cascade tag plan; stdout:\n%s", stdout)
	}
	// Even a single tag-next after a cycle reject is a successful cascade body.
	if nTag >= 1 && strings.Contains(strings.ToLower(stdout), "cycle") {
		// Allow diagnostics that mention tags; fail only if it looks like plan vocabulary.
		if strings.Contains(stdout, "would: tag-next ") {
			t.Fatalf("cycle path must not emit would: tag-next cascade lines\nstdout:\n%s", stdout)
		}
	}
}

// assertCascadeAfterPeels requires every cascade tag/pin line (if any) to appear
// after the last peel line when peels are non-empty.
func assertCascadeAfterPeels(t *testing.T, stdout string, peelDisplays []string) {
	t.Helper()
	if len(peelDisplays) == 0 {
		return
	}
	lastPeel := -1
	for _, d := range peelDisplays {
		idx := indexPeelLine(stdout, d)
		if idx > lastPeel {
			lastPeel = idx
		}
	}
	if lastPeel < 0 {
		return
	}
	// Find first top-level cascade line (no indent).
	firstCascade := -1
	for _, raw := range strings.Split(stdout, "\n") {
		if strings.HasPrefix(raw, "would: tag-next ") ||
			(strings.HasPrefix(raw, "would: pin ") && strings.Contains(raw, " <- ")) {
			idx := strings.Index(stdout, raw)
			if idx >= 0 && (firstCascade < 0 || idx < firstCascade) {
				firstCascade = idx
			}
		}
	}
	if firstCascade >= 0 && firstCascade < lastPeel {
		t.Fatalf("cascade lines must appear after peel lines\nstdout:\n%s", stdout)
	}
}

// setupCascadeSingleRepoTwoModules builds one dirty main with root + pkgs/shared.
// shared is owned-changed after tag pkgs/shared/v0.0.1; root requires shared@v0.0.1
// with local replace. Root baseline tag v0.0.1. PeelOrder = ["."].
// Pending free-first cascade: tag shared next, then pin root <- shared.
func setupCascadeSingleRepoTwoModules(t *testing.T, req *Request) {
	t.Helper()

	mainRepo := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, mainRepo)

	sharedDir := filepath.Join(mainRepo, filepath.FromSlash(cascadeSharedDir))
	mkdirAll(t, sharedDir)
	writeGoModRequire(t, sharedDir, cascadeSharedModule)
	writeFile(t, filepath.Join(sharedDir, "shared.go"),
		"package shared\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	writeGoModRequire(t, mainRepo, unwindRootModule, cascadeSharedModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(mainRepo, "root.go"),
		"package root\n\nimport _ \""+cascadeSharedModule+"\"\n")
	appendLocalReplace(t, mainRepo, cascadeSharedModule, "./"+cascadeSharedDir)

	runGitIsolated(t, mainRepo, "add", "go.mod", "root.go", "pkgs")
	runGitIsolated(t, mainRepo, "commit", "-m", "root + shared modules")
	createLightweightTag(t, mainRepo, unwindApplyOldTag, "HEAD")
	createLightweightTag(t, mainRepo, cascadeSharedOldTag, "HEAD")

	// Owned change on shared leaf only → tagscope plans pkgs/shared/v0.0.2.
	writeFile(t, filepath.Join(sharedDir, "shared.go"),
		"package shared\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, mainRepo, "add", "pkgs")
	runGitIsolated(t, mainRepo, "commit", "-m", "shared owned change for next tag")

	mainRepo = resolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.LeafModulePath = cascadeSharedModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag
	markDirty(t, mainRepo)
	req.PeelOrder = []string{"."}
}

// setupCascadeSingleRepoTwoModulesWithTestdata is like single-repo two modules
// plus a testdata/x nested tree with go.mod + path-scope tags (must not get
// would: tag-next). Real free module remains shared (positive cascade RED).
func setupCascadeSingleRepoTwoModulesWithTestdata(t *testing.T, req *Request) {
	t.Helper()
	setupCascadeSingleRepoTwoModules(t, req)

	td := filepath.Join(req.MainRepo, filepath.FromSlash(cascadeTestdataRelDir))
	mkdirAll(t, td)
	writeGoModRequire(t, td, cascadeTestdataModule)
	writeFile(t, filepath.Join(td, "x.go"),
		"package x\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	runGitIsolated(t, req.MainRepo, "add", "testdata")
	runGitIsolated(t, req.MainRepo, "commit", "-m", "add testdata module tree")
	createLightweightTag(t, req.MainRepo, cascadeTestdataOldTag, "HEAD")

	// Owned change under testdata scope (if tagscope plans next, cascade must skip).
	writeFile(t, filepath.Join(td, "x.go"),
		"package x\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, req.MainRepo, "add", "testdata")
	runGitIsolated(t, req.MainRepo, "commit", "-m", "testdata owned change")
	// Keep DIRTY for peel pending.
	markDirty(t, req.MainRepo)
}

// setupCascadeMultiRepoBothDirty: root main + leaf external; both dirty.
// Leaf has owned change → next tag v0.0.2; root requires leaf@v0.0.1.
// Free-first peels: external leaf then . ; module cascade: leaf module then pin root.
func setupCascadeMultiRepoBothDirty(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	leafMain := filepath.Join(req.WorkRoot, labelDotPkgs)
	initGitRepoOnMain(t, leafMain)
	writeGoModRequire(t, leafMain, unwindDotPkgsModule)
	writeFile(t, filepath.Join(leafMain, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	runGitIsolated(t, leafMain, "add", "go.mod", "pkg.go")
	runGitIsolated(t, leafMain, "commit", "-m", "add dot-pkgs module")
	createLightweightTag(t, leafMain, unwindApplyOldTag, "")
	leafMain = resolvePath(t, leafMain)
	req.SecondRepo = leafMain

	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root consumer")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	extDir := filepath.Join(rootMain, "external")
	mkdirAll(t, extDir)
	leafExt := filepath.Join(extDir, labelDotPkgs+"-"+branchNameMainDate())
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate(), leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt
	req.WtBranch = branchNameMainDate()

	// Owned change on leaf WT for tag-next v0.0.2.
	writeFile(t, filepath.Join(leafExt, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, leafExt, "add", "pkg.go")
	runGitIsolated(t, leafExt, "commit", "-m", "leaf feature for next tag")
	markDirty(t, leafExt)

	writeFile(t, filepath.Join(rootMain, ".gitignore"), "/external\n")
	runGitIsolated(t, rootMain, "add", ".gitignore")
	runGitIsolated(t, rootMain, "commit", "-m", "ignore external stack members")
	// Both dirty: root also pending peel.
	markDirty(t, rootMain)

	req.RepoDir = rootMain
	setPeelOrderDisplays(t, req, leafExt, rootMain)
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	// Keep cascade helpers referenced for the generator.
	_ = cascadeTagNextLine
	_ = cascadePinLine
	_ = hasCascadeTagNext
	_ = hasCascadePin
	_ = assertNoCascadeModuleLines
	_ = assertNoSuccessfulCascadeBody
	_ = assertCascadeAfterPeels
	_ = setupCascadeSingleRepoTwoModules
	_ = setupCascadeSingleRepoTwoModulesWithTestdata
	_ = setupCascadeMultiRepoBothDirty
	return nil
}
```
