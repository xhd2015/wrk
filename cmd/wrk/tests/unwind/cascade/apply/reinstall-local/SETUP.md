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
	_ = enableIsolatedReinstallGOBIN
	_ = toolsGoModPath
	_ = assertNestedSkipConsumerPinned
	_ = assertNoCascadeTagNextForModule
	_ = assertNoTidyOrUnknownRevisionFail
	_ = assertReinstallTailNoHardFail
	_ = assertReinstallInstalledAtLeastOne
	_ = cascadeToolsModule
	_ = cascadeToolsDir
	_ = cascadeToolsBin
	return nil
}
```
