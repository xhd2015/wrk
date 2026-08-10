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
- **P1 dry-run** sealed leaves under `cascade/dry-run/` (C-DR1–C-DR6) — do **not**
  rewrite their ASSERT meaning. Plan printer is GREEN for those.
- **P3 G1** (`dry-run/with-tag-next/false-freehost-b1-interleave`): intended B1
  interleaved dry-run (early free / cascade / deferred consumer) for T-spl-like
  graphs — may RED until `FormatUnwindDryRun` interleaves peels with cascade.
- **P1 extension (this cycle):** external stack replace ⇒ needs-pin even when the
  free dep is clean and require already matches (no drift / no tag-next). New
  leaves under `cascade/dry-run/with-tag-next/replace-only-*` — Classic **RED**
  until planner treats droppable external replaces as pin triggers.
- **P2 apply** clean / `--add-all` leaves under `cascade/apply/clean/` and
  `dirty-gomod/with-add-all/` are **sealed GREEN**.
- **P3 partial edit** leaves under `cascade/apply/partial-edit/` and C-AP5
  (`dirty-gomod/without-add-all`) — do not break dry-run or clean apply leaves.
- **P4 reinstall** leaves under `cascade/apply/reinstall-local/` — nested skip
  consumer pin + reinstall-local regression (and multi-repo tail). Do not break
  the sealed cascade leaves above.
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

// setupCascadeReplaceOnlyExternalCleanDep builds the T3-core replace-only pin fixture:
//
//	leaf (dot-pkgs) nested external under root: **clean**, HEAD at release tag v0.0.1
//	root main: dirty; require leaf@v0.0.1 (matches tag — no require-drift);
//	  replace => ./external/dot-pkgs-main-<date> (droppable external stack replace)
//
// No owned-changed on leaf → no tag-next. Planner must still emit cascade pin at
// current require version (v0.0.1) solely because of the external replace (D1/D3).
// PeelOrder = ["."] only (clean free dep never peels).
func setupCascadeReplaceOnlyExternalCleanDep(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	// D3: no tag/drift → pin keeps current require version.
	req.ExpectedPinVersion = unwindApplyOldTag

	// --- leaf main: baseline tag only (no post-tag commits) ---
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

	// --- root consumer: require matches tag ---
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

	// Nest clean leaf WT under root/external (stack member + replace target).
	extDir := filepath.Join(rootMain, "external")
	mkdirAll(t, extDir)
	leafExt := filepath.Join(extDir, labelDotPkgs+"-"+branchNameMainDate())
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate(), leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt
	req.WtBranch = branchNameMainDate()
	// Intentionally leave leaf clean (no markDirty; HEAD still at v0.0.1 tag tree).

	writeFile(t, filepath.Join(rootMain, ".gitignore"), "/external\n")
	// External stack replace: droppable at apply; planner must treat as needs-pin.
	appendLocalReplace(t, rootMain, unwindDotPkgsModule, relLocalReplace(t, rootMain, leafExt))
	runGitIsolated(t, rootMain, "add", ".gitignore", "go.mod")
	runGitIsolated(t, rootMain, "commit", "-m", "ignore external + replace to clean leaf")

	// Consumer dirty only → peel `.`; free dep stays out of PeelOrder.
	markDirty(t, rootMain)
	req.RepoDir = rootMain
	req.PeelOrder = []string{"."}
}

