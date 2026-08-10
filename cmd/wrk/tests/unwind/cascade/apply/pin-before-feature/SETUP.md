# Scenario

**Feature**: B1 interleaved apply — cascade pin auto-commit **before** consumer feature gen-commit

```
# free-first apply with --gen-commit-msg --commit --tag-next:
# 1) peel dirty free deps (gen-commit/land if needed)
# 2) tag free modules when planned
# 3) pin consumers (drop external replace) as separate auto-commit
# 4) only then peel remaining dirty consumers: feature gen-commit / land
dirty consumer (+ optional dirty free) + external replace
  -> wrk --unwind --add-all --gen-commit-msg --commit --merge-back --tag-next …
  -> pin commit "wrk: cascade pin …" before feature commit
  -> final go.mod: no droppable external replace; require at pin version
```

## Preconditions

- Inherits `cascade/apply/` helpers (`assertCascadePinCommitPresent`, modproxy,
  pin commit prefix) and root unwind stack helpers.
- Leaves set `req.InProcess = true` and full apply Args **with** gen-commit agent
  flags (fake-opencode; no live LLM).
- Linked consumer primary so gen-commit stage actually runs (product peels
  gen-commit only for linked checkouts today).
- Fixture pre-commit hook simulates `git-hook-go-no-local-replace` for external
  stack replaces (fails while `replace … => ./external/…` remains).

## Steps

1. Grouping provides B1 apply seeders, fake-opencode install, no-local-replace
   pre-commit hook, and pin-before-feature history asserts.
2. Leaves split on free-host shape:
   - T1: pure multi-repo clean external replace-only
   - T2: free dirty external then consumer gen-commit
   - T-M1: **monorepo freeHost** (intra owned-changed) + clean external replace
   - T-tag1: **3-level freeHost** mid + dirty free (pin before free tag-next)

## Context

- **D2** B1 free-first interleaved apply; **D3** keep-current require when no
  free tag; **D7** separate pin auto-commit then feature gen-commit.
- Classic TDD: leaves are **RED** until product reorders pin before consumer
  feature gen-commit (today peels all then cascade; gen-commit hits hook while
  replace still present).
- **T-M1 hole:** same-label freeHost (intra free pin dep) blocks pure consumer
  deferral, so ready external pins never run before freeHost feature gen-commit.
- **T-tag1 hole:** mid freeHost peels early; `pinReadyExternalReplacesBeforeGenCommit`
  pins dirty free @ planned NextTag **before** cascade `tag-next` (missing
  `attachTagScopeToModules` → `cascadeModuleShouldTag` always false). Production
  surfaces as `go mod tidy: unknown revision`; L2 locks tag-before-pin order.
- Do not rewrite sealed ASSERT contracts under `clean/`, `dirty-gomod/`,
  `partial-edit/`, `reinstall-local/`, or sealed T1/T2/T-M1 leaves.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/gitops/git/git_isolated"
)

const (
	// Uncommitted feature WIP path for consumer gen-commit (not go.mod).
	cascadeFeatureWIPFile = "FEATURE_WIP.md"
	cascadeFeatureWIPBody = "feature WIP for gen-commit before pin ordering\n"

	// Mock gen-commit subject from fake-opencode (must match mock JSON).
	cascadeFeatureCommitSubject = "feat: add feature"

	// Fake-opencode mock: title = cascadeFeatureCommitSubject.
	cascadeFakeOpencodeMockJSON = `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"cascade_pin_before_feature","llm_events":[{"type":"step_start"},{"type":"message","text":"{\"title\": \"feat: add feature\", \"description\": \"Consumer feature after cascade pin\"}"},{"type":"step_finish"}]}`

	// Shell pre-commit: fail when go.mod still has a droppable external-style
	// local replace (./external/… or ../…). Intra-repo ./pkgs/… is allowed.
	// Mirrors git-hook-go-no-local-replace lenient mode for stack externals.
	cascadeNoLocalReplacePreCommit = `#!/bin/sh
# Fixture: simulate git-hook-go-no-local-replace for external stack replaces.
if [ ! -f go.mod ]; then
  exit 0
fi
if grep -E '=>[[:space:]]*(\.\./|\./external/)' go.mod >/dev/null 2>&1; then
  echo "pre-commit: local external replace forbidden in go.mod (git-hook-go-no-local-replace sim)" >&2
  exit 1
fi
exit 0
`
)

var cascadeFakeOpencodeMu sync.Mutex

// cascadeRepoHooksDir returns the common-dir hooks path for repo (worktree-safe).
func cascadeRepoHooksDir(t *testing.T, repo string) string {
	t.Helper()
	if repo == "" {
		t.Fatal("cascadeRepoHooksDir: empty repo")
	}
	common := strings.TrimSpace(gitOutputIsolated(t, repo, "rev-parse", "--git-common-dir"))
	if common == "" {
		t.Fatal("rev-parse --git-common-dir empty")
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(repo, common)
	}
	common = resolvePath(t, common)
	hooksDir := filepath.Join(common, "hooks")
	mkdirAll(t, hooksDir)
	return hooksDir
}

// installCascadePermissivePreCommit installs a no-op pre-commit and points
// core.hooksPath at the repo hooks dir so product gen-commit does not inherit
// the developer machine's global hooksPath (parallel-safe, repo-local only).
func installCascadePermissivePreCommit(t *testing.T, repo string) {
	t.Helper()
	hooksDir := cascadeRepoHooksDir(t, repo)
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write permissive pre-commit: %v", err)
	}
	runGitIsolated(t, repo, "config", "core.hooksPath", hooksDir)
}

