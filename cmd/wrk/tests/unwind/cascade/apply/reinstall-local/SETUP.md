# Scenario

**Feature**: apply free-module cascade with `--reinstall-local` tail (P4 ship gate)

```
# after land prelude + global cascade (tag one scope → pin keep-replace → commit)
# reinstall-local runs only on collected mains — must not fail with tidy /
# unknown revision when nested skip consumers were pinned
leaf owned-changed + nested consumer (skip tag) old require + local replace
  -> wrk --unwind --tag-next --push --reinstall-local
  -> cascade pin nested consumer require to new tag (replace kept)
  -> reinstall-local exit 0 / soft-skip; no unknown revision; no tidy failure
```

## Preconditions

- Inherits `cascade/apply/` helpers (`setupApplyCascade*`, pin asserts) and
  parent cascade constants (`cascadeSharedModule`, tags, line helpers).
- Leaves set `req.InProcess = true` and full `req.Args` **with**
  `--reinstall-local` and **without** `--dry-run`.
- **P4 Classic TDD:** nested-skip + reinstall regression is the primary leaf
  (expect **RED** if product still fails that path after P2/P3; **GREEN**
  backfill OK if cascade already fixed it — still design the leaf).
- **C-RI3:** `nested-cmd-requires-parent/` — agent-pro shape: nested **cmd/**
  requires **parent** root with `replace => ../` (not sibling tools→shared).
- Do **not** rewrite sealed P1 dry-run (13→ leave dry-run untouched) or P2/P3
  clean / dirty / partial-edit ASSERT meanings.
- Dry-run reinstall vocabulary remains sealed as **C-DR6**
  (`cascade/dry-run/with-tag-next/reinstall-local-tail/`). Apply polish
  (cascade lines before reinstall) is asserted on the nested-skip leaf.

## Steps

1. Grouping marks apply cascade + reinstall-local integration (P4).
2. Leaves seed monorepo nested-skip or multi-repo free-first fixtures, enable
   isolated `GOBIN`, and assert pin side effects + reinstall tail success.

## Context

- **Nested skip consumer:** a nested module that is **not** owned-changed (no
  cascade `tag-next` for it) but **requires** a pending free module with old
  require + local replace. Cascade must still **pin** it; reinstall must then
  resolve without `unknown revision` / go-mod-tidy hard failure.
- Reinstall soft: install item failures are soft (exit 0 + warning). Asserts
  prefer **require bump + keep-replace** (cascade correctness) and absence of
  hard tidy/revision diagnostics; when a `cmd/` + GOBIN stub is seeded, prefer
  `reinstalled N… failed 0`.
- Isolated `GOBIN={WorkRoot}/gobin` via `req.ExtraEnv` (Capture applyEnvPairs).

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	// Nested skip consumer (C-RI1): not owned-changed → no cascade tag-next;
	// requires shared@old with local replace → cascade pin only.
	cascadeToolsModule = "example.com/root/tools"
	cascadeToolsDir    = "tools"
	cascadeToolsBin    = "tool"

	// Nested cmd requires parent (C-RI3 / agent-pro shape):
	// free monorepo root + cmd/ requires parent @old with keep-local replace => ../
	// and a stale transitive require (extra) that go mod tidy must bump — like
	// agent-pro/cmd requiring an older go-pkgs while parent is newer.
	cascadeCmdHarnessModule = "example.com/dot-pkgs/cmd-harness"
	cascadeCmdDir           = "cmd"
	cascadeCmdBin           = "tool"
	cascadeExtraModule      = "example.com/extra"
	cascadeExtraOld         = "v0.0.1"
	cascadeExtraNext        = "v0.0.2"
)

// setupApplyCascadeNestedSkipConsumer builds a single-repo monorepo where:
//   - pkgs/shared is owned-changed → cascade tag-next pkgs/shared/v0.0.2
//   - tools/ is a nested module that is NOT owned-changed ("skip" for tag)
//     but requires shared@v0.0.1 with replace => ../pkgs/shared
//   - tools/cmd/tool package main imports shared (reinstall exercises go install)
//   - root go.mod does NOT require shared (consumer is nested tools only)
// Free-first cascade: tag shared → pin tools <- shared @ v0.0.2 (no tools tag).
// GOBIN isolated under WorkRoot/gobin with tool stub present.
func setupApplyCascadeNestedSkipConsumer(t *testing.T, req *Request) {
	t.Helper()

	mainRepo := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, mainRepo)

	// Free module: pkgs/shared @ baseline tag.
	sharedDir := filepath.Join(mainRepo, filepath.FromSlash(cascadeSharedDir))
	mkdirAll(t, sharedDir)
	writeGoModRequire(t, sharedDir, cascadeSharedModule)
	writeFile(t, filepath.Join(sharedDir, "shared.go"),
		"package shared\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	// Root module: no require of shared (pin target is nested tools only).
	writeGoModRequire(t, mainRepo, unwindRootModule)
	writeFile(t, filepath.Join(mainRepo, "root.go"), "package root\n")

	// Nested skip consumer: require + local replace; not owned-changed later.
	toolsDir := filepath.Join(mainRepo, cascadeToolsDir)
	mkdirAll(t, toolsDir)
	writeGoModRequire(t, toolsDir, cascadeToolsModule, cascadeSharedModule+"@"+unwindApplyOldTag)
	appendLocalReplace(t, toolsDir, cascadeSharedModule, "../"+cascadeSharedDir)
	writeFile(t, filepath.Join(toolsDir, "tools.go"),
		"package tools\n\nimport _ \""+cascadeSharedModule+"\"\n")
	toolMain := filepath.Join(toolsDir, "cmd", cascadeToolsBin)
	mkdirAll(t, toolMain)
	writeFile(t, filepath.Join(toolMain, "main.go"),
		"package main\n\nimport (\n\t\"fmt\"\n\t_ \""+cascadeSharedModule+"\"\n)\n\nfunc main() { fmt.Println(\"tool\") }\n")

	runGitIsolated(t, mainRepo, "add", "go.mod", "root.go", "pkgs", "tools")
	runGitIsolated(t, mainRepo, "commit", "-m", "root + shared + nested tools consumer")
	createLightweightTag(t, mainRepo, unwindApplyOldTag, "HEAD")
	createLightweightTag(t, mainRepo, cascadeSharedOldTag, "HEAD")

	// Owned change on shared only → tools stays skip for cascade tag-next.
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

	bare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, req.MainRepo, bare)
	if tagRefExists(t, req.MainRepo, unwindApplyOldTag) {
		runGitIsolated(t, req.MainRepo, "push", "origin", unwindApplyOldTag)
	}
	if tagRefExists(t, req.MainRepo, cascadeSharedOldTag) {
		runGitIsolated(t, req.MainRepo, "push", "origin", cascadeSharedOldTag)
	}
	req.OriginBare = bare

	// Isolated GOBIN so reinstall-local installs only into the leaf workspace.
	enableIsolatedReinstallGOBIN(t, req, cascadeToolsBin)
}

// setupApplyCascadeNestedCmdRequiresParent builds agent-pro production shape:
//
//	free monorepo (dot-pkgs) under consumer/external linked worktree:
//	  example.com/dot-pkgs              tagged @ v0.0.2; requires extra@v0.0.2
//	  example.com/dot-pkgs/cmd-harness  under cmd/ (skip tag)
//	    require parent@v0.0.1 + extra@v0.0.1   ← parent require drift + stale transitive
//	    replace parent => ../                 ← keep-local (agent-pro)
//	    package main at cmd/tool
//	consumer root (primary RepoDir):
//	  require free@v0.0.1 + droppable external replace => ./external/…
//	  dirty peel so unwind runs cascade pin on free linked Path
//
// Production: pin agent-pro/cmd ← agent-pro @ existing tag (replace kept) on the
// free linked WT; --reinstall-local scans free **main** (useMain). If pin never
// lands on main, nested cmd stays untidy → go install "updates to go.mod needed".
//
// Distinct from C-RI1 (tools → sibling pkgs/shared on same main).
func setupApplyCascadeNestedCmdRequiresParent(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- free monorepo main: parent + nested cmd (agent-pro shape) ---
	leafMain := filepath.Join(req.WorkRoot, labelDotPkgs)
	initGitRepoOnMain(t, leafMain)

	// Parent requires extra@next (like agent-pro root requiring go-pkgs@newer).
	writeGoModRequire(t, leafMain, unwindDotPkgsModule, cascadeExtraModule+"@"+cascadeExtraNext)
	writeFile(t, filepath.Join(leafMain, "pkg.go"),
		"package dotpkgs\n\nimport _ \""+cascadeExtraModule+"\"\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	cmdDir := filepath.Join(leafMain, cascadeCmdDir)
	mkdirAll(t, cmdDir)
	// Nested cmd: parent@old + extra@old + keep-local replace => ../ .
	// extra@old is the "updates to go.mod needed" surface when parent pulls extra@next.
	writeGoModRequire(t, cmdDir, cascadeCmdHarnessModule,
		unwindDotPkgsModule+"@"+unwindApplyOldTag,
		cascadeExtraModule+"@"+cascadeExtraOld)
	appendLocalReplace(t, cmdDir, unwindDotPkgsModule, "../")
	writeFile(t, filepath.Join(cmdDir, "harness.go"),
		"package harness\n\nimport (\n\t_ \""+unwindDotPkgsModule+"\"\n\t_ \""+cascadeExtraModule+"\"\n)\n")
	toolMain := filepath.Join(cmdDir, cascadeCmdBin)
	mkdirAll(t, toolMain)
	writeFile(t, filepath.Join(toolMain, "main.go"),
		"package main\n\nimport (\n\t\"fmt\"\n\t_ \""+unwindDotPkgsModule+"\"\n\t_ \""+cascadeExtraModule+"\"\n)\n\nfunc main() { fmt.Println(\"tool\") }\n")

	runGitIsolated(t, leafMain, "add", "go.mod", "pkg.go", cascadeCmdDir)
	runGitIsolated(t, leafMain, "commit", "-m", "free monorepo root + cmd requires parent")
	createLightweightTag(t, leafMain, unwindApplyOldTag, "HEAD")

	// Advance parent release; leave cmd require at old (pin target).
	writeFile(t, filepath.Join(leafMain, "pkg.go"),
		"package dotpkgs\n\nimport _ \""+cascadeExtraModule+"\"\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, leafMain, "add", "pkg.go")
	runGitIsolated(t, leafMain, "commit", "-m", "free parent release next (cmd require still old)")
	createLightweightTag(t, leafMain, unwindApplyNextTag, "HEAD")

	leafMain = resolvePath(t, leafMain)
	req.SecondRepo = leafMain

	leafBare := setupBareOrigin(t, req.WorkRoot, "leaf-origin")
	attachOriginAndPushMain(t, leafMain, leafBare)
	if tagRefExists(t, leafMain, unwindApplyOldTag) {
		runGitIsolated(t, leafMain, "push", "origin", unwindApplyOldTag)
	}
	if tagRefExists(t, leafMain, unwindApplyNextTag) {
		runGitIsolated(t, leafMain, "push", "origin", unwindApplyNextTag)
	}
	req.OriginBare = leafBare

	// --- consumer main (primary) ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "consumer requires free parent")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain
	req.RepoDir = rootMain

	rootBare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, rootMain, rootBare)
	if tagRefExists(t, rootMain, unwindApplyOldTag) {
		runGitIsolated(t, rootMain, "push", "origin", unwindApplyOldTag)
	}
	writeFile(t, filepath.Join(req.WorkRoot, "root-origin.path"), rootBare+"\n")

	// Linked free WT under consumer/external (pin lands here; reinstall scans free main).
	extDir := filepath.Join(rootMain, "external")
	mkdirAll(t, extDir)
	leafExtName := labelDotPkgs + "-" + branchNameMainDate()
	leafExt := filepath.Join(extDir, leafExtName)
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate(), leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt
	req.WtBranch = branchNameMainDate()

	// Droppable external replace → cascade pin path for free.
	relReplace := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, rootMain, unwindDotPkgsModule, "./"+relReplace)
	writeFile(t, filepath.Join(rootMain, ".gitignore"), "/external\n")
	runGitIsolated(t, rootMain, "add", "go.mod", ".gitignore")
	runGitIsolated(t, rootMain, "commit", "-m", "external free replace + ignore")

	// Dirty consumer so peel runs (production peel .); free stays clean for pin-only.
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n\nfunc Dirty() {}\n")
	runGitIsolated(t, rootMain, "add", "root.go")
	// leave uncommitted so peel sees dirt
	markDirty(t, rootMain)
	req.PeelOrder = []string{"."}

	// Offline proxy: extra old+next + free parent old+next (network pin of consumer).
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedExtraVersions(t, req, proxyRoot)
	seedFreeParentVersions(t, req, proxyRoot, leafMain)
	enableFileModuleProxy(t, req, proxyRoot)

	// Reinstall GOBIN stub for free nested cmd/tool (scanned from free main).
	enableIsolatedReinstallGOBIN(t, req, cascadeCmdBin)
}

// seedExtraVersions seeds example.com/extra @ old and next into file:// modproxy.
func seedExtraVersions(t *testing.T, req *Request, proxyRoot string) {
	t.Helper()
	for _, ver := range []string{cascadeExtraOld, cascadeExtraNext} {
		seed := filepath.Join(req.WorkRoot, "seed-extra-"+ver)
		mkdirAll(t, seed)
		writeGoModRequire(t, seed, cascadeExtraModule)
		writeFile(t, filepath.Join(seed, "extra.go"),
			"package extra\n\nfunc Version() string { return \""+ver+"\" }\n")
		seedFileModuleProxy(t, proxyRoot, cascadeExtraModule, ver, seed)
	}
}

// seedFreeParentVersions seeds free parent module old+next for consumer network pin.
func seedFreeParentVersions(t *testing.T, req *Request, proxyRoot, leafMain string) {
	t.Helper()
	// Old: minimal parent @ old without needing network for extra during seed pack.
	oldSeed := filepath.Join(req.WorkRoot, "seed-free-"+unwindApplyOldTag)
	mkdirAll(t, oldSeed)
	writeGoModRequire(t, oldSeed, unwindDotPkgsModule, cascadeExtraModule+"@"+cascadeExtraNext)
	writeFile(t, filepath.Join(oldSeed, "pkg.go"),
		"package dotpkgs\n\nimport _ \""+cascadeExtraModule+"\"\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyOldTag, oldSeed)

	nextSeed := filepath.Join(req.WorkRoot, "seed-free-"+unwindApplyNextTag)
	mkdirAll(t, nextSeed)
	writeGoModRequire(t, nextSeed, unwindDotPkgsModule, cascadeExtraModule+"@"+cascadeExtraNext)
	writeFile(t, filepath.Join(nextSeed, "pkg.go"),
		"package dotpkgs\n\nimport _ \""+cascadeExtraModule+"\"\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyNextTag, nextSeed)
	_ = leafMain
}

// freeMainCmdHarnessGoModPath returns free **main** cmd/go.mod (reinstall scan
// root and cascade pin target when free linked Path is clean).
func freeMainCmdHarnessGoModPath(req *Request) string {
	return filepath.Join(req.SecondRepo, cascadeCmdDir, "go.mod")
}

// assertNestedCmdRequiresParentPinned checks free **main** cmd/go.mod (pin lands
// on MainRepo when linked Path is clean): require parent at ExpectedPinVersion
// and keep-local replace => ../ still present.
func assertNestedCmdRequiresParentPinned(t *testing.T, req *Request) {
	t.Helper()
	if req.LeafModulePath == "" || req.ExpectedPinVersion == "" || req.SecondRepo == "" {
		t.Fatal("LeafModulePath, ExpectedPinVersion, SecondRepo required")
	}
	cmdMod := freeMainCmdHarnessGoModPath(req)
	got := requireVersionInGoMod(t, cmdMod, req.LeafModulePath)
	if got != req.ExpectedPinVersion {
		t.Fatalf("nested cmd require parent %s = %q, want %s (cascade pin bump on free main)\n%s:\n%s",
			req.LeafModulePath, got, req.ExpectedPinVersion, cmdMod, readFile(t, cmdMod))
	}
	if !goModHasReplace(t, cmdMod, req.LeafModulePath) {
		t.Fatalf("nested cmd must KEEP local replace for parent %s:\n%s",
			req.LeafModulePath, readFile(t, cmdMod))
	}
	content := readFile(t, cmdMod)
	if !strings.Contains(content, "=> ..") {
		t.Fatalf("nested cmd replace for parent must remain => .. ; go.mod:\n%s", content)
	}
}

// enableIsolatedReinstallGOBIN sets GOBIN={WorkRoot}/gobin and optionally
// touches a present bin stub so Action=install (filter requires present bin).
// binName empty → only create GOBIN dir (no install candidates required).
func enableIsolatedReinstallGOBIN(t *testing.T, req *Request, binName string) {
	t.Helper()
	if req.WorkRoot == "" {
		t.Fatal("WorkRoot required for isolated GOBIN")
	}
	gobin := filepath.Join(req.WorkRoot, "gobin")
	mkdirAll(t, gobin)
	if binName != "" {
		writeFile(t, filepath.Join(gobin, binName), "stub\n")
	}
	// Prefer a single GOBIN entry (overwrite prior if any).
	filtered := make([]string, 0, len(req.ExtraEnv)+1)
	for _, e := range req.ExtraEnv {
		if strings.HasPrefix(e, "GOBIN=") {
			continue
		}
		filtered = append(filtered, e)
	}
	req.ExtraEnv = append(filtered, "GOBIN="+gobin)
}

// toolsGoModPath returns MainRepo/tools/go.mod.
func toolsGoModPath(req *Request) string {
	return filepath.Join(req.MainRepo, cascadeToolsDir, "go.mod")
}

// assertNestedSkipConsumerPinned checks tools/go.mod: require at ExpectedPinVersion
// and local replace for LeafModulePath (shared) still present. Root go.mod must
// not gain a shared require (consumer is nested tools only).
func assertNestedSkipConsumerPinned(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo == "" || req.LeafModulePath == "" || req.ExpectedPinVersion == "" {
		t.Fatal("MainRepo, LeafModulePath, ExpectedPinVersion required")
	}
	toolsMod := toolsGoModPath(req)
	got := requireVersionInGoMod(t, toolsMod, req.LeafModulePath)
	if got != req.ExpectedPinVersion {
		t.Fatalf("nested skip consumer require %s = %q, want %s (cascade pin bump)\n%s:\n%s",
			req.LeafModulePath, got, req.ExpectedPinVersion, toolsMod, readFile(t, toolsMod))
	}
	if !goModHasReplace(t, toolsMod, req.LeafModulePath) {
		t.Fatalf("nested skip consumer must KEEP local replace for %s:\n%s",
			req.LeafModulePath, readFile(t, toolsMod))
	}
	// Root must not be forced into the shared require (pin target is tools only).
	rootMod := filepath.Join(req.MainRepo, "go.mod")
	if ver := requireVersionInGoMod(t, rootMod, req.LeafModulePath); ver != "" {
		t.Fatalf("root go.mod must NOT require %s (nested tools owns the edge); got %q\n%s",
			req.LeafModulePath, ver, readFile(t, rootMod))
	}
}

// assertNoCascadeTagNextForModule fails if apply/dry-run stdout plans/tags that module.
func assertNoCascadeTagNextForModule(t *testing.T, stdout, modulePath string) {
	t.Helper()
	// Dry-run form and apply form.
	for _, needle := range []string{
		"would: tag-next " + modulePath + " @",
		"tag-next " + modulePath + " @",
	} {
		if strings.Contains(stdout, needle) {
			t.Fatalf("module %s is skip consumer — must not get cascade tag-next (%q)\nstdout:\n%s",
				modulePath, needle, stdout)
		}
	}
}

// assertNoTidyOrUnknownRevisionFail fails on classic pre-cascade reinstall/tidy
// failure surfaces (unknown revision, go mod tidy required / failed).
func assertNoTidyOrUnknownRevisionFail(t *testing.T, resp *Response) {
	t.Helper()
	combined := resp.Stdout + "\n" + resp.Stderr
	lower := strings.ToLower(combined)
	if strings.Contains(lower, "unknown revision") {
		t.Fatalf("reinstall/cascade must not surface unknown revision\ncombined:\n%s", combined)
	}
	// "go mod tidy required" / updates to go.mod / failed to execute go mod tidy
	if strings.Contains(lower, "go mod tidy required") ||
		strings.Contains(lower, "updates to go.mod needed") ||
		strings.Contains(lower, "failed to execute go mod tidy") {
		t.Fatalf("reinstall/cascade must not fail with go mod tidy diagnostics\ncombined:\n%s", combined)
	}
}

// assertReinstallTailNoHardFail requires exit 0 already checked; when a summary
// line is present, prefer failed 0. Empty reinstall (no candidates) is OK.
func assertReinstallTailNoHardFail(t *testing.T, resp *Response) {
	t.Helper()
	assertNoTidyOrUnknownRevisionFail(t, resp)
	out := resp.Stdout
	// Soft reinstall summary when plan ran.
	if strings.Contains(out, "reinstalled ") {
		// Prefer no failed installs when summary is printed.
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "reinstalled ") {
				continue
			}
			if strings.Contains(line, "failed 0") {
				return
			}
			// Soft-fail installs still exit 0 product-side; still report for
			// regression visibility when we seeded a GOBIN candidate.
			if strings.Contains(line, "failed ") && !strings.Contains(line, "failed 0") {
				t.Fatalf("reinstall summary must report failed 0 after cascade pin (got %q)\nstdout:\n%s\nstderr:\n%s",
					line, resp.Stdout, resp.Stderr)
			}
		}
	}
}

// assertReinstallInstalledAtLeastOne requires a successful go install of the
// seeded nested tools binary (stricter C-RI1 path).
func assertReinstallInstalledAtLeastOne(t *testing.T, resp *Response) {
	t.Helper()
	out := resp.Stdout
	if !strings.Contains(out, "go install") && !strings.Contains(out, "reinstalled ") {
		t.Fatalf("expected reinstall-local to run installs (go install / reinstalled summary)\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	// Summary preferred when present.
	if strings.Contains(out, "reinstalled ") {
		ok := false
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "reinstalled ") &&
				!strings.HasPrefix(line, "reinstalled 0,") &&
				strings.Contains(line, "failed 0") {
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("want reinstalled N>=1 with failed 0\nstdout:\n%s\nstderr:\n%s",
				resp.Stdout, resp.Stderr)
		}
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	_ = setupApplyCascadeNestedSkipConsumer
	_ = setupApplyCascadeNestedCmdRequiresParent
	_ = enableIsolatedReinstallGOBIN
	_ = toolsGoModPath
	_ = assertNestedSkipConsumerPinned
	_ = assertNestedCmdRequiresParentPinned
	_ = assertNoCascadeTagNextForModule
	_ = assertNoTidyOrUnknownRevisionFail
	_ = assertReinstallTailNoHardFail
	_ = assertReinstallInstalledAtLeastOne
	_ = cascadeToolsModule
	_ = cascadeToolsDir
	_ = cascadeToolsBin
	_ = cascadeCmdHarnessModule
	_ = cascadeCmdDir
	_ = cascadeCmdBin
	_ = cascadeExtraModule
	_ = cascadeExtraOld
	_ = cascadeExtraNext
	_ = freeMainCmdHarnessGoModPath
	_ = seedExtraVersions
	_ = seedFreeParentVersions
	return nil
}
```