// setupCascadeFalseFreeHostIntraPins is G1 / T-spl-like dry-run fixture:
//
//	external free (dot-pkgs): dirty owned-changed → tag-next v0.0.2
//	consumer monorepo (primary main):
//	  - root + pkgs/shared (no owned-change on shared after pkgs/shared/v0.0.1)
//	  - root requires shared @ v0.0.0 (drifts vs LatestTag → noise pin @ LatestTag
//	    without tag-next on shared — false freeHost trigger on apply)
//	  - root requires free @ v0.0.1 + droppable replace => ./external/…
//	  - consumer dirty (DIRTY only)
//
// Dry-run must reflect B1 apply order for this graph (G1): early peel free →
// cascade tag free + pin consumer → deferred peel consumer (not peels-then-cascade
// only). No land required when free is nested external under primary main without
// a linked free worktree origin (free is still Linked if worktree-added).
func setupCascadeFalseFreeHostIntraPins(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- free leaf main (owned-changed on linked external) ---
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

	// --- monorepo consumer: root + shared noise + free require ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)

	sharedDir := filepath.Join(rootMain, filepath.FromSlash(cascadeSharedDir))
	mkdirAll(t, sharedDir)
	writeGoModRequire(t, sharedDir, cascadeSharedModule)
	writeFile(t, filepath.Join(sharedDir, "shared.go"),
		"package shared\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	// Shared require drifts from LatestTag (pkgs/shared/v0.0.1 → v0.0.1) so
	// cascade plans noise pin root←shared without tag-next on shared.
	writeGoModRequire(t, rootMain, unwindRootModule,
		cascadeSharedModule+"@v0.0.0",
		unwindDotPkgsModule+"@"+unwindApplyOldTag,
	)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport (\n\t_ \""+cascadeSharedModule+"\"\n\t_ \""+unwindDotPkgsModule+"\"\n)\n")
	appendLocalReplace(t, rootMain, cascadeSharedModule, "./"+cascadeSharedDir)

	runGitIsolated(t, rootMain, "add", "go.mod", "root.go", "pkgs")
	runGitIsolated(t, rootMain, "commit", "-m", "root + shared + require free (noise shared drift)")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "HEAD")
	createLightweightTag(t, rootMain, cascadeSharedOldTag, "HEAD")
	// No post-tag owned change on shared → no cascade tag-next for shared.

	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	// Dirty free leaf under external (owned-changed for next tag).
	extDir := filepath.Join(rootMain, "external")
	mkdirAll(t, extDir)
	leafExtName := labelDotPkgs + "-" + branchNameMainDate()
	leafExt := filepath.Join(extDir, leafExtName)
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate(), leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt
	req.WtBranch = branchNameMainDate()

	writeFile(t, filepath.Join(leafExt, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, leafExt, "add", "pkg.go")
	runGitIsolated(t, leafExt, "commit", "-m", "leaf feature for next tag")
	markDirty(t, leafExt)

	// Droppable external replace; consumer dirty.
	relReplace := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, rootMain, unwindDotPkgsModule, "./"+relReplace)
	writeFile(t, filepath.Join(rootMain, ".gitignore"), "/external\n")
	runGitIsolated(t, rootMain, "add", "go.mod", ".gitignore")
	runGitIsolated(t, rootMain, "commit", "-m", "external stack replace + ignore")
	markDirty(t, rootMain)

	req.RepoDir = rootMain
	// Free-first peel among dirty: free external then consumer primary.
	setPeelOrderDisplays(t, req, leafExt, rootMain)
}

// setupCascadeReplaceOnlyIntraCleanShared is the D4 control: dirty root with
// only an **intra-repo** replace to pkgs/shared, require matches latest tag,
// shared **not** owned-changed. Must **not** invent a cascade pin solely because
// a local replace exists (keep-local / not droppable).
func setupCascadeReplaceOnlyIntraCleanShared(t *testing.T, req *Request) {
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
	runGitIsolated(t, mainRepo, "commit", "-m", "root + shared modules (no owned change)")
	createLightweightTag(t, mainRepo, unwindApplyOldTag, "HEAD")
	createLightweightTag(t, mainRepo, cascadeSharedOldTag, "HEAD")

	// No post-tag owned change on shared → no tag-next / no require-drift.
	mainRepo = resolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.LeafModulePath = cascadeSharedModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = "" // must not pin
	markDirty(t, mainRepo)
	req.PeelOrder = []string{"."}
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
	_ = setupCascadeReplaceOnlyExternalCleanDep
	_ = setupCascadeReplaceOnlyIntraCleanShared
	_ = setupCascadeFalseFreeHostIntraPins
	return nil
}
```