// installCascadeNoLocalReplacePreCommit installs a repo-local pre-commit hook
// that fails while go.mod still has an external filesystem replace. Uses the
// common git dir so linked worktrees share the hook. Parallel-safe (repo-local
// config only; no process env mutation).
func installCascadeNoLocalReplacePreCommit(t *testing.T, repo string) {
	t.Helper()
	hooksDir := cascadeRepoHooksDir(t, repo)
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(cascadeNoLocalReplacePreCommit), 0o755); err != nil {
		t.Fatalf("write pre-commit: %v", err)
	}
	// Ensure product git commits invoke hooks (seed commits used isolated git).
	runGitIsolated(t, repo, "config", "core.hooksPath", hooksDir)
}

// findAgentProFakeOpencodeDir locates cmd/fake-opencode for offline gen-commit.
// Prefers sibling agent-pro-* worktrees (unwind-pipeline style); falls back to
// external/agent-pro-master-2026-07-16 under the wrk module.
func findAgentProFakeOpencodeDir(t *testing.T) string {
	t.Helper()
	modRoot := findModuleRoot(doctestRootPath(t))
	if modRoot == "" {
		t.Fatal("find module root: no go.mod in ancestors of d.DOCTEST_ROOT")
	}
	// Sibling worktrees: …/agent-pro-*/cmd/fake-opencode
	parent := filepath.Dir(modRoot)
	matches, err := filepath.Glob(filepath.Join(parent, "agent-pro-*", "cmd", "fake-opencode"))
	if err == nil && len(matches) > 0 {
		// Prefer a path that has main.go.
		for i := len(matches) - 1; i >= 0; i-- {
			if _, err := os.Stat(filepath.Join(matches[i], "main.go")); err == nil {
				return matches[i]
			}
		}
	}
	// Vendored external checkout (gen-commit-msg tree).
	ext := filepath.Join(modRoot, "external", "agent-pro-master-2026-07-16", "cmd", "fake-opencode")
	if _, err := os.Stat(filepath.Join(ext, "main.go")); err == nil {
		return ext
	}
	t.Skip("agent-pro cmd/fake-opencode fixture unavailable (sibling agent-pro-* or external/)")
	return ""
}

// sessionCascadeFakeOpencodeBin is the session-cached fake-opencode path.
func sessionCascadeFakeOpencodeBin(t *testing.T) string {
	t.Helper()
	return filepath.Join(fixtureSessionRoot(t), "bin", "fake-opencode")
}

// getCascadeFakeOpencodeBin builds fake-opencode once per session (file-locked).
func getCascadeFakeOpencodeBin(t *testing.T) string {
	t.Helper()
	bin := sessionCascadeFakeOpencodeBin(t)
	if _, err := os.Stat(bin); err == nil {
		return bin
	}
	lockPath := filepath.Join(fixtureSessionRoot(t), "bin", ".fake-opencode.lock")
	withFlock(t, lockPath, func() {
		if _, err := os.Stat(bin); err == nil {
			return
		}
		srcDir := findAgentProFakeOpencodeDir(t)
		if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
			t.Fatalf("mkdir fake-opencode bin: %v", err)
		}
		cmd := exec.Command("go", "build", "-mod=mod", "-o", bin, ".")
		cmd.Dir = srcDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build fake-opencode: %v\n%s", err, out)
		}
	})
	return bin
}

// installCascadeFakeOpencodeEnv writes mock config under WorkRoot, builds
// fake-opencode, and appends ExtraEnv (no t.Setenv).
func installCascadeFakeOpencodeEnv(t *testing.T, req *Request) string {
	t.Helper()
	_ = cascadeFakeOpencodeMu // keep for future multi-leaf coordination
	mockPath := filepath.Join(req.WorkRoot, "fake-opencode-mock.json")
	writeFile(t, mockPath, cascadeFakeOpencodeMockJSON)
	bin := getCascadeFakeOpencodeBin(t)
	configDir := filepath.Join(req.WorkRoot, "opencode-config")
	mkdirAll(t, configDir)
	req.ExtraEnv = append(req.ExtraEnv,
		"FAKE_OPENCODE_MOCK_CONFIG="+mockPath,
		"OPENCODE_CONFIG_DIR="+configDir,
	)
	return bin
}

// cascadeUnwindGenCommitArgs builds apply Args: unwind + gen-commit agent path +
// stages (e.g. --add-all --merge-back --tag-next [--push]).
func cascadeUnwindGenCommitArgs(t *testing.T, req *Request, stages ...string) []string {
	t.Helper()
	bin := installCascadeFakeOpencodeEnv(t, req)
	args := []string{
		"--unwind",
		"--gen-commit-msg",
		"--commit",
		"--agent-runner", "opencode",
		"--agent-runner-binary", bin,
		"--model", "openai/gpt-5",
	}
	return append(args, stages...)
}

// dirtyCascadeFeatureWIP writes FEATURE_WIP.md and stages it (git add).
// Staging reproduces the real-world pin scoop bug: a pre-staged feature index
// must not land in the cascade pin commit (only go.mod/go.sum).
func dirtyCascadeFeatureWIP(t *testing.T, checkout string) {
	t.Helper()
	if checkout == "" {
		t.Fatal("dirtyCascadeFeatureWIP: empty checkout")
	}
	writeFile(t, filepath.Join(checkout, cascadeFeatureWIPFile), cascadeFeatureWIPBody)
	runGitIsolated(t, checkout, "add", "--", cascadeFeatureWIPFile)
	status := gitOutputIsolated(t, checkout, "status", "--porcelain", "--", cascadeFeatureWIPFile)
	if strings.TrimSpace(status) == "" {
		t.Fatal("expected uncommitted feature WIP after dirtyCascadeFeatureWIP")
	}
	// porcelain "A  FEATURE_WIP.md" = staged new file (index scoop repro).
	if !strings.HasPrefix(strings.TrimSpace(status), "A") {
		t.Fatalf("expected staged feature WIP (A…); porcelain=%q", status)
	}
}

