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
2. Leaves split on free-host shape / consumer dirt / replace form:
   - T1: pure multi-repo clean external replace-only
   - A-clean-tag: **clean free already at LatestTag** → no free next tag / no free commit
     (`external-clean-already-tagged-no-tag`; T1 seed + hard free-skip assert)
   - T2: free dirty external then consumer gen-commit
   - T-M1: **monorepo freeHost** (intra owned-changed) + clean external replace
   - T-tag1: **3-level freeHost** mid + dirty free (pin before free tag-next)
   - A-root-tag: **diamond A←B, A←C, C←B all dirty** → consumer root **must** tag-next
     (`diamond-all-dirty-consumer-must-tag`; hard assert A @ next at main HEAD)
   - A-wip-tag: **free dirty + consumer HEAD==LatestTag + uncommitted WIP only** →
     consumer **must** next tag at main HEAD
     (`consumer-at-latest-wip-must-tag`; crime-scene agent-pro hole)
   - T-spl / A1: **false freeHost** via noise intra pins + dirty free + replace-only
   - A2: false freeHost monorepo + **FEATURE_WIP** (pin then deferred feature)
   - A4: **clean** consumer porcelain + committed replace + dirty free
   - A5: **resume** free already landed clean/untagged + consumer replace dirt
   - D1: **absolute-path** external replace (droppable; pin drops it)
   - C1: **free multi-module** (root + `cmd/`) both tag-next; consumer pins **root only**
   - CS-repin: **free multi-module + uncommitted free feature** → deferred tag must re-pin consumer
     (`deferred-tag-repin-after-free-uncommitted-feature`; crime-scene go-pkgs@120 miss)
   - C1-sync: **free multi-module + `--merge-back --sync`** — free linked WT tracks
     free main after cascade pin (`free-multimodule-merge-back-sync-wt-tracks-main`)
   - P-empty: **mid dirty + leaf clean + root go.mod-only** with `--add-all`
     (`pin-only-consumer-empty-gen-commit-with-add-all`) — pinReady empties
     index; must soft-skip gen-commit + land (not hard-fail / diverge)

## Context

- **D2** B1 free-first interleaved apply; **D3** keep-current require when no
  free tag; **D7** separate pin auto-commit then feature gen-commit.
- Classic TDD: leaves are **RED** until product reorders pin before consumer
  feature gen-commit (today peels all then cascade; gen-commit hits hook while
  replace still present).
- **P2 coverage backfill (A2/A4/A5/D1):** P1 freeHost fix already in tree;
  new leaves may be GREEN immediately (backfill OK). Mixed GREEN/RED OK.
- **T-M1 hole:** same-label freeHost (intra free pin dep) blocks pure consumer
  deferral, so ready external pins never run before freeHost feature gen-commit.
- **T-tag1 hole:** mid freeHost peels early; `pinReadyExternalReplacesBeforeGenCommit`
  pins dirty free @ planned NextTag **before** cascade `tag-next` (missing
  `attachTagScopeToModules` → `cascadeModuleShouldTag` always false). Production
  surfaces as `go mod tidy: unknown revision`; L2 locks tag-before-pin order.
- **T-spl / A1 hole:** `splitPeelOrderB1` marks freeHost for **every** pin dep
  label, including noise intra pins @ LatestTag with **no** tag-next. Monorepo
  consumer peels early with unready external free replace → hook fail. freeHost
  must be true tag hosts only so pure pin-consumers defer.
- Do not rewrite sealed ASSERT contracts under `clean/`, `dirty-gomod/`,
  `partial-edit/`, `reinstall-local/`, or sealed T1/T2/T-M1/T-tag1/T-spl leaves.