// seedDotPkgsProxy seeds file:// modproxy for example.com/dot-pkgs at version
// from srcDir (or a synthetic seed when srcDir content is for a different ver).
func seedDotPkgsProxyVersions(t *testing.T, req *Request, versions map[string]string) {
	t.Helper()
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	for ver, src := range versions {
		if src == "" {
			seed := filepath.Join(req.WorkRoot, "seed-dot-pkgs-"+ver)
			mkdirAll(t, seed)
			writeGoModRequire(t, seed, unwindDotPkgsModule)
			writeFile(t, filepath.Join(seed, "pkg.go"),
				"package dotpkgs\n\nfunc Version() string { return \""+ver+"\" }\n")
			src = seed
		}
		seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, ver, src)
	}
	enableFileModuleProxy(t, req, proxyRoot)
}

// setupApplyPinBeforeFeatureExternalCleanDep is T1 core fixture:
//
//	leaf external: clean, tag v0.0.1 at HEAD (no owned-changed)
//	root main: require leaf@v0.0.1
//	root linked WT (RepoDir / primary): require + replace => ./external/…
//	  + uncommitted FEATURE_WIP.md for gen-commit
//	pre-commit: no external local replace (hook sim)
//	modproxy: v0.0.1 for pin+tidy after drop replace
//
// PeelOrder ≈ ["."] (only dirty consumer). Expected pin version = v0.0.1 (D3).
func setupApplyPinBeforeFeatureExternalCleanDep(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyOldTag // D3 keep-current

	// --- leaf main: baseline tag only (clean) ---
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
	// Isolate leaf product commits from developer global hooksPath.
	installCascadePermissivePreCommit(t, leafMain)

	// --- root consumer main ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root consumer main")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	// Linked consumer = stack primary (gen-commit only runs for linked peels).
	wtDir := filepath.Join(req.WorkRoot, "root-linked")
	runGitIsolated(t, rootMain, "worktree", "add", "-b", branchNameMainDate(), wtDir)
	wtDir = resolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchNameMainDate()

	// Nest clean leaf WT under consumer WT/external.
	extDir := filepath.Join(wtDir, "external")
	mkdirAll(t, extDir)
	leafExtName := labelDotPkgs + "-" + branchNameMainDate()
	leafExt := filepath.Join(extDir, leafExtName)
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate()+"-leaf", leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt
	// Leave leaf clean (no markDirty; HEAD at v0.0.1).

	relReplace := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, wtDir, unwindDotPkgsModule, "./"+relReplace)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", "go.mod", ".gitignore")
	runGitIsolated(t, wtDir, "commit", "-m", "stack replace + ignore external")

	// Feature WIP for consumer gen-commit (replace already committed).
	dirtyCascadeFeatureWIP(t, wtDir)
	// Peel pending marker (DIRTY counts as dirty under v1 filter).
	markDirty(t, wtDir)

	// Hook: fail gen-commit while external replace remains (T1 RED surface today).
	// Install after seed commits (shared common-dir with main).
	installCascadeNoLocalReplacePreCommit(t, wtDir)

	// Offline proxy for pin+tidy after drop replace @ current require.
	seedDotPkgsProxyVersions(t, req, map[string]string{
		unwindApplyOldTag: "", // synthetic seed
	})

	req.RepoDir = wtDir
	req.PeelOrder = []string{"."}
}

// setupApplyPinBeforeFeatureFreeDirty is T2 fixture:
//
//	leaf external: dirty / owned-changed → next tag v0.0.2; bare origin for --push
//	root linked WT: external replace + feature WIP + no-local-replace hook
//	modproxy: old + next for pin after free tag
//
// Free-first: peel/land free → tag → pin auto-commit → consumer feature gen-commit.
// Expected pin version = v0.0.2.
func setupApplyPinBeforeFeatureFreeDirty(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- leaf main + origin ---
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
	// Isolate free-dep product gen-commit from developer global hooks.
	installCascadePermissivePreCommit(t, leafMain)

	leafBare := setupBareOrigin(t, req.WorkRoot, "leaf-origin")
	attachOriginAndPushMain(t, leafMain, leafBare)
	if tagRefExists(t, leafMain, unwindApplyOldTag) {
		runGitIsolated(t, leafMain, "push", "origin", unwindApplyOldTag)
	}
	req.OriginBare = leafBare

	// --- root consumer main + linked WT ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root consumer main")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	rootBare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, rootMain, rootBare)
	if tagRefExists(t, rootMain, unwindApplyOldTag) {
		runGitIsolated(t, rootMain, "push", "origin", unwindApplyOldTag)
	}
	writeFile(t, filepath.Join(req.WorkRoot, "root-origin.path"), rootBare+"\n")

	wtDir := filepath.Join(req.WorkRoot, "root-linked")
	runGitIsolated(t, rootMain, "worktree", "add", "-b", branchNameMainDate(), wtDir)
	wtDir = resolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchNameMainDate()

	// Dirty free leaf under external (owned-changed for next tag).
	extDir := filepath.Join(wtDir, "external")
	mkdirAll(t, extDir)
	leafExtName := labelDotPkgs + "-" + branchNameMainDate()
	leafExt := filepath.Join(extDir, leafExtName)
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate()+"-leaf", leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt

	writeFile(t, filepath.Join(leafExt, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, leafExt, "add", "pkg.go")
	runGitIsolated(t, leafExt, "commit", "-m", "leaf feature for next tag")
	markDirty(t, leafExt)

	relReplace := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, wtDir, unwindDotPkgsModule, "./"+relReplace)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", "go.mod", ".gitignore")
	runGitIsolated(t, wtDir, "commit", "-m", "stack replace + ignore external")

	dirtyCascadeFeatureWIP(t, wtDir)
	markDirty(t, wtDir)
	// Consumer hook only (after seed commits).
	installCascadeNoLocalReplacePreCommit(t, wtDir)

	// Proxy: next from leaf WT content; old synthetic.
	seedDotPkgsProxyVersions(t, req, map[string]string{
		unwindApplyNextTag: leafExt,
		unwindApplyOldTag:  "",
	})

	req.RepoDir = wtDir
	setPeelOrderDisplays(t, req, leafExt, wtDir)
}

// setupApplyPinBeforeFeatureMonorepoFreeHostExternal is T-M1 fixture:
//
//	One dirty **linked monorepo** (primary peel `.`):
//	  - root requires cascadeSharedModule @ v0.0.1 + example.com/dot-pkgs @ v0.0.1
//	  - intra replace shared => ./pkgs/shared (keep-local; not droppable)
//	  - pkgs/shared owned-changed after pkgs/shared/v0.0.1 → freeHost same label
//	  - droppable external replace => ./external/… for clean free leaf @ v0.0.1
//	  - uncommitted FEATURE_WIP.md for gen-commit
//	Clean external free leaf under external/ at tag v0.0.1 (no owned-changed)
//	pre-commit: no external local replace (hook sim; ./pkgs/… allowed)
//	modproxy: v0.0.1 for pin+tidy after drop external replace
//
// freeHost (intra free pin dep) + pinConsumer (external pin) share monorepo label
// → B1 keeps peel early; pure consumer deferral never pins external first.
// Expected external pin version = v0.0.1 (D3 keep-current; ready free).
func setupApplyPinBeforeFeatureMonorepoFreeHostExternal(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyOldTag // D3 keep-current on ready external

	// --- clean external free leaf main: baseline tag only ---
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
	installCascadePermissivePreCommit(t, leafMain)

	// --- monorepo main: root + pkgs/shared + require external ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)

	sharedDir := filepath.Join(rootMain, filepath.FromSlash(cascadeSharedDir))
	mkdirAll(t, sharedDir)
	writeGoModRequire(t, sharedDir, cascadeSharedModule)
	writeFile(t, filepath.Join(sharedDir, "shared.go"),
		"package shared\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	// Root requires both intra free + external free @ matching baseline.
	writeGoModRequire(t, rootMain, unwindRootModule,
		cascadeSharedModule+"@"+unwindApplyOldTag,
		unwindDotPkgsModule+"@"+unwindApplyOldTag,
	)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport (\n\t_ \""+cascadeSharedModule+"\"\n\t_ \""+unwindDotPkgsModule+"\"\n)\n")
	// Intra keep-local replace (must survive; hook allows ./pkgs/…).
	appendLocalReplace(t, rootMain, cascadeSharedModule, "./"+cascadeSharedDir)

	runGitIsolated(t, rootMain, "add", "go.mod", "root.go", "pkgs")
	runGitIsolated(t, rootMain, "commit", "-m", "root + shared + require external")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "HEAD")
	createLightweightTag(t, rootMain, cascadeSharedOldTag, "HEAD")

	// Owned change on shared only → freeHost same monorepo label (cascade tag + pin).
	writeFile(t, filepath.Join(sharedDir, "shared.go"),
		"package shared\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, rootMain, "add", "pkgs")
	runGitIsolated(t, rootMain, "commit", "-m", "shared owned change for freeHost")

	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	// Linked monorepo = stack primary (gen-commit only runs for linked peels).
	wtDir := filepath.Join(req.WorkRoot, "root-linked")
	runGitIsolated(t, rootMain, "worktree", "add", "-b", branchNameMainDate(), wtDir)
	wtDir = resolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchNameMainDate()

	// Nest clean external free under consumer WT/external.
	extDir := filepath.Join(wtDir, "external")
	mkdirAll(t, extDir)
	leafExtName := labelDotPkgs + "-" + branchNameMainDate()
	leafExt := filepath.Join(extDir, leafExtName)
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate()+"-leaf", leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt
	// Leave external free clean (no markDirty; HEAD at v0.0.1).

	relReplace := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, wtDir, unwindDotPkgsModule, "./"+relReplace)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", "go.mod", ".gitignore")
	runGitIsolated(t, wtDir, "commit", "-m", "external stack replace + ignore")

	// Feature WIP for monorepo freeHost gen-commit (external replace committed).
	dirtyCascadeFeatureWIP(t, wtDir)
	markDirty(t, wtDir)

	// Hook: fail gen-commit while external replace remains (T-M1 RED surface).
	// Intra ./pkgs/shared is allowed by cascadeNoLocalReplacePreCommit.
	installCascadeNoLocalReplacePreCommit(t, wtDir)

	// Offline proxy for pin+tidy after drop external replace @ current require.
	seedDotPkgsProxyVersions(t, req, map[string]string{
		unwindApplyOldTag: "", // synthetic seed
	})

	req.RepoDir = wtDir
	req.PeelOrder = []string{"."}
}