- **P3 C1:** free monorepo multi-module tags (go-pkgs root + nested `cmd/`) +
  consumer require **root only** — pin must not force-add nested free/cmd.

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
	// Free multi-module nested cmd (C1 / go-pkgs shape): path-scoped tags.
	// Root module reuses unwindDotPkgsModule (example.com/dot-pkgs).
	freeMultiCmdModule  = "example.com/dot-pkgs/cmd"
	freeMultiCmdDir     = "cmd" // rel under free monorepo main
	freeMultiCmdOldTag  = "cmd/v0.0.1"
	freeMultiCmdNextTag = "cmd/v0.0.2"

	// Uncommitted feature WIP path for consumer gen-commit (not go.mod).
	cascadeFeatureWIPFile = "FEATURE_WIP.md"
	cascadeFeatureWIPBody = "feature WIP for gen-commit before pin ordering\n"

	// Mock gen-commit subject from fake-opencode (must match mock JSON).
	cascadeFeatureCommitSubject = "feat: add feature"

	// Fake-opencode mock: title = cascadeFeatureCommitSubject.
	cascadeFakeOpencodeMockJSON = `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"cascade_pin_before_feature","llm_events":[{"type":"step_start"},{"type":"message","text":"{\"title\": \"feat: add feature\", \"description\": \"Consumer feature after cascade pin\"}"},{"type":"step_finish"}]}`

	// Shell pre-commit: fail when go.mod still has a droppable external-style
	// local replace (./external/…, ../…, or absolute /…). Intra-repo ./pkgs/…
	// is allowed. Mirrors git-hook-go-no-local-replace lenient mode for stack
	// externals (D1 absolute-path coverage; relative-only leaves unchanged).
	cascadeNoLocalReplacePreCommit = `#!/bin/sh
# Fixture: simulate git-hook-go-no-local-replace for external stack replaces.
if [ ! -f go.mod ]; then
  exit 0
fi
if grep -E '=>[[:space:]]*(\.\./|\./external/|/)' go.mod >/dev/null 2>&1; then
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

// setupApplyPinOnlyConsumerEmptyGenCommitWithAddAll is P-empty fixture
// (production repro: spl go.mod-only + dirty agent-pro + clean dot-pkgs):
//
//	leaf (dot-pkgs): clean at release tag v0.0.1 (no peel)
//	mid (agent-pro): dirty FEATURE_WIP + optional replace→leaf; freeHost peel
//	  gen-commits feature then lands → next tag v0.0.2
//	root linked WT: **only uncommitted go.mod** with droppable external replaces
//	  → mid (+ leaf); **no** FEATURE_WIP — pinReady/cascade consume all dirt
//	modproxy old+next for mid/leaf; bare origins for --push
//
// Desired: exit 0; mid feature lands; root peels after pin (empty gen-commit
// soft-skip with --add-all); merge-back so root branch not diverged from main;
// cascade pin commits never scoop external replaces into pin subjects.
// Classic RED today: --add-all disables partial-edit (pin scoop) + empty
// gen-commit hard-fails → abort before root merge-back → Master diverged.
func setupApplyPinOnlyConsumerEmptyGenCommitWithAddAll(t *testing.T, req *Request) {
	t.Helper()

	// Root pins mid after mid tags next; LeafModulePath = mid module for require asserts.
	req.LeafModulePath = unwindAgentProModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- leaf free: clean @ v0.0.1 ---
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

	// --- mid freeHost: requires leaf; feature WIP peels early ---
	midMain := filepath.Join(req.WorkRoot, labelAgentPro)
	initGitRepoOnMain(t, midMain)
	writeGoModRequire(t, midMain, unwindAgentProModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(midMain, "agent.go"),
		"package agentpro\n\nimport _ \""+unwindDotPkgsModule+"\"\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	runGitIsolated(t, midMain, "add", "go.mod", "agent.go")
	runGitIsolated(t, midMain, "commit", "-m", "add agent-pro mid")
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

	// --- root pure pinConsumer main + origin ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	// Require mid + leaf (production-like dual require); baseline without replace.
	writeGoModRequire(t, rootMain, unwindRootModule,
		unwindAgentProModule+"@"+unwindApplyOldTag,
		unwindDotPkgsModule+"@"+unwindApplyOldTag,
	)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport (\n\t_ \""+unwindAgentProModule+"\"\n\t_ \""+unwindDotPkgsModule+"\"\n)\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root consumer")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	rootBare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, rootMain, rootBare)
	if tagRefExists(t, rootMain, unwindApplyOldTag) {
		runGitIsolated(t, rootMain, "push", "origin", unwindApplyOldTag)
	}
	writeFile(t, filepath.Join(req.WorkRoot, "root-origin.path"), rootBare+"\n")

	// Linked root = primary (gen-commit only for linked peels).
	wtDir := filepath.Join(req.WorkRoot, "root-linked")
	runGitIsolated(t, rootMain, "worktree", "add", "-b", branchNameMainDate(), wtDir)
	wtDir = resolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchNameMainDate()

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
	// Leaf stays clean (no markDirty / no post-tag commits).

	// Mid: committed replace → clean leaf + staged FEATURE_WIP (feature peel).
	appendLocalReplace(t, midExt, unwindDotPkgsModule, relLocalReplace(t, midExt, leafExt))
	runGitIsolated(t, midExt, "add", "go.mod")
	runGitIsolated(t, midExt, "commit", "-m", "mid replace to clean leaf")
	dirtyCascadeFeatureWIP(t, midExt)
	// Feature WIP alone is dirty; no extra DIRTY file required.
	installCascadePermissivePreCommit(t, midExt)

	// Root: ignore external/ committed; then **uncommitted** go.mod replaces only
	// (staged) — pin-only consumer, no FEATURE_WIP.
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", ".gitignore")
	runGitIsolated(t, wtDir, "commit", "-m", "ignore external stack members")

	relMid := filepath.ToSlash(filepath.Join("external", midExtName))
	relLeaf := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, wtDir, unwindAgentProModule, "./"+relMid)
	appendLocalReplace(t, wtDir, unwindDotPkgsModule, "./"+relLeaf)
	// Stage go.mod only (repro: "Changes to be committed: go.mod").
	runGitIsolated(t, wtDir, "add", "--", "go.mod")
	status := gitOutputIsolated(t, wtDir, "status", "--porcelain", "--", "go.mod")
	if strings.TrimSpace(status) == "" {
		t.Fatal("expected staged go.mod dirt on pin-only root consumer")
	}
	// Ensure no FEATURE_WIP on root (pin-only).
	if _, err := os.Stat(filepath.Join(wtDir, cascadeFeatureWIPFile)); err == nil {
		t.Fatal("pin-only root must not have FEATURE_WIP.md")
	}
	installCascadePermissivePreCommit(t, wtDir)

	// Offline proxy: leaf @ old only (clean free); mid old+next for root pin after mid tag.
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	oldLeafSeed := filepath.Join(req.WorkRoot, "seed-leaf-"+unwindApplyOldTag)
	mkdirAll(t, oldLeafSeed)
	writeGoModRequire(t, oldLeafSeed, unwindDotPkgsModule)
	writeFile(t, filepath.Join(oldLeafSeed, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyOldTag, oldLeafSeed)

	seedFileModuleProxy(t, proxyRoot, unwindAgentProModule, unwindApplyOldTag, midMain)
	midNextSeed := filepath.Join(req.WorkRoot, "seed-mid-"+unwindApplyNextTag)
	mkdirAll(t, midNextSeed)
	writeGoModRequire(t, midNextSeed, unwindAgentProModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(midNextSeed, "agent.go"),
		"package agentpro\n\nimport _ \""+unwindDotPkgsModule+"\"\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	writeFile(t, filepath.Join(midNextSeed, cascadeFeatureWIPFile), cascadeFeatureWIPBody)
	seedFileModuleProxy(t, proxyRoot, unwindAgentProModule, unwindApplyNextTag, midNextSeed)
	enableFileModuleProxy(t, req, proxyRoot)

	req.RepoDir = wtDir
	// Peel dirty only: mid + root (leaf clean).
	setPeelOrderDisplays(t, req, midExt, wtDir)
}

// setupApplyPinBeforeFeatureDiamondAllDirtyConsumerTag is A-root-tag fixture:
//
//	Diamond multi-repo stack (production-like A/external/B + A/external/C):
//	  B leaf free (dot-pkgs): dirty owned-changed → next v0.0.2; bare origin
//	  C mid freeHost (agent-pro): requires B; external replace → B; owned-changed
//	    → next v0.0.2; FEATURE_WIP for early freeHost peel
//	  A root consumer: requires B and C; external replaces → both; owned-changed
//	    → next v0.0.2; FEATURE_WIP deferred peel after cascade pin
//	modproxy: old+next for B and C (network pin after drop replace)
//
// Coverage gap: B1 multi-repo leaves lock free tags + consumer pins but never
// assert consumer/root A receives tag-next at main HEAD after full recipe.
// Expected GREEN: B, C, and A all tagged @ v0.0.2; A tag at main HEAD.
func setupApplyPinBeforeFeatureDiamondAllDirtyConsumerTag(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- B leaf free main + origin (owned-changed → next tag) ---
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

	// --- C mid freeHost main + origin (requires free @ baseline) ---
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

	// --- A root consumer main + origin (requires B and C @ baseline) ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule,
		unwindAgentProModule+"@"+unwindApplyOldTag,
		unwindDotPkgsModule+"@"+unwindApplyOldTag,
	)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport (\n\t_ \""+unwindAgentProModule+"\"\n\t_ \""+unwindDotPkgsModule+"\"\n)\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root diamond consumer")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	rootBare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, rootMain, rootBare)
	if tagRefExists(t, rootMain, unwindApplyOldTag) {
		runGitIsolated(t, rootMain, "push", "origin", unwindApplyOldTag)
	}
	writeFile(t, filepath.Join(req.WorkRoot, "root-origin.path"), rootBare+"\n")

	// Linked A = stack primary (cwd for inventory / deferred peel gen-commit).
	wtDir := filepath.Join(req.WorkRoot, "root-linked")
	runGitIsolated(t, rootMain, "worktree", "add", "-b", branchNameMainDate(), wtDir)
	wtDir = resolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchNameMainDate()

	// Nest C + B linked externals under A (sibling externals, production-like).
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

	// B dirty owned-changed after baseline tag → next v0.0.2.
	writeFile(t, filepath.Join(leafExt, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, leafExt, "add", "pkg.go")
	runGitIsolated(t, leafExt, "commit", "-m", "leaf B feature for next tag")
	markDirty(t, leafExt)

	// C mid freeHost: droppable external replace → B + owned-changed + FEATURE_WIP.
	appendLocalReplace(t, midExt, unwindDotPkgsModule, relLocalReplace(t, midExt, leafExt))
	writeFile(t, filepath.Join(midExt, "agent.go"),
		"package agentpro\n\nimport _ \""+unwindDotPkgsModule+"\"\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, midExt, "add", "go.mod", "agent.go")
	runGitIsolated(t, midExt, "commit", "-m", "mid C replace to B + owned change")
	dirtyCascadeFeatureWIP(t, midExt)
	markDirty(t, midExt)
	installCascadePermissivePreCommit(t, midExt)

	// A root: droppable replaces → B and C; owned-changed → next; FEATURE_WIP deferred.
	relMid := filepath.ToSlash(filepath.Join("external", midExtName))
	relLeaf := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, wtDir, unwindAgentProModule, "./"+relMid)
	appendLocalReplace(t, wtDir, unwindDotPkgsModule, "./"+relLeaf)
	writeFile(t, filepath.Join(wtDir, "root.go"),
		"package root\n\nimport (\n\t_ \""+unwindAgentProModule+"\"\n\t_ \""+unwindDotPkgsModule+"\"\n)\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", "go.mod", "root.go", ".gitignore")
	runGitIsolated(t, wtDir, "commit", "-m", "root A diamond replaces + owned change")
	dirtyCascadeFeatureWIP(t, wtDir)
	markDirty(t, wtDir)
	// Root hook: feature gen-commit must not see external replaces (pin first).
	installCascadeNoLocalReplacePreCommit(t, wtDir)

	// Offline proxy: B old+next, C old+next for pin+tidy after drop replace.
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyNextTag, leafExt)
	oldLeafSeed := filepath.Join(req.WorkRoot, "seed-leaf-"+unwindApplyOldTag)
	mkdirAll(t, oldLeafSeed)
	writeGoModRequire(t, oldLeafSeed, unwindDotPkgsModule)
	writeFile(t, filepath.Join(oldLeafSeed, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyOldTag, oldLeafSeed)

	seedFileModuleProxy(t, proxyRoot, unwindAgentProModule, unwindApplyOldTag, midMain)
	midNextSeed := filepath.Join(req.WorkRoot, "seed-mid-"+unwindApplyNextTag)
	mkdirAll(t, midNextSeed)
	writeGoModRequire(t, midNextSeed, unwindAgentProModule, unwindDotPkgsModule+"@"+unwindApplyNextTag)
	writeFile(t, filepath.Join(midNextSeed, "agent.go"),
		"package agentpro\n\nimport _ \""+unwindDotPkgsModule+"\"\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindAgentProModule, unwindApplyNextTag, midNextSeed)
	// Mid next zip may require free@next during tidy of root.
	enableFileModuleProxy(t, req, proxyRoot)

	req.RepoDir = wtDir
	setPeelOrderDisplays(t, req, leafExt, midExt, wtDir)
}

// setupApplyPinBeforeFeatureFreeMultiModuleCmdRootOnly is C1 fixture:
//
//	free monorepo (go-pkgs shape) under external:
//	  example.com/dot-pkgs (root) owned-changed → tag-next v0.0.2
//	  example.com/dot-pkgs/cmd under cmd/ requires free @ v0.0.1 + keep-local
//	    replace => ../ ; owned-changed → tag-next cmd/v0.0.2 after pin free
//	  baseline tags: v0.0.1 + cmd/v0.0.1; bare origin for --push
//	consumer main (primary RepoDir):
//	  require free root only @ v0.0.1 + droppable replace => ./external/…
//	  never requires free/cmd
//	modproxy: free root old+next only (no free/cmd next — spurious pin fails tidy)
//
// Free-first cascade on free monorepo: tag free root → pin free/cmd ← free →
// tag free/cmd; then pin consumer ← free root only (not nested).
// Expected consumer pin version = v0.0.2.
func setupApplyPinBeforeFeatureFreeMultiModuleCmdRootOnly(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.NestedModulePath = freeMultiCmdModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- free monorepo main: root + cmd ---
	leafMain := filepath.Join(req.WorkRoot, labelDotPkgs)
	initGitRepoOnMain(t, leafMain)
	writeGoModRequire(t, leafMain, unwindDotPkgsModule)
	writeFile(t, filepath.Join(leafMain, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	cmdDir := filepath.Join(leafMain, freeMultiCmdDir)
	mkdirAll(t, cmdDir)
	// Nested cmd requires free root; keep-local replace to parent (intra monorepo).
	// Library package (not package main) so SETUP fixtures avoid package-main anti-pattern.
	writeGoModRequire(t, cmdDir, freeMultiCmdModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	appendLocalReplace(t, cmdDir, unwindDotPkgsModule, "../")
	writeFile(t, filepath.Join(cmdDir, "cmd.go"),
		"package freecmd\n\nimport _ \""+unwindDotPkgsModule+"\"\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	runGitIsolated(t, leafMain, "add", "go.mod", "pkg.go",
		filepath.Join(freeMultiCmdDir, "go.mod"),
		filepath.Join(freeMultiCmdDir, "cmd.go"))
	runGitIsolated(t, leafMain, "commit", "-m", "add free monorepo root + cmd modules")
	createLightweightTag(t, leafMain, unwindApplyOldTag, "")
	createLightweightTag(t, leafMain, freeMultiCmdOldTag, "")
	leafMain = resolvePath(t, leafMain)
	req.SecondRepo = leafMain
	installCascadePermissivePreCommit(t, leafMain)

	leafBare := setupBareOrigin(t, req.WorkRoot, "leaf-origin")
	attachOriginAndPushMain(t, leafMain, leafBare)
	if tagRefExists(t, leafMain, unwindApplyOldTag) {
		runGitIsolated(t, leafMain, "push", "origin", unwindApplyOldTag)
	}
	if tagRefExists(t, leafMain, freeMultiCmdOldTag) {
		runGitIsolated(t, leafMain, "push", "origin", freeMultiCmdOldTag)
	}
	req.OriginBare = leafBare

	// --- consumer main: require free root only ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root consumer require free root only")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	// Consumer bare origin so cascade push after root tag-next succeeds.
	rootBare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, rootMain, rootBare)
	if tagRefExists(t, rootMain, unwindApplyOldTag) {
		runGitIsolated(t, rootMain, "push", "origin", unwindApplyOldTag)
	}
	writeFile(t, filepath.Join(req.WorkRoot, "root-origin.path"), rootBare+"\n")

	// Nest free multi-module WT under consumer/external (stack member).
	extDir := filepath.Join(rootMain, "external")
	mkdirAll(t, extDir)
	leafExtName := labelDotPkgs + "-" + branchNameMainDate()
	leafExt := filepath.Join(extDir, leafExtName)
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate(), leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt
	req.WtBranch = branchNameMainDate()

	// Owned change on free root + free/cmd (both need tag-next).
	writeFile(t, filepath.Join(leafExt, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	writeFile(t, filepath.Join(leafExt, freeMultiCmdDir, "cmd.go"),
		"package freecmd\n\nimport _ \""+unwindDotPkgsModule+"\"\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, leafExt, "add", "pkg.go", filepath.Join(freeMultiCmdDir, "cmd.go"))
	runGitIsolated(t, leafExt, "commit", "-m", "free root + cmd owned change for next tags")
	markDirty(t, leafExt)

	// Droppable external replace free root only (consumer never requires free/cmd).
	relReplace := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, rootMain, unwindDotPkgsModule, "./"+relReplace)
	writeFile(t, filepath.Join(rootMain, ".gitignore"), "/external\n")
	runGitIsolated(t, rootMain, "add", "go.mod", ".gitignore")
	runGitIsolated(t, rootMain, "commit", "-m", "external free replace + ignore")
	// Consumer not required dirty for pin (clean Base + replace committed).
	// Free dirty drives peel; cascade pins consumer require root only.

	// Offline proxy: free root old+next only. Do NOT seed free/cmd@next — a
	// cartesian pin that force-adds nested require would fail tidy (RED surface).
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	// Single-module seeds (avoid multi-module zip pollution from cmd/).
	oldSeed := filepath.Join(req.WorkRoot, "seed-free-root-"+unwindApplyOldTag)
	mkdirAll(t, oldSeed)
	writeGoModRequire(t, oldSeed, unwindDotPkgsModule)
	writeFile(t, filepath.Join(oldSeed, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyOldTag, oldSeed)

	nextSeed := filepath.Join(req.WorkRoot, "seed-free-root-"+unwindApplyNextTag)
	mkdirAll(t, nextSeed)
	writeGoModRequire(t, nextSeed, unwindDotPkgsModule)
	writeFile(t, filepath.Join(nextSeed, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyNextTag, nextSeed)

	// Nested cmd baseline only (if anything wrongly pins nested@old, still OK).
	cmdOldSeed := filepath.Join(req.WorkRoot, "seed-free-cmd-"+unwindApplyOldTag)
	mkdirAll(t, cmdOldSeed)
	writeGoModRequire(t, cmdOldSeed, freeMultiCmdModule)
	writeFile(t, filepath.Join(cmdOldSeed, "cmd.go"),
		"package freecmd\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, freeMultiCmdModule, unwindApplyOldTag, cmdOldSeed)
	enableFileModuleProxy(t, req, proxyRoot)

	req.RepoDir = rootMain
	setPeelOrderDisplays(t, req, leafExt)
}


// setupApplyDeferredTagRepinAfterFreeUncommittedFeature is the crime-scene
// formalization of remote-agent bash / wrk cascade miss (2026-08-12):
//
//	free monorepo (go-pkgs shape) under external linked WT:
//	  example.com/dot-pkgs + example.com/dot-pkgs/cmd (cmd requires free @ v0.0.1
//	    + keep-local replace => ../)
//	  baseline tags v0.0.1 + cmd/v0.0.1
//	  **uncommitted** free root owned change on WT (not committed → tagscope
//	    NextTag empty at cascade plan time; gen-commit lands it during peel)
//	consumer primary (clean Base + committed droppable replace to free WT):
//	  require free root @ v0.0.1 only
//
// Hole: free/cmd requires free @ v0.0.0 (drift vs LatestTag v0.0.1) so cascade
// pins free/cmd without free TagNext (uncommitted free WIP ⇒ empty NextTag).
// That marks free monorepo pinConsumer without freeHost → free peel DEFERRED →
// auto-commit + applyDeferredCascadeTags create free @ v0.0.2 → **no consumer
// re-pin** to v0.0.2 (applyDeferredCascadeTags is tag-only).
// Desired: after full apply, consumer require free @ v0.0.2 (and free tagged).
func setupApplyDeferredTagRepinAfterFreeUncommittedFeature(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.NestedModulePath = freeMultiCmdModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- free monorepo main: root + cmd at baseline tags ---
	leafMain := filepath.Join(req.WorkRoot, labelDotPkgs)
	initGitRepoOnMain(t, leafMain)
	writeGoModRequire(t, leafMain, unwindDotPkgsModule)
	writeFile(t, filepath.Join(leafMain, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	cmdDir := filepath.Join(leafMain, freeMultiCmdDir)
	mkdirAll(t, cmdDir)
	writeGoModRequire(t, cmdDir, freeMultiCmdModule, unwindDotPkgsModule+"@v0.0.0") // stale require → drift pin free/cmd without free TagNext
	appendLocalReplace(t, cmdDir, unwindDotPkgsModule, "../")
	writeFile(t, filepath.Join(cmdDir, "cmd.go"),
		"package freecmd\n\nimport _ \""+unwindDotPkgsModule+"\"\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	runGitIsolated(t, leafMain, "add", "go.mod", "pkg.go",
		filepath.Join(freeMultiCmdDir, "go.mod"),
		filepath.Join(freeMultiCmdDir, "cmd.go"))
	runGitIsolated(t, leafMain, "commit", "-m", "add free monorepo root + cmd modules")
	createLightweightTag(t, leafMain, unwindApplyOldTag, "")
	createLightweightTag(t, leafMain, freeMultiCmdOldTag, "")
	leafMain = resolvePath(t, leafMain)
	req.SecondRepo = leafMain
	// Free gen-commit must not inherit developer hooks; permissive only.
	installCascadePermissivePreCommit(t, leafMain)

	leafBare := setupBareOrigin(t, req.WorkRoot, "leaf-origin")
	attachOriginAndPushMain(t, leafMain, leafBare)
	if tagRefExists(t, leafMain, unwindApplyOldTag) {
		runGitIsolated(t, leafMain, "push", "origin", unwindApplyOldTag)
	}
	if tagRefExists(t, leafMain, freeMultiCmdOldTag) {
		runGitIsolated(t, leafMain, "push", "origin", freeMultiCmdOldTag)
	}
	req.OriginBare = leafBare

	// --- consumer main: require free root only @ latest baseline ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root consumer require free root only")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	rootBare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, rootMain, rootBare)
	if tagRefExists(t, rootMain, unwindApplyOldTag) {
		runGitIsolated(t, rootMain, "push", "origin", unwindApplyOldTag)
	}
	writeFile(t, filepath.Join(req.WorkRoot, "root-origin.path"), rootBare+"\n")

	// Nest free multi-module WT under consumer/external (stack member).
	extDir := filepath.Join(rootMain, "external")
	mkdirAll(t, extDir)
	leafExtName := labelDotPkgs + "-" + branchNameMainDate()
	leafExt := filepath.Join(extDir, leafExtName)
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate(), leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt
	req.WtBranch = branchNameMainDate()

	// Crime-scene key: free root owned change is **uncommitted** on the WT.
	// tagscope.Plan(HEAD) still sees LatestTag only → empty NextTag at cascade plan.
	// --add-all gen-commit during free peel must land this content then tag next.
	writeFile(t, filepath.Join(leafExt, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	// Leave unstaged/uncommitted; markDirty so peel inventory includes free.
	markDirty(t, leafExt)

	// Droppable external replace free root only (consumer never requires free/cmd).
	relReplace := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, rootMain, unwindDotPkgsModule, "./"+relReplace)
	writeFile(t, filepath.Join(rootMain, ".gitignore"), "/external\n")
	runGitIsolated(t, rootMain, "add", "go.mod", ".gitignore")
	runGitIsolated(t, rootMain, "commit", "-m", "external free replace + ignore")
	// Consumer clean Base after replace commit — pin alone owns require bump.

	// Offline proxy: free root old+next (next needed when pin correctly repins).
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	oldSeed := filepath.Join(req.WorkRoot, "seed-free-root-"+unwindApplyOldTag)
	mkdirAll(t, oldSeed)
	writeGoModRequire(t, oldSeed, unwindDotPkgsModule)
	writeFile(t, filepath.Join(oldSeed, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyOldTag, oldSeed)

	nextSeed := filepath.Join(req.WorkRoot, "seed-free-root-"+unwindApplyNextTag)
	mkdirAll(t, nextSeed)
	writeGoModRequire(t, nextSeed, unwindDotPkgsModule)
	writeFile(t, filepath.Join(nextSeed, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyNextTag, nextSeed)

	cmdOldSeed := filepath.Join(req.WorkRoot, "seed-free-cmd-"+unwindApplyOldTag)
	mkdirAll(t, cmdOldSeed)
	writeGoModRequire(t, cmdOldSeed, freeMultiCmdModule)
	writeFile(t, filepath.Join(cmdOldSeed, "cmd.go"),
		"package freecmd\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, freeMultiCmdModule, unwindApplyOldTag, cmdOldSeed)
	enableFileModuleProxy(t, req, proxyRoot)

	req.RepoDir = rootMain
	setPeelOrderDisplays(t, req, leafExt)
}

// setupApplyPinBeforeFeatureFalseFreeHostIntraPins is T-spl / A1 fixture:
//
//	external free (dot-pkgs): dirty owned-changed → next tag v0.0.2; bare origin + --push
//	consumer monorepo (linked WT primary):
//	  - root + pkgs/shared (no owned-change on shared after pkgs/shared/v0.0.1)
//	  - root requires shared with **pre-pin drift** vs LatestTag so cascade plans
//	    noise pin root←shared @ LatestTag **without** tag-next on shared
//	  - root requires free @ v0.0.1 + droppable replace => ./external/…
//	  - consumer dirt = go.mod/go.sum (committed replace) + DIRTY only — **no** FEATURE_WIP
//	pre-commit: no external local replace (hook sim; ./pkgs/… allowed)
//	modproxy: old + next for free pin after tag
//
// Bug: splitPeelOrderB1 freeHost[depLabel] for every pin dep — noise intra pin
// marks monorepo freeHost → early peel with unready free replace → hook fail.
// Correct: freeHost only true tag hosts; monorepo pure pin-consumer defers.
// Expected free pin version = v0.0.2.
func setupApplyPinBeforeFeatureFalseFreeHostIntraPins(t *testing.T, req *Request) {
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

	// --- monorepo main: root + pkgs/shared (no owned-change) + require free ---
	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)

	sharedDir := filepath.Join(rootMain, filepath.FromSlash(cascadeSharedDir))
	mkdirAll(t, sharedDir)
	writeGoModRequire(t, sharedDir, cascadeSharedModule)
	writeFile(t, filepath.Join(sharedDir, "shared.go"),
		"package shared\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	// Require shared at a version that drifts from LatestTag (pkgs/shared/v0.0.1 →
	// v0.0.1) so cascade plans noise pin root←shared @ LatestTag without tag-next
	// on shared (A1 false freeHost trigger). Free require matches baseline tag.
	writeGoModRequire(t, rootMain, unwindRootModule,
		cascadeSharedModule+"@v0.0.0",
		unwindDotPkgsModule+"@"+unwindApplyOldTag,
	)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport (\n\t_ \""+cascadeSharedModule+"\"\n\t_ \""+unwindDotPkgsModule+"\"\n)\n")
	// Intra keep-local replace (must survive; hook allows ./pkgs/…).
	appendLocalReplace(t, rootMain, cascadeSharedModule, "./"+cascadeSharedDir)

	runGitIsolated(t, rootMain, "add", "go.mod", "root.go", "pkgs")
	runGitIsolated(t, rootMain, "commit", "-m", "root + shared + require free (noise shared drift)")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "HEAD")
	createLightweightTag(t, rootMain, cascadeSharedOldTag, "HEAD")
	// No post-tag owned change on shared → no cascade tag-next for shared.

	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	rootBare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, rootMain, rootBare)
	if tagRefExists(t, rootMain, unwindApplyOldTag) {
		runGitIsolated(t, rootMain, "push", "origin", unwindApplyOldTag)
	}
	writeFile(t, filepath.Join(req.WorkRoot, "root-origin.path"), rootBare+"\n")

	// Linked monorepo = stack primary (gen-commit only runs for linked peels).
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

	// Committed droppable external replace; consumer dirt = DIRTY only (replace-only;
	// no FEATURE_WIP — gen-commit must not be the path that drops replace).
	relReplace := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, wtDir, unwindDotPkgsModule, "./"+relReplace)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", "go.mod", ".gitignore")
	runGitIsolated(t, wtDir, "commit", "-m", "external stack replace + ignore")

	// Replace-only consumer dirt: porcelain DIRTY, no feature WIP file.
	markDirty(t, wtDir)
	// Consumer hook only (after seed commits).
	installCascadeNoLocalReplacePreCommit(t, wtDir)

	// Offline proxy: next from leaf WT content; old synthetic.
	seedDotPkgsProxyVersions(t, req, map[string]string{
		unwindApplyNextTag: leafExt,
		unwindApplyOldTag:  "",
	})

	req.RepoDir = wtDir
	setPeelOrderDisplays(t, req, leafExt, wtDir)
}

// setupApplyPinBeforeFeatureFalseFreeHostIntraPinsWithFeatureWIP is A2:
// same monorepo false freeHost stack as T-spl/A1 plus staged FEATURE_WIP so
// deferred consumer peels run feature gen-commit after cascade pin (D7).
func setupApplyPinBeforeFeatureFalseFreeHostIntraPinsWithFeatureWIP(t *testing.T, req *Request) {
	t.Helper()
	setupApplyPinBeforeFeatureFalseFreeHostIntraPins(t, req)
	if req.WtDir == "" {
		t.Fatal("WtDir required after false freeHost seed")
	}
	dirtyCascadeFeatureWIP(t, req.WtDir)
}

// setupApplyPinBeforeFeatureCleanConsumerCommittedReplace is A4:
//
//	leaf external: dirty owned-changed → next v0.0.2; bare origin + --push
//	root linked WT: committed droppable external replace; **clean porcelain**
//	  (no DIRTY, no FEATURE_WIP) — consumer not in dirty peel order
//	modproxy: old + next for pin after free tag
//
// Expect: peel free only; cascade pin on consumer drops replace; no consumer
// gen-commit failure; exit 0; final require free @ next.
func setupApplyPinBeforeFeatureCleanConsumerCommittedReplace(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- leaf free main + origin ---
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

	// Committed replace; consumer working tree left **clean** (no DIRTY / WIP).
	relReplace := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, wtDir, unwindDotPkgsModule, "./"+relReplace)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", "go.mod", ".gitignore")
	runGitIsolated(t, wtDir, "commit", "-m", "stack replace + ignore external")
	// A4: do not markDirty / FEATURE_WIP — clean porcelain; peel free only.
	status := strings.TrimSpace(gitOutputIsolated(t, wtDir, "status", "--porcelain"))
	if status != "" {
		t.Fatalf("A4: consumer must be clean porcelain after seed; got:\n%s", status)
	}
	installCascadeNoLocalReplacePreCommit(t, wtDir)

	seedDotPkgsProxyVersions(t, req, map[string]string{
		unwindApplyNextTag: leafExt,
		unwindApplyOldTag:  "",
	})

	req.RepoDir = wtDir
	// Only free is dirty; consumer omitted from peel order intentionally.
	setPeelOrderDisplays(t, req, leafExt)
}

// setupApplyPinBeforeFeatureResumeFreeLandedUntagged is A5:
//
//	Simulate partial success after free peel/land: free is **already clean on
//	main** at the feature commit (owned-changed past LatestTag) but **untagged**.
//	Consumer still dirty with droppable external replace (replace-only dirt).
//
//	Re-run full apply flags → cascade tags free @ next, pins consumer (drop
//	replace), finishes; exit 0. No double wrong tags.
func setupApplyPinBeforeFeatureResumeFreeLandedUntagged(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- free main: baseline tag + feature commit already landed; clean ---
	leafMain := filepath.Join(req.WorkRoot, labelDotPkgs)
	initGitRepoOnMain(t, leafMain)
	writeGoModRequire(t, leafMain, unwindDotPkgsModule)
	writeFile(t, filepath.Join(leafMain, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	runGitIsolated(t, leafMain, "add", "go.mod", "pkg.go")
	runGitIsolated(t, leafMain, "commit", "-m", "add dot-pkgs module")
	createLightweightTag(t, leafMain, unwindApplyOldTag, "")
	// Feature already landed on main (resume after free peel success) — untagged next.
	writeFile(t, filepath.Join(leafMain, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, leafMain, "add", "pkg.go")
	runGitIsolated(t, leafMain, "commit", "-m", "leaf feature already landed (untagged)")
	leafMain = resolvePath(t, leafMain)
	req.SecondRepo = leafMain
	installCascadePermissivePreCommit(t, leafMain)
	// Clean porcelain on free (no DIRTY) — not in dirty peel order.
	if st := strings.TrimSpace(gitOutputIsolated(t, leafMain, "status", "--porcelain")); st != "" {
		t.Fatalf("A5: free main must be clean after land sim; got:\n%s", st)
	}

	leafBare := setupBareOrigin(t, req.WorkRoot, "leaf-origin")
	attachOriginAndPushMain(t, leafMain, leafBare)
	if tagRefExists(t, leafMain, unwindApplyOldTag) {
		runGitIsolated(t, leafMain, "push", "origin", unwindApplyOldTag)
	}
	// Push main tip (feature commit) so origin matches; no next tag yet.
	runGitIsolated(t, leafMain, "push", "origin", "main")
	req.OriginBare = leafBare

	// --- consumer main + linked WT with droppable replace → free main ---
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

	// Replace points at free main (already landed checkout), not a dirty linked WT.
	// Relative sibling form (../dot-pkgs) — distinct from D1 absolute-path leaf.
	relFree := relLocalReplace(t, wtDir, leafMain)
	appendLocalReplace(t, wtDir, unwindDotPkgsModule, relFree)
	runGitIsolated(t, wtDir, "add", "go.mod")
	runGitIsolated(t, wtDir, "commit", "-m", "external replace to free main (resume)")
	// Consumer still dirty (DIRTY only; replace committed).
	markDirty(t, wtDir)
	installCascadeNoLocalReplacePreCommit(t, wtDir)

	// Proxy: next content from free main tip; old synthetic.
	seedDotPkgsProxyVersions(t, req, map[string]string{
		unwindApplyNextTag: leafMain,
		unwindApplyOldTag:  "",
	})

	req.RepoDir = wtDir
	// Free clean → not peeled; only consumer dirty (deferred pure pin-consumer).
	setPeelOrderDisplays(t, req, wtDir)
}

// setupApplyPinBeforeFeatureFreeDirtyConsumerAtLatestWIP is A-wip-tag fixture:
//
//	leaf external: dirty / owned-changed → next tag v0.0.2; bare origin for --push
//	root main + linked WT: baseline tag v0.0.1 at HEAD (no commits past tag)
//	  + uncommitted droppable replace => ./external/…
//	  + uncommitted FEATURE_WIP.md for gen-commit
//	  + no-local-replace hook + modproxy old+next
//
// Distinct from T2: T2 commits replace past LatestTag so cascade tagscope sees
// owned-changed; this seed keeps consumer HEAD == LatestTag with porcelain only
// so cascade NextTag is empty today and deferred applyDeferredCascadeTags is a
// no-op after feature land (crime scene: agent-pro missing next tag).
// Expected: free @ v0.0.2; consumer root @ v0.0.2 at main HEAD after full B1.
func setupApplyPinBeforeFeatureFreeDirtyConsumerAtLatestWIP(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- leaf free main + origin ---
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

	// --- root consumer main + origin (tagged at baseline; no post-tag commits) ---
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

	// Dirty free under external (owned-changed for next tag after early peel).
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

	// Consumer: uncommitted replace + FEATURE_WIP only — HEAD stays at LatestTag.
	// (T2 commits replace past tag; that lets cascade plan consumer NextTag.)
	relReplace := filepath.ToSlash(filepath.Join("external", leafExtName))
	appendLocalReplace(t, wtDir, unwindDotPkgsModule, "./"+relReplace)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	// Do NOT git add/commit go.mod or .gitignore — keep HEAD == unwindApplyOldTag.
	dirtyCascadeFeatureWIP(t, wtDir)
	markDirty(t, wtDir)

	// Seed invariant: linked consumer HEAD must equal LatestTag (crime scene).
	head := revParseHEAD(t, wtDir)
	tagSHA := revParseRef(t, wtDir, "refs/tags/"+unwindApplyOldTag)
	if head != tagSHA {
		t.Fatalf("A-wip-tag seed: consumer HEAD %s must equal LatestTag %s %s",
			head, unwindApplyOldTag, tagSHA)
	}
	// Porcelain must show uncommitted replace/WIP (otherwise fixture is wrong).
	st := gitOutputIsolated(t, wtDir, "status", "--porcelain")
	if strings.TrimSpace(st) == "" {
		t.Fatal("A-wip-tag seed: expected dirty consumer porcelain")
	}

	installCascadeNoLocalReplacePreCommit(t, wtDir)

	seedDotPkgsProxyVersions(t, req, map[string]string{
		unwindApplyNextTag: leafExt,
		unwindApplyOldTag:  "",
	})
	// Consumer next zip for post-fix tag consumers / tidy if product self-pins.
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedConsNext := filepath.Join(req.WorkRoot, "seed-root-"+unwindApplyNextTag)
	mkdirAll(t, seedConsNext)
	writeGoModRequire(t, seedConsNext, unwindRootModule, unwindDotPkgsModule+"@"+unwindApplyNextTag)
	writeFile(t, filepath.Join(seedConsNext, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n")
	seedFileModuleProxy(t, proxyRoot, unwindRootModule, unwindApplyNextTag, seedConsNext)
	seedConsOld := filepath.Join(req.WorkRoot, "seed-root-"+unwindApplyOldTag)
	mkdirAll(t, seedConsOld)
	writeGoModRequire(t, seedConsOld, unwindRootModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(seedConsOld, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n")
	seedFileModuleProxy(t, proxyRoot, unwindRootModule, unwindApplyOldTag, seedConsOld)

	req.RepoDir = wtDir
	setPeelOrderDisplays(t, req, leafExt, wtDir)
}

// setupApplyPinBeforeFeatureAbsolutePathReplace is D1:
//
//	Same free-dirty + consumer FEATURE_WIP shape as T2, but replace NewPath is
//	an **absolute** path to the free external checkout (production wrk style).
//	Droppable; cascade pin must drop it; hook sim forbids absolute external too.
func setupApplyPinBeforeFeatureAbsolutePathReplace(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	// --- leaf free main + origin ---
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

	// Dirty free under external (same as T2 layout).
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

	// Absolute-path replace (D1) — production-style, not ./external/….
	absReplace := filepath.ToSlash(leafExt)
	if !filepath.IsAbs(leafExt) {
		t.Fatalf("D1: leafExt must be absolute for abs replace seed; got %q", leafExt)
	}
	appendLocalReplace(t, wtDir, unwindDotPkgsModule, absReplace)
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n")
	runGitIsolated(t, wtDir, "add", "go.mod", ".gitignore")
	runGitIsolated(t, wtDir, "commit", "-m", "absolute path stack replace + ignore external")

	// Confirm seed go.mod carries absolute replace NewPath.
	seedMod := readFile(t, filepath.Join(wtDir, "go.mod"))
	if !strings.Contains(seedMod, "=> "+absReplace) && !strings.Contains(seedMod, "=>\t"+absReplace) {
		// Tolerant: slash form already in absReplace; also accept space variants.
		if !strings.Contains(seedMod, absReplace) {
			t.Fatalf("D1 seed go.mod missing absolute replace path %q:\n%s", absReplace, seedMod)
		}
	}

	dirtyCascadeFeatureWIP(t, wtDir)
	markDirty(t, wtDir)
	// Shared hook now also forbids absolute external NewPath (D1).
	installCascadeNoLocalReplacePreCommit(t, wtDir)

	seedDotPkgsProxyVersions(t, req, map[string]string{
		unwindApplyNextTag: leafExt,
		unwindApplyOldTag:  "",
	})

	req.RepoDir = wtDir
	setPeelOrderDisplays(t, req, leafExt, wtDir)
}

// assertFreeTagNextBeforeConsumerPinOfFree locks free tag-next before consumer
// pin of free @ next (T-spl / T2 order). free must be tagged before pin drops
// replace onto the next version.
func assertFreeTagNextBeforeConsumerPinOfFree(t *testing.T, out string) {
	t.Helper()
	tagNeedle := "tag-next " + unwindDotPkgsModule + " @ " + unwindApplyNextTag
	// pin log uses stack repo labels (not module paths).
	pinNeedle := "pin " + labelRoot + " <- " + labelDotPkgs + " @ " + unwindApplyNextTag
	tagIdx := strings.Index(out, tagNeedle)
	pinIdx := strings.Index(out, pinNeedle)
	if pinIdx < 0 {
		alt := "<- " + labelDotPkgs + " @ " + unwindApplyNextTag
		pinIdx = strings.Index(out, alt)
		if pinIdx >= 0 {
			pinNeedle = alt
		}
	}
	if tagIdx < 0 {
		t.Fatalf("T-spl: missing free tag-next line %q (free must be tagged before consumer pin)\nout:\n%s",
			tagNeedle, out)
	}
	if pinIdx < 0 {
		t.Fatalf("T-spl: missing consumer pin of free @ next (want %q or pin … <- %s @ %s)\nout:\n%s",
			pinNeedle, labelDotPkgs, unwindApplyNextTag, out)
	}
	if pinIdx < tagIdx {
		t.Fatalf("T-spl: free tag-next must precede consumer pin of free @ next\ntag@%d pin@%d\nneedles: %q then %q\nout:\n%s",
			tagIdx, pinIdx, tagNeedle, pinNeedle, out)
	}
}

// assertNoLocalReplaceGenCommitFail fails when combined output shows the fixture
// pre-commit / gen-commit blocked on local external replace (T-spl RED surface).
func assertNoLocalReplaceGenCommitFail(t *testing.T, out string) {
	t.Helper()
	low := strings.ToLower(out)
	if strings.Contains(low, "local external replace forbidden") ||
		strings.Contains(low, "git-hook-go-no-local-replace") {
		t.Fatalf("T-spl: gen-commit hit no-local-replace hook (false freeHost early peel?)\nout:\n%s", out)
	}
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

// assertNoAbsoluteExternalReplaceForLeaf fails if consumer go.mod still has an
// absolute-path replace NewPath for LeafModulePath (D1 pin must drop it).
func assertNoAbsoluteExternalReplaceForLeaf(t *testing.T, req *Request) {
	t.Helper()
	if req.LeafModulePath == "" {
		t.Fatal("LeafModulePath required")
	}
	checkout := req.MainRepo
	if req.WtDir != "" {
		if _, err := os.Stat(req.WtDir); err == nil {
			checkout = req.WtDir
		}
	}
	if checkout == "" {
		t.Fatal("MainRepo or WtDir required")
	}
	goMod := readFile(t, filepath.Join(checkout, "go.mod"))
	for _, line := range strings.Split(goMod, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.Contains(trim, req.LeafModulePath) || !strings.Contains(trim, "=>") {
			continue
		}
		// replace module => /abs/path or module => C:\...
		parts := strings.SplitN(trim, "=>", 2)
		if len(parts) != 2 {
			continue
		}
		newPath := strings.TrimSpace(parts[1])
		if filepath.IsAbs(newPath) || strings.HasPrefix(newPath, "/") {
			t.Fatalf("D1: consumer go.mod must DROP absolute external replace for %s:\n%s",
				req.LeafModulePath, goMod)
		}
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

// assertLinkedBranchNotDivergedFromMain fails when linked branch and main have
// different tips (post-land --merge-back should leave them identical or
// fast-forwardable with zero unique commits either side).
func assertLinkedBranchNotDivergedFromMain(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo == "" || req.WtBranch == "" {
		t.Fatal("MainRepo and WtBranch required for diverge assert")
	}
	main := req.MainRepo
	// Count commits only on main vs only on linked branch.
	// Format: <left>\t<right> for main...branch (left=main-only, right=branch-only).
	out := strings.TrimSpace(gitOutputIsolated(t, main, "rev-list", "--left-right", "--count",
		"main..."+req.WtBranch))
	parts := strings.Fields(out)
	if len(parts) != 2 {
		t.Fatalf("rev-list left-right count parse %q", out)
	}
	if parts[0] != "0" || parts[1] != "0" {
		t.Fatalf("P-empty: root branch must not diverge from main after land (want 0 0, got %s %s)\nmain log:\n%s\nbranch log:\n%s",
			parts[0], parts[1],
			gitOutputIsolated(t, main, "log", "--oneline", "--decorate", "-12", "main"),
			gitOutputIsolated(t, main, "log", "--oneline", "--decorate", "-12", req.WtBranch))
	}
}

// assertCascadePinCommitsNoExternalReplace fails if any cascade pin commit's
// go.mod tree still contains an external-style replace (./external/ or ../).
// Locks --add-all pin isolation: pin subjects must not scoop WIP replaces.
func assertCascadePinCommitsNoExternalReplace(t *testing.T, repo string) {
	t.Helper()
	if repo == "" {
		t.Fatal("repo required")
	}
	// List SHAs of cascade pin commits (newest first).
	log := gitOutputIsolated(t, repo, "log", "--all", "--format=%H", "--grep", cascadePinCommitPrefix)
	if strings.TrimSpace(log) == "" {
		// No pin commits at all is a separate failure (caller may require pins).
		return
	}
	for _, sha := range strings.Split(strings.TrimSpace(log), "\n") {
		sha = strings.TrimSpace(sha)
		if sha == "" {
			continue
		}
		goMod, err := git_isolated.CombinedOutput(repo, "show", sha+":go.mod")
		if err != nil {
			// Pin may touch nested module only; skip missing root go.mod.
			continue
		}
		body := string(goMod)
		for _, line := range strings.Split(body, "\n") {
			trim := strings.TrimSpace(line)
			if !strings.Contains(trim, "=>") {
				continue
			}
			if strings.Contains(trim, "./external/") || strings.Contains(trim, "=> ../") ||
				strings.Contains(trim, "=> /") {
				subj := strings.TrimSpace(gitOutputIsolated(t, repo, "log", "-1", "--format=%s", sha))
				short := sha
				if len(short) > 7 {
					short = short[:7]
				}
				t.Fatalf("P-empty: cascade pin commit must not scoop external replace\ncommit %s %s\nline: %s\ngo.mod:\n%s",
					short, subj, trim, body)
			}
		}
	}
}

// assertMidFeatureLanded checks mid main (DepPath) has the gen-commit feature subject.
func assertMidFeatureLanded(t *testing.T, req *Request) {
	t.Helper()
	mid := req.DepPath
	if mid == "" {
		t.Fatal("DepPath (mid main) required")
	}
	log := gitOutputIsolated(t, mid, "log", "--oneline", "-15")
	if !strings.Contains(log, cascadeFeatureCommitSubject) {
		t.Fatalf("mid main must land feature gen-commit %q\nlog:\n%s",
			cascadeFeatureCommitSubject, log)
	}
	// FEATURE_WIP present on mid history.
	assertFeatureWIPLanded(t, mid)
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
	_ = setupApplyPinOnlyConsumerEmptyGenCommitWithAddAll
	_ = setupApplyPinBeforeFeatureDiamondAllDirtyConsumerTag
	_ = setupApplyPinBeforeFeatureFreeDirtyConsumerAtLatestWIP
	_ = setupApplyPinBeforeFeatureFalseFreeHostIntraPins
	_ = setupApplyPinBeforeFeatureFalseFreeHostIntraPinsWithFeatureWIP
	_ = setupApplyPinBeforeFeatureCleanConsumerCommittedReplace
	_ = setupApplyPinBeforeFeatureResumeFreeLandedUntagged
	_ = setupApplyPinBeforeFeatureAbsolutePathReplace
	_ = setupApplyPinBeforeFeatureFreeMultiModuleCmdRootOnly
	_ = setupApplyDeferredTagRepinAfterFreeUncommittedFeature
	_ = freeMultiCmdModule
	_ = freeMultiCmdNextTag
	_ = assertFreeTagNextBeforeConsumerPinOfFree
	_ = assertNoLocalReplaceGenCommitFail
	_ = assertNoAbsoluteExternalReplaceForLeaf
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
	_ = assertLinkedBranchNotDivergedFromMain
	_ = assertCascadePinCommitsNoExternalReplace
	_ = assertMidFeatureLanded
	_ = historyRepoForConsumer
	_ = featureCommitSHA
	_ = cascadeFeatureWIPFile
	_ = cascadeFeatureCommitSubject
	_ = cascadePinCommitPrefix
	return nil
}
```