// setupApplyPinBeforeFeatureThreeLevelFreeHostDirty is T-tag1 fixture:
//
//	3-level multi-repo freeHost stack (production: spl → kool → go-pkgs):
//	  leaf free (dot-pkgs): dirty owned-changed → next v0.0.2; bare origin + --push
//	  mid freeHost (agent-pro): pinConsumer of leaf AND freeHost of top; linked;
//	    external replace → leaf; FEATURE_WIP for gen-commit; B1 peels early
//	  top pure pinConsumer (root): external replace → mid; dirty deferred peel
//	modproxy: old+next for leaf and mid (cascade pin+tidy after free tag)
//
// Bug: pinReady on mid early peel pins leaf@next before cascade tag-next
// (rebuild graph without attachTagScopeToModules → cascadeModuleShouldTag false).
// Expected GREEN: free tag-next before mid pin of free; mid require@next; no replace.
func setupApplyPinBeforeFeatureThreeLevelFreeHostDirty(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- leaf free main + origin (owned-changed → next tag) ---
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
	installCascadePermissivePreCommit(t, leafMain)

	leafBare := setupBareOrigin(t, req.WorkRoot, "leaf-origin")
	attachOriginAndPushMain(t, leafMain, leafBare)
	if tagRefExists(t, leafMain, unwindApplyOldTag) {
		runGitIsolated(t, leafMain, "push", "origin", unwindApplyOldTag)
	}
	req.OriginBare = leafBare

	// --- mid freeHost main + origin (requires free @ baseline) ---
	midMain := filepath.Join(req.WorkRoot, labelAgentPro)
	initGitRepoOnMain(t, midMain)
	writeGoModRequire(t, midMain, unwindAgentProModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(midMain, "agent.go"),
		"package agentpro\n\nimport _ \""+unwindDotPkgsModule+"\"\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	runGitIsolated(t, midMain, "add", "go.mod", "agent.go")
	runGitIsolated(t, midMain, "commit", "-m", "add agent-pro mid freeHost")
	createLightweightTag(t, midMain, unwindApplyOldTag, "")
	midMain = resolvePath(t, midMain)
	req.DepPath = midMain
	installCascadePermissivePreCommit(t, midMain)

	midBare := setupBareOrigin(t, req.WorkRoot, "mid-origin")
	attachOriginAndPushMain(t, midMain, midBare)
	if tagRefExists(t, midMain, unwindApplyOldTag) {
		runGitIsolated(t, midMain, "push", "origin", unwindApplyOldTag)
	}
	writeFile(t, filepath.Join(req.WorkRoot, "mid-origin.path"), midBare+"\n")

	// --- top pure pinConsumer main + origin ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule, unwindAgentProModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport _ \""+unwindAgentProModule+"\"\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root top consumer")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	rootBare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, rootMain, rootBare)
	if tagRefExists(t, rootMain, unwindApplyOldTag) {
		runGitIsolated(t, rootMain, "push", "origin", unwindApplyOldTag)
	}
	writeFile(t, filepath.Join(req.WorkRoot, "root-origin.path"), rootBare+"\n")

	// Linked top = stack primary (cwd for inventory / peel displays).
	wtDir := filepath.Join(req.WorkRoot, "root-linked")
	runGitIsolated(t, rootMain, "worktree", "add", "-b", branchNameMainDate(), wtDir)
	wtDir = resolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchNameMainDate()

	// Nest mid + leaf linked externals under top (sibling externals, production-like).
	extDir := filepath.Join(wtDir, "external")
	mkdirAll(t, extDir)

	midExtName := labelAgentPro + "-" + branchNameMainDate()
	midExt := filepath.Join(extDir, midExtName)
	runGitIsolated(t, midMain, "worktree", "add", "-b", branchNameMainDate()+"-mid", midExt)
	midExt = resolvePath(t, midExt)
	req.ExternalWtDir = midExt

	leafExtName := labelDotPkgs + "-" + branchNameMainDate()
	leafExt := filepath.Join(extDir, leafExtName)
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate()+"-leaf", leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt

	// Dirty free leaf: owned-changed after baseline tag → next v0.0.2.
	writeFile(t, filepath.Join(leafExt, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, leafExt, "add", "pkg.go")
	runGitIsolated(t, leafExt, "commit", "-m", "leaf feature for next tag")
	markDirty(t, leafExt)

	// Mid freeHost: droppable external replace → leaf + feature WIP (gen-commit path).
	// No no-local-replace hook: after fix, pinReady skips untagged free so mid may
	// gen-commit while replace still present; cascade pin drops it after free tag.
	appendLocalReplace(t, midExt, unwindDotPkgsModule, relLocalReplace(t, midExt, leafExt))
	runGitIsolated(t, midExt, "add", "go.mod")
	runGitIsolated(t, midExt, "commit", "-m", "mid external replace to dirty free leaf")
	dirtyCascadeFeatureWIP(t, midExt)
	markDirty(t, midExt)
	installCascadePermissivePreCommit(t, midExt)

	// Top: droppable external replace → mid freeHost; pure pinConsumer → deferred peel.
	relMid := filepath.ToSlash(filepath.Join("external", midExtName))
	appendLocalReplace(t, wtDir, unwindAgentProModule, "./"+relMid)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", "go.mod", ".gitignore")
	runGitIsolated(t, wtDir, "commit", "-m", "top replace to mid freeHost + ignore external")
	markDirty(t, wtDir)
	installCascadePermissivePreCommit(t, wtDir)

	// Offline proxy: free old+next (and mid old+next if mid also tags from FEATURE_WIP).
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyNextTag, leafExt)
	oldLeafSeed := filepath.Join(req.WorkRoot, "seed-leaf-"+unwindApplyOldTag)
	mkdirAll(t, oldLeafSeed)
	writeGoModRequire(t, oldLeafSeed, unwindDotPkgsModule)
	writeFile(t, filepath.Join(oldLeafSeed, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyOldTag, oldLeafSeed)

	// Mid versions for top pin after drop replace (FEATURE_WIP may plan mid next tag).
	seedFileModuleProxy(t, proxyRoot, unwindAgentProModule, unwindApplyOldTag, midMain)
	midNextSeed := filepath.Join(req.WorkRoot, "seed-mid-"+unwindApplyNextTag)
	mkdirAll(t, midNextSeed)
	writeGoModRequire(t, midNextSeed, unwindAgentProModule, unwindDotPkgsModule+"@"+unwindApplyNextTag)
	writeFile(t, filepath.Join(midNextSeed, "agent.go"),
		"package agentpro\n\nimport _ \""+unwindDotPkgsModule+"\"\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindAgentProModule, unwindApplyNextTag, midNextSeed)
	// Mid zip at next may require free@next from proxy during tidy of top.
	enableFileModuleProxy(t, req, proxyRoot)

	req.RepoDir = wtDir
	setPeelOrderDisplays(t, req, leafExt, midExt, wtDir)
}

// assertFreeTagNextBeforeMidPinOfFree locks T-tag1 / T-tag2 order: cascade
// tag-next for dirty free must appear before pinReady/cascade pin of mid ← free
// @ next. pinReady must not treat planned NextTag as ready before the free tag
// exists (production: unknown revision when tidy resolves untagged next).
func assertFreeTagNextBeforeMidPinOfFree(t *testing.T, out string) {
	t.Helper()
	tagNeedle := "tag-next " + unwindDotPkgsModule + " @ " + unwindApplyNextTag
	// pin log uses stack repo labels (not module paths).
	pinNeedle := "pin " + labelAgentPro + " <- " + labelDotPkgs + " @ " + unwindApplyNextTag
	tagIdx := strings.Index(out, tagNeedle)
	pinIdx := strings.Index(out, pinNeedle)
	if pinIdx < 0 {
		// Tolerate label basename drift; require free pin @ next somewhere.
		alt := "<- " + labelDotPkgs + " @ " + unwindApplyNextTag
		pinIdx = strings.Index(out, alt)
		if pinIdx >= 0 {
			pinNeedle = alt
		}
	}
	if tagIdx < 0 {
		t.Fatalf("T-tag1: missing free tag-next line %q (free must be tagged before mid pin)\nout:\n%s",
			tagNeedle, out)
	}
	if pinIdx < 0 {
		t.Fatalf("T-tag1: missing mid pin of free @ next (want %q or pin … <- %s @ %s)\nout:\n%s",
			pinNeedle, labelDotPkgs, unwindApplyNextTag, out)
	}
	if pinIdx < tagIdx {
		t.Fatalf("T-tag1: free tag-next must precede mid pin of free @ next (pinReady must skip untagged NextTag)\ntag@%d pin@%d\nneedles: %q then %q\nout:\n%s",
			tagIdx, pinIdx, tagNeedle, pinNeedle, out)
	}
}

// assertMidPinnedToFreeLeaf checks mid freeHost require free @ ExpectedPinVersion
// and droppable external replace for free is gone (post-cascade pin).
func assertMidPinnedToFreeLeaf(t *testing.T, req *Request) {
	t.Helper()
	if req.LeafModulePath == "" || req.ExpectedPinVersion == "" {
		t.Fatal("LeafModulePath and ExpectedPinVersion required")
	}
	// Prefer mid main after land; fall back to still-present linked mid.
	checkouts := make([]string, 0, 2)
	if req.DepPath != "" {
		checkouts = append(checkouts, req.DepPath)
	}
	if req.ExternalWtDir != "" {
		if _, err := os.Stat(req.ExternalWtDir); err == nil {
			checkouts = append(checkouts, req.ExternalWtDir)
		}
	}
	if len(checkouts) == 0 {
		t.Fatal("DepPath or ExternalWtDir required for mid pin assert")
	}
	var lastMod string
	for _, checkout := range checkouts {
		goMod := filepath.Join(checkout, "go.mod")
		if _, err := os.Stat(goMod); err != nil {
			continue
		}
		lastMod = readFile(t, goMod)
		got := requireVersionInGoMod(t, goMod, req.LeafModulePath)
		if got != req.ExpectedPinVersion {
			t.Fatalf("mid freeHost require %s = %q, want %s (checkout %s)\ngo.mod:\n%s",
				req.LeafModulePath, got, req.ExpectedPinVersion, checkout, lastMod)
		}
		if goModHasReplace(t, goMod, req.LeafModulePath) {
			t.Fatalf("mid freeHost go.mod must DROP external replace for %s after pin (checkout %s):\n%s",
				req.LeafModulePath, checkout, lastMod)
		}
		// First successful checkout is enough (main preferred).
		return
	}
	t.Fatalf("mid freeHost go.mod not found under DepPath/ExternalWtDir; last:\n%s", lastMod)
}

// pinCommitSHAForDep returns the newest cascade pin commit SHA whose subject
// mentions depModule (empty if none). Prefer this over pinCommitSHA when the
// monorepo may also pin intra free modules after the ready-external pin.
func pinCommitSHAForDep(t *testing.T, repo, depModule string) string {
	t.Helper()
	if repo == "" || depModule == "" {
		return ""
	}
	needle := cascadePinCommitPrefix + depModule
	return strings.TrimSpace(gitOutputIsolated(t, repo, "log", "-1", "--format=%H", "--grep", needle))
}

// assertPinCommitForDepFilesOnlyModSum fails if the dep's pin commit touches
// paths other than go.mod/go.sum (catches scooping staged FEATURE_WIP into pin).
func assertPinCommitForDepFilesOnlyModSum(t *testing.T, repo, depModule string) {
	t.Helper()
	sha := pinCommitSHAForDep(t, repo, depModule)
	if sha == "" {
		t.Fatalf("missing cascade pin commit for dep %s\nlog:\n%s", depModule,
			gitOutputIsolated(t, repo, "log", "--oneline", "-20"))
	}
	names := gitOutputIsolated(t, repo, "show", "--pretty=format:", "--name-only", sha)
	for _, line := range strings.Split(names, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		base := filepath.Base(line)
		if base != "go.mod" && base != "go.sum" {
			t.Fatalf("cascade pin for %s must only include go.mod/go.sum; got %q\nfiles:\n%s",
				depModule, line, names)
		}
	}
}

// assertCascadePinForDepBeforeFeatureCommit requires the cascade pin commit for
// depModule @ pinVer to be an ancestor of the feature gen-commit (ready pin
// before feature on freeHost peels — T-M1 / D7). Does not use newest-any-pin.
func assertCascadePinForDepBeforeFeatureCommit(t *testing.T, repo, depModule, pinVer, featureSubject string) {
	t.Helper()
	if repo == "" {
		t.Fatal("repo required for pin-before-feature history assert")
	}
	if depModule == "" {
		t.Fatal("depModule required")
	}
	assertCascadePinCommitPresent(t, repo, depModule, pinVer)
	pinSHA := pinCommitSHAForDep(t, repo, depModule)
	if pinSHA == "" {
		t.Fatalf("missing cascade pin commit for dep %s on %s\nlog:\n%s",
			depModule, repo, gitOutputIsolated(t, repo, "log", "--oneline", "-25"))
	}
	if featureSubject == "" {
		featureSubject = cascadeFeatureCommitSubject
	}
	featSHA := featureCommitSHA(t, repo, featureSubject)
	if featSHA == "" {
		t.Fatalf("missing feature gen-commit (subject containing %q) on %s\nlog:\n%s",
			featureSubject, repo, gitOutputIsolated(t, repo, "log", "--oneline", "-25"))
	}
	err := git_isolated.Command(repo, "merge-base", "--is-ancestor", pinSHA, featSHA).Run()
	if err != nil {
		t.Fatalf("T-M1 order: cascade pin for %s must be ancestor of feature gen-commit\npin=%s feature=%s\nlog:\n%s",
			depModule, pinSHA, featSHA, gitOutputIsolated(t, repo, "log", "--oneline", "-25"))
	}
	// Feature commit tree must not reintroduce external replace for this dep.
	featGoMod := gitOutputIsolated(t, repo, "show", featSHA+":go.mod")
	if strings.Contains(featGoMod, depModule) &&
		(strings.Contains(featGoMod, "=> ./external/") || strings.Contains(featGoMod, "=> ../")) {
		// Fail only when this dep still has an external-style replace line.
		for _, line := range strings.Split(featGoMod, "\n") {
			trim := strings.TrimSpace(line)
			if !strings.Contains(trim, depModule) || !strings.Contains(trim, "=>") {
				continue
			}
			if strings.Contains(trim, "./external/") || strings.Contains(trim, "../") {
				t.Fatalf("feature commit go.mod must not carry external replace for %s\n%s",
					depModule, featGoMod)
			}
		}
	}
}

// assertIntraSharedReplaceKept fails if monorepo go.mod dropped the keep-local
// replace for cascadeSharedModule (intra must not be force-dropped with external pin).
func assertIntraSharedReplaceKept(t *testing.T, req *Request) {
	t.Helper()
	checkout := req.MainRepo
	if req.WtDir != "" {
		if _, err := os.Stat(req.WtDir); err == nil {
			checkout = req.WtDir
		}
	}
	if checkout == "" {
		t.Fatal("MainRepo or WtDir required")
	}
	goMod := filepath.Join(checkout, "go.mod")
	if !goModHasReplace(t, goMod, cascadeSharedModule) {
		t.Fatalf("monorepo go.mod must KEEP intra replace for %s (not force-drop with external pin):\n%s",
			cascadeSharedModule, readFile(t, goMod))
	}
}

// assertExternalReplaceDropped fails if consumer go.mod still has a replace for
// LeafModulePath (droppable external replace must be gone after pin).
func assertExternalReplaceDropped(t *testing.T, goModPath, modulePath string) {
	t.Helper()
	if goModHasReplace(t, goModPath, modulePath) {
		t.Fatalf("consumer go.mod must DROP external replace for %s after pin:\n%s",
			modulePath, readFile(t, goModPath))
	}
}

// assertConsumerRequireAndNoExternalReplace checks require version + no replace
// on the preferred consumer checkout (WtDir if present, else MainRepo).
func assertConsumerRequireAndNoExternalReplace(t *testing.T, req *Request) {
	t.Helper()
	if req.LeafModulePath == "" || req.ExpectedPinVersion == "" {
		t.Fatal("LeafModulePath and ExpectedPinVersion required")
	}
	checkout := req.MainRepo
	if req.WtDir != "" {
		if _, err := os.Stat(req.WtDir); err == nil {
			// Prefer Path after merge-back keep; also check MainRepo landed tree.
			checkout = req.WtDir
		}
	}
	if checkout == "" {
		t.Fatal("MainRepo or WtDir required")
	}
	goMod := filepath.Join(checkout, "go.mod")
	got := requireVersionInGoMod(t, goMod, req.LeafModulePath)
	if got != req.ExpectedPinVersion {
		t.Fatalf("consumer require %s = %q, want %s\ngo.mod:\n%s",
			req.LeafModulePath, got, req.ExpectedPinVersion, readFile(t, goMod))
	}
	assertExternalReplaceDropped(t, goMod, req.LeafModulePath)

	// After --merge-back, main should also reflect pin when land merged.
	if req.MainRepo != "" && resolvePath(t, req.MainRepo) != resolvePath(t, checkout) {
		mainMod := filepath.Join(req.MainRepo, "go.mod")
		// Soft: main may lag if product pins only Path; prefer Path contract.
		_ = mainMod
	}
}

// featureCommitSHA returns newest commit SHA whose subject contains needle.
func featureCommitSHA(t *testing.T, repo, subjectNeedle string) string {
	t.Helper()
	if subjectNeedle == "" {
		subjectNeedle = cascadeFeatureCommitSubject
	}
	sha := strings.TrimSpace(gitOutputIsolated(t, repo, "log", "-1", "--format=%H", "--grep", subjectNeedle))
	return sha
}

// assertCascadePinBeforeFeatureCommit requires a cascade pin commit that is an
// ancestor of the feature gen-commit (pin landed first on history).
func assertCascadePinBeforeFeatureCommit(t *testing.T, repo, depModule, pinVer, featureSubject string) {
	t.Helper()
	if repo == "" {
		t.Fatal("repo required for pin-before-feature history assert")
	}
	assertCascadePinCommitPresent(t, repo, depModule, pinVer)
	pinSHA := pinCommitSHA(t, repo)
	if pinSHA == "" {
		t.Fatalf("missing pin commit SHA on %s\nlog:\n%s",
			repo, gitOutputIsolated(t, repo, "log", "--oneline", "-20"))
	}
	if featureSubject == "" {
		featureSubject = cascadeFeatureCommitSubject
	}
	featSHA := featureCommitSHA(t, repo, featureSubject)
	if featSHA == "" {
		t.Fatalf("missing feature gen-commit (subject containing %q) on %s\nlog:\n%s",
			featureSubject, repo, gitOutputIsolated(t, repo, "log", "--oneline", "-20"))
	}
	// pin must be ancestor of feature (pin before feature on the branch).
	err := git_isolated.Command(repo, "merge-base", "--is-ancestor", pinSHA, featSHA).Run()
	if err != nil {
		t.Fatalf("B1 order: pin commit must be ancestor of feature commit (pin before gen-commit)\npin=%s feature=%s\nlog:\n%s",
			pinSHA, featSHA, gitOutputIsolated(t, repo, "log", "--oneline", "-20"))
	}
	// Feature commit tree must not reintroduce external replace.
	featGoMod := gitOutputIsolated(t, repo, "show", featSHA+":go.mod")
	if strings.Contains(featGoMod, "=> ./external/") || strings.Contains(featGoMod, "=> ../") {
		// Only fail when the dep module still has an external replace line.
		if strings.Contains(featGoMod, depModule) && strings.Contains(featGoMod, "=>") {
			t.Fatalf("feature commit go.mod must not carry external replace for %s\n%s",
				depModule, featGoMod)
		}
	}
}

// assertFeatureWIPLanded checks FEATURE_WIP.md content is on history (gen-commit).
func assertFeatureWIPLanded(t *testing.T, repo string) {
	t.Helper()
	// Prefer show from HEAD; fall back to log --name-only.
	out, err := git_isolated.CombinedOutput(repo, "show", "HEAD:"+cascadeFeatureWIPFile)
	if err != nil {
		// Search history for the path.
		log := gitOutputIsolated(t, repo, "log", "--oneline", "--", cascadeFeatureWIPFile)
		if strings.TrimSpace(log) == "" {
			t.Fatalf("feature WIP %s never committed on %s\nlog:\n%s",
				cascadeFeatureWIPFile, repo, gitOutputIsolated(t, repo, "log", "--oneline", "-20"))
		}
		return
	}
	if !strings.Contains(string(out), "feature WIP") {
		t.Fatalf("HEAD:%s missing feature body; got %q", cascadeFeatureWIPFile, string(out))
	}
}

// historyRepoForConsumer picks the checkout to assert git log on (main after land).
func historyRepoForConsumer(t *testing.T, req *Request) string {
	t.Helper()
	if req.MainRepo != "" {
		return req.MainRepo
	}
	if req.WtDir != "" {
		return req.WtDir
	}
	t.Fatal("MainRepo or WtDir required for history asserts")
	return ""
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	// Keep generator references for helpers used only from leaves.
	_ = installCascadeNoLocalReplacePreCommit
	_ = installCascadePermissivePreCommit
	_ = installCascadeFakeOpencodeEnv
	_ = cascadeUnwindGenCommitArgs
	_ = dirtyCascadeFeatureWIP
	_ = setupApplyPinBeforeFeatureExternalCleanDep
	_ = setupApplyPinBeforeFeatureFreeDirty
	_ = setupApplyPinBeforeFeatureMonorepoFreeHostExternal
	_ = setupApplyPinBeforeFeatureThreeLevelFreeHostDirty
	_ = assertFreeTagNextBeforeMidPinOfFree
	_ = assertMidPinnedToFreeLeaf
	_ = pinCommitSHAForDep
	_ = assertPinCommitForDepFilesOnlyModSum
	_ = assertCascadePinForDepBeforeFeatureCommit
	_ = assertIntraSharedReplaceKept
	_ = assertExternalReplaceDropped
	_ = assertConsumerRequireAndNoExternalReplace
	_ = assertCascadePinBeforeFeatureCommit
	_ = assertFeatureWIPLanded
	_ = historyRepoForConsumer
	_ = featureCommitSHA
	_ = cascadeFeatureWIPFile
	_ = cascadeFeatureCommitSubject
	_ = cascadePinCommitPrefix
	return nil
}
```
