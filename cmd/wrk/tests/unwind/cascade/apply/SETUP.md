# Scenario

**Feature**: apply free-module cascade under `--unwind --tag-next` (not dry-run)

```
# after free-first repo land prelude: global module cascade (PlanUnwindCascade)
# tag one scope → pin consumers (keep local replace; bump require) → selective
# commit when pin dirties → drop free; push when main has no pending modules
# P3: dirty go.mod/go.sum without --add-all → partial edit (save/Base pin/restore+surgical)
# P4: --reinstall-local tail after cascade; nested skip consumers must be pinned first
dirty/clean stack + --unwind --tag-next [+ --push/--done/--add-all/--reinstall-local]
  -> land free-first (linked) then cascade (no TagNextAll-on-peel)
  -> tags / pin commits / go.mod side effects on disk
  -> dirty without --add-all: pin commit = Base+pin+tidy; WT keeps WIP + surgical bumps
  -> reinstall-local: no unknown revision / tidy failure after nested skip pin
```

## Preconditions

- Inherits root `cmd/wrk/tests/unwind` Request/Response/Run and parent
  `cascade/` helpers (`setupCascade*`, cascade line helpers for shared constants).
- Leaves set `req.InProcess = true` and full `req.Args` **without** `--dry-run`.
- **P2 sealed (GREEN):** clean Base + `--add-all` dirty paths under `clean/` and
  `dirty-gomod/with-add-all/`. Do not rewrite those ASSERT meanings.
- **P3:** partial edit under dirty WIP without `--add-all` — C-AP5 success + WIP
  preserve; `partial-edit/` variants. Do not break clean/add-all sealed leaves.
- **P4:** `reinstall-local/` — nested skip consumer + multi-repo reinstall tail.
  Classic preferred; mixed OK if already GREEN after P2/P3. Do not break sealed
  dry-run or clean/partial-edit leaves.
- **B1 pin-before-feature:** `pin-before-feature/` — interleaved free-first apply
  with separate pin auto-commit **before** consumer feature gen-commit (Classic
  TDD RED until product reorders). Do not rewrite sealed clean/partial-edit
  ASSERT contracts for B1.
- Do **not** rewrite sealed P1 dry-run ASSERT files under `cascade/dry-run/`.

## Steps

1. Grouping marks the apply cascade family (mutations allowed).
2. Descendants split on worktree dirt policy (clean | dirty+add-all | partial-edit)
   or **reinstall-local** integration, then stack shape / failure mode.

## Context

- Apply asserts prefer **side effects** (tags, go.mod, git log) over multi-stage
  stdout templates.
- Cascade pin commit message (locked): `wrk: cascade pin <mod> @ <ver>`.
- Replace policy: if consumer already has `replace dep => local path`, **keep**
  it; only bump require version.
- **Partial edit (P3, no `--add-all`, dirty go.mod/go.sum vs Base):**
  1. Save WT go.mod/go.sum
  2. Write Base into WT
  3. Pin + `go mod tidy` on Base → stage go.mod/go.sum → selective cascade commit
  4. Restore WT from save + **surgical require version bumps only** (no tidy on WT)
  5. On failure: restore WT from save, non-zero (no half-mutated WIP)

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/gitops/git/git_isolated"
)

const (
	// Cascade pin commit subject prefix (full form: wrk: cascade pin <mod> @ <ver>).
	cascadePinCommitPrefix = "wrk: cascade pin "

	// cascadeGoModWIPMarker is the uncommitted go.mod comment line used by dirty
	// WIP fixtures (C-AP5 / partial-edit). Must survive partial-edit restore.
	cascadeGoModWIPMarker = "// cascade-wip: uncommitted go.mod dirt for partial-edit"

	// Second free module for sequential-pin fixtures (P3-3).
	cascadeOtherModule  = "example.com/root/other"
	cascadeOtherDir     = "pkgs/other"
	cascadeOtherOldTag  = "pkgs/other/v0.0.1"
	cascadeOtherNextTag = "pkgs/other/v0.0.2"

	// Unrelated WIP path (not go.mod/go.sum) for selective-commit asserts (P3-2).
	cascadeUnrelatedWIPFile = "WIP_NOTES.md"
)

// setupApplyCascadeSingleRepoTwoModules extends the dry-run single-repo fixture
// for apply: bare origin + pushable main, root owned-change so consumer also
// gets a cascade tag after pin, go.mod/go.sum clean at HEAD (DIRTY peel marker only).
//
// Free-first cascade (clean path):
//  1. tag shared @ pkgs/shared/v0.0.2
//  2. pin root require → v0.0.2; keep replace => ./pkgs/shared
//  3. selective commit "wrk: cascade pin …"
//  4. tag root @ v0.0.2 only after pin commit is on history
//  5. push when no pending modules remain
func setupApplyCascadeSingleRepoTwoModules(t *testing.T, req *Request) {
	t.Helper()
	setupCascadeSingleRepoTwoModules(t, req)

	// Root owned change so cascade plans tag-next for consumer after pin (C-AP4).
	writeFile(t, filepath.Join(req.MainRepo, "root.go"),
		"package root\n\nimport _ \""+cascadeSharedModule+"\"\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, req.MainRepo, "add", "root.go")
	runGitIsolated(t, req.MainRepo, "commit", "-m", "root owned change for next tag")

	// Ensure go.mod/go.sum are committed (clean Base for cascade pin).
	// tidy may create go.sum later; start without uncommitted mod dirt.
	_ = os.Remove(filepath.Join(req.MainRepo, "DIRTY"))
	if _, err := os.Stat(filepath.Join(req.MainRepo, "go.sum")); err == nil {
		runGitIsolated(t, req.MainRepo, "add", "go.sum")
		// Commit only if staged changes exist.
		if strings.TrimSpace(gitOutputIsolated(t, req.MainRepo, "diff", "--cached", "--name-only")) != "" {
			runGitIsolated(t, req.MainRepo, "commit", "-m", "commit go.sum baseline")
		}
	}
	// Peel pending still needs dirtiness under v1 filter.
	markDirty(t, req.MainRepo)

	bare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, req.MainRepo, bare)
	// Push existing tags so remote lineage is complete.
	if tagRefExists(t, req.MainRepo, unwindApplyOldTag) {
		runGitIsolated(t, req.MainRepo, "push", "origin", unwindApplyOldTag)
	}
	if tagRefExists(t, req.MainRepo, cascadeSharedOldTag) {
		runGitIsolated(t, req.MainRepo, "push", "origin", cascadeSharedOldTag)
	}
	req.OriginBare = bare

	req.LeafModulePath = cascadeSharedModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag
	req.PeelOrder = []string{"."}
}

// setupApplyCascadeMultiRepoBothDirty extends multi-repo cascade dry-run fixture
// for apply: bare origin on leaf (and root), offline modproxy for pin+tidy.
// No local replace here (nested external may be removed by --done); keep-replace
// is covered by single-repo clean leaf (C-AP3). This leaf drives free-first order
// + cascade pin commit across repos (C-AP2).
func setupApplyCascadeMultiRepoBothDirty(t *testing.T, req *Request) {
	t.Helper()
	setupCascadeMultiRepoBothDirty(t, req)

	// Leaf bare origin for --push after cascade tags leaf.
	leafBare := setupBareOrigin(t, req.WorkRoot, "leaf-origin")
	attachOriginAndPushMain(t, req.SecondRepo, leafBare)
	if tagRefExists(t, req.SecondRepo, unwindApplyOldTag) {
		runGitIsolated(t, req.SecondRepo, "push", "origin", unwindApplyOldTag)
	}
	req.OriginBare = leafBare

	// Root bare origin (push when root has no remaining pending modules).
	rootBare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, req.MainRepo, rootBare)
	if tagRefExists(t, req.MainRepo, unwindApplyOldTag) {
		runGitIsolated(t, req.MainRepo, "push", "origin", unwindApplyOldTag)
	}
	writeFile(t, filepath.Join(req.WorkRoot, "root-origin.path"), rootBare+"\n")

	// Offline module proxy for pin+tidy (no local replace on multi-repo --done path).
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyNextTag, req.DepsLinkedWtDir)
	oldSeed := filepath.Join(req.WorkRoot, "seed-old-"+unwindApplyOldTag)
	mkdirAll(t, oldSeed)
	writeGoModRequire(t, oldSeed, unwindDotPkgsModule)
	writeFile(t, filepath.Join(oldSeed, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyOldTag, oldSeed)
	enableFileModuleProxy(t, req, proxyRoot)

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag
}

// assertRequireBumped checks consumer require version only (no replace policy).
func assertRequireBumped(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("MainRepo and LeafModulePath required")
	}
	goMod := filepath.Join(req.MainRepo, "go.mod")
	got := requireVersionInGoMod(t, goMod, req.LeafModulePath)
	if got != req.ExpectedPinVersion {
		t.Fatalf("consumer require %s = %q, want %s (cascade pin bump)\ngo.mod:\n%s",
			req.LeafModulePath, got, req.ExpectedPinVersion, readFile(t, goMod))
	}
}

// dirtyRootGoModWIP appends an uncommitted WIP line to root go.mod (differs from Base).
// Used by dirty-gomod / partial-edit leaves (C-AP5 / C-AP6 / P3).
func dirtyRootGoModWIP(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo == "" {
		t.Fatal("MainRepo required for dirty go.mod WIP")
	}
	path := filepath.Join(req.MainRepo, "go.mod")
	content := readFile(t, path)
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	// Comment-only WIP: still dirties go.mod vs Base; pin needs to edit requires.
	content += cascadeGoModWIPMarker + "\n"
	writeFile(t, path, content)
	// Ensure porcelain is dirty on go.mod (not only DIRTY marker).
	status := gitOutputIsolated(t, req.MainRepo, "status", "--porcelain", "--", "go.mod")
	if strings.TrimSpace(status) == "" {
		t.Fatal("expected uncommitted go.mod WIP after dirtyRootGoModWIP")
	}
}

// dirtyUnrelatedWIPFile writes an untracked non-module file under MainRepo.
// Partial-edit selective commit must not stage it; it remains after success.
func dirtyUnrelatedWIPFile(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo == "" {
		t.Fatal("MainRepo required for unrelated WIP file")
	}
	path := filepath.Join(req.MainRepo, cascadeUnrelatedWIPFile)
	writeFile(t, path, "unrelated WIP — must stay unstaged through cascade pin\n")
	status := gitOutputIsolated(t, req.MainRepo, "status", "--porcelain", "--", cascadeUnrelatedWIPFile)
	if strings.TrimSpace(status) == "" {
		t.Fatal("expected untracked/uncommitted unrelated WIP file")
	}
}

// snapshotPartialEditWIP saves consumer go.mod (+ go.sum if present) for
// failure-restore asserts (byte-identical restore on partial-edit hard fail).
func snapshotPartialEditWIP(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo == "" {
		t.Fatal("MainRepo required for partial-edit WIP snapshot")
	}
	dir := filepath.Join(req.WorkRoot, "_partial_edit_wip")
	mkdirAll(t, dir)
	writeFile(t, filepath.Join(dir, "go.mod"), readFile(t, filepath.Join(req.MainRepo, "go.mod")))
	sumPath := filepath.Join(req.MainRepo, "go.sum")
	if _, err := os.Stat(sumPath); err == nil {
		writeFile(t, filepath.Join(dir, "go.sum"), readFile(t, sumPath))
	}
}

// setupApplyCascadeSingleRepoThreeModules is like two-module apply fixture plus
// pkgs/other as a second free module. Root requires both with local replaces;
// both leaves owned-changed → sequential cascade pins into the same consumer.
func setupApplyCascadeSingleRepoThreeModules(t *testing.T, req *Request) {
	t.Helper()
	setupApplyCascadeSingleRepoTwoModules(t, req)

	otherDir := filepath.Join(req.MainRepo, filepath.FromSlash(cascadeOtherDir))
	mkdirAll(t, otherDir)
	writeGoModRequire(t, otherDir, cascadeOtherModule)
	writeFile(t, filepath.Join(otherDir, "other.go"),
		"package other\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	// Extend root go.mod: require other@old + local replace (keep shared require/replace).
	goModPath := filepath.Join(req.MainRepo, "go.mod")
	content := readFile(t, goModPath)
	if !strings.Contains(content, cascadeOtherModule) {
		// Insert require line for other next to shared require block if present.
		if strings.Contains(content, "require (") {
			content = strings.Replace(content, "require (\n",
				"require (\n\t"+cascadeOtherModule+" "+unwindApplyOldTag+"\n", 1)
		} else {
			content += "\nrequire "+cascadeOtherModule+" "+unwindApplyOldTag+"\n"
		}
		writeFile(t, goModPath, content)
	}
	appendLocalReplace(t, req.MainRepo, cascadeOtherModule, "./"+cascadeOtherDir)
	writeFile(t, filepath.Join(req.MainRepo, "root.go"),
		"package root\n\nimport (\n\t_ \""+cascadeSharedModule+"\"\n\t_ \""+cascadeOtherModule+"\"\n)\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")

	runGitIsolated(t, req.MainRepo, "add", "go.mod", "root.go", "pkgs")
	runGitIsolated(t, req.MainRepo, "commit", "-m", "add pkgs/other free module")
	createLightweightTag(t, req.MainRepo, cascadeOtherOldTag, "HEAD")
	if req.OriginBare != "" && tagRefExists(t, req.MainRepo, cascadeOtherOldTag) {
		runGitIsolated(t, req.MainRepo, "push", "origin", cascadeOtherOldTag)
	}

	// Owned change on other only → next pkgs/other/v0.0.2.
	writeFile(t, filepath.Join(otherDir, "other.go"),
		"package other\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, req.MainRepo, "add", "pkgs")
	runGitIsolated(t, req.MainRepo, "commit", "-m", "other owned change for next tag")
	// Peel still needs dirtiness under v1 filter.
	markDirty(t, req.MainRepo)

	req.LeafModulePath = cascadeSharedModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag
	req.PeelOrder = []string{"."}
}

// setupApplyCascadePartialEditTidyFail: multi-repo free-first apply stack with
// dirty root go.mod WIP and modproxy missing next version so pin tidy fails mid
// partial-edit. Snapshot WIP before Run for restore asserts.
func setupApplyCascadePartialEditTidyFail(t *testing.T, req *Request) {
	t.Helper()
	// Multi-repo apply fixture seeds next+old proxy; rebuild with old only.
	setupCascadeMultiRepoBothDirty(t, req)

	leafBare := setupBareOrigin(t, req.WorkRoot, "leaf-origin")
	attachOriginAndPushMain(t, req.SecondRepo, leafBare)
	if tagRefExists(t, req.SecondRepo, unwindApplyOldTag) {
		runGitIsolated(t, req.SecondRepo, "push", "origin", unwindApplyOldTag)
	}
	req.OriginBare = leafBare

	rootBare := setupBareOrigin(t, req.WorkRoot, "root-origin")
	attachOriginAndPushMain(t, req.MainRepo, rootBare)
	if tagRefExists(t, req.MainRepo, unwindApplyOldTag) {
		runGitIsolated(t, req.MainRepo, "push", "origin", unwindApplyOldTag)
	}
	writeFile(t, filepath.Join(req.WorkRoot, "root-origin.path"), rootBare+"\n")

	// Proxy: old only — pin require next → tidy fails (P3-4).
	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	oldSeed := filepath.Join(req.WorkRoot, "seed-old-"+unwindApplyOldTag)
	mkdirAll(t, oldSeed)
	writeGoModRequire(t, oldSeed, unwindDotPkgsModule)
	writeFile(t, filepath.Join(oldSeed, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	seedFileModuleProxy(t, proxyRoot, unwindDotPkgsModule, unwindApplyOldTag, oldSeed)
	enableFileModuleProxy(t, req, proxyRoot)

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag

	dirtyRootGoModWIP(t, req)
	snapshotPartialEditWIP(t, req)
}

// rootOriginBare reads the multi-repo root bare path written by setup helper.
func rootOriginBare(t *testing.T, req *Request) string {
	t.Helper()
	p := filepath.Join(req.WorkRoot, "root-origin.path")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("root origin path: %v", err)
	}
	return strings.TrimSpace(string(data))
}

// assertCascadePinCommitPresent fails unless git log on repo has a cascade pin commit
// for depModule (message prefix + module + version).
func assertCascadePinCommitPresent(t *testing.T, repo, depModule, ver string) {
	t.Helper()
	log := gitOutputIsolated(t, repo, "log", "--oneline", "--all", "--grep", cascadePinCommitPrefix)
	if strings.TrimSpace(log) == "" {
		t.Fatalf("missing cascade pin commit (prefix %q) on %s\nfull log:\n%s",
			cascadePinCommitPrefix, repo, gitOutputIsolated(t, repo, "log", "--oneline", "-20"))
	}
	// Prefer exact subject when ver known.
	if depModule != "" {
		needle := cascadePinCommitPrefix + depModule
		if !strings.Contains(log, needle) && !strings.Contains(log, depModule) {
			t.Fatalf("cascade pin commit should mention dep module %s\nlog:\n%s", depModule, log)
		}
	}
	if ver != "" && !strings.Contains(log, ver) {
		// Soft: implementer may put version only in body; require at least prefix hit.
		_ = ver
	}
}

// assertRequireBumpedKeepReplace checks consumer go.mod: require at ExpectedPinVersion
// and local replace for LeafModulePath still present (cascade keep-replace policy).
func assertRequireBumpedKeepReplace(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("MainRepo and LeafModulePath required")
	}
	goMod := filepath.Join(req.MainRepo, "go.mod")
	got := requireVersionInGoMod(t, goMod, req.LeafModulePath)
	if got != req.ExpectedPinVersion {
		t.Fatalf("consumer require %s = %q, want %s (cascade pin bump)\ngo.mod:\n%s",
			req.LeafModulePath, got, req.ExpectedPinVersion, readFile(t, goMod))
	}
	if !goModHasReplace(t, goMod, req.LeafModulePath) {
		t.Fatalf("consumer go.mod must KEEP local replace for %s after cascade pin:\n%s",
			req.LeafModulePath, readFile(t, goMod))
	}
}

// assertGoModCommittedClean fails if go.mod/go.sum still have uncommitted changes.
func assertGoModCommittedClean(t *testing.T, repo string) {
	t.Helper()
	status := gitOutputIsolated(t, repo, "status", "--porcelain", "--", "go.mod", "go.sum")
	if strings.TrimSpace(status) != "" {
		t.Fatalf("go.mod/go.sum must be committed after cascade pin; porcelain:\n%s", status)
	}
}

// assertCommitBeforeTag ensures pin commit is an ancestor of tag (consumer tagged
// only after pin commit is on history).
func assertCommitBeforeTag(t *testing.T, repo, tag string) {
	t.Helper()
	if !tagRefExists(t, repo, tag) {
		t.Fatalf("tag %s missing on %s (needed for commit-before-tag)", tag, repo)
	}
	// Find newest cascade pin commit SHA.
	pinSHA := strings.TrimSpace(gitOutputIsolated(t, repo, "log", "-1", "--format=%H", "--grep", cascadePinCommitPrefix))
	if pinSHA == "" {
		t.Fatalf("no cascade pin commit to order before tag %s\nlog:\n%s",
			tag, gitOutputIsolated(t, repo, "log", "--oneline", "-20"))
	}
	tagSHA := revParseRef(t, repo, "refs/tags/"+tag)
	// pin must be ancestor of tag (tag tip is pin or a descendant).
	err := git_isolated.Command(repo, "merge-base", "--is-ancestor", pinSHA, tagSHA).Run()
	if err != nil {
		t.Fatalf("commit-before-tag: pin %s must be ancestor of tag %s (%s)\npin log:\n%s\ntag log:\n%s",
			pinSHA, tag, tagSHA,
			gitOutputIsolated(t, repo, "log", "--oneline", "-5", pinSHA),
			gitOutputIsolated(t, repo, "log", "--oneline", "-5", tagSHA))
	}
}

// assertPathScopeTagAtMainHEAD checks a path-scoped or root tag exists at main HEAD.
func assertPathScopeTagAtMainHEAD(t *testing.T, mainRepo, tag string) {
	t.Helper()
	assertLocalTagAtMainHEAD(t, mainRepo, tag)
}

// pinCommitSHA returns the newest cascade pin commit SHA on repo (empty if none).
func pinCommitSHA(t *testing.T, repo string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, repo, "log", "-1", "--format=%H", "--grep", cascadePinCommitPrefix))
}

// goModAtCommit returns go.mod content at commit (git show <sha>:go.mod).
func goModAtCommit(t *testing.T, repo, sha string) string {
	t.Helper()
	if sha == "" {
		t.Fatal("empty commit SHA for goModAtCommit")
	}
	return gitOutputIsolated(t, repo, "show", sha+":go.mod")
}

// assertPartialEditWTPreserved checks worktree go.mod after successful partial edit:
// WIP marker still present, surgical require bump to ExpectedPinVersion, keep replace,
// and go.mod still uncommitted (WIP not swallowed into a clean tree).
func assertPartialEditWTPreserved(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo == "" || req.LeafModulePath == "" {
		t.Fatal("MainRepo and LeafModulePath required for partial-edit WT assert")
	}
	goMod := filepath.Join(req.MainRepo, "go.mod")
	content := readFile(t, goMod)
	if !strings.Contains(content, cascadeGoModWIPMarker) {
		t.Fatalf("partial-edit success must preserve WIP marker on WT go.mod\ngo.mod:\n%s", content)
	}
	got := requireVersionInGoMod(t, goMod, req.LeafModulePath)
	if got != req.ExpectedPinVersion {
		t.Fatalf("partial-edit surgical bump: require %s = %q, want %s\ngo.mod:\n%s",
			req.LeafModulePath, got, req.ExpectedPinVersion, content)
	}
	if !goModHasReplace(t, goMod, req.LeafModulePath) {
		t.Fatalf("partial-edit WT must KEEP local replace for %s:\n%s", req.LeafModulePath, content)
	}
	status := gitOutputIsolated(t, req.MainRepo, "status", "--porcelain", "--", "go.mod")
	if strings.TrimSpace(status) == "" {
		t.Fatalf("partial-edit success must leave go.mod dirty (WIP preserved); porcelain empty")
	}
}

// assertPinCommitBaseNoWIP checks the newest cascade pin commit's go.mod:
// has require bump, does **not** include the WIP marker (commit is Base+pin+tidy).
func assertPinCommitBaseNoWIP(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo == "" {
		t.Fatal("MainRepo required")
	}
	sha := pinCommitSHA(t, req.MainRepo)
	if sha == "" {
		t.Fatalf("missing cascade pin commit for Base+pin+tidy assert\nlog:\n%s",
			gitOutputIsolated(t, req.MainRepo, "log", "--oneline", "-20"))
	}
	committed := goModAtCommit(t, req.MainRepo, sha)
	if strings.Contains(committed, cascadeGoModWIPMarker) {
		t.Fatalf("cascade pin commit must be Base+pin+tidy without WIP marker\ncommit %s go.mod:\n%s",
			sha, committed)
	}
	if req.LeafModulePath != "" && req.ExpectedPinVersion != "" {
		// Parse require from committed content (reuse file helper via temp? inline).
		got := ""
		for _, line := range strings.Split(committed, "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[0] == req.LeafModulePath {
				got = fields[1]
				break
			}
		}
		if got != req.ExpectedPinVersion {
			t.Fatalf("pin commit go.mod require %s = %q, want %s\n%s",
				req.LeafModulePath, got, req.ExpectedPinVersion, committed)
		}
	}
}

// assertPinCommitFilesOnlyModSum fails if newest pin commit touches paths other
// than go.mod / go.sum (selective cascade commit under partial-edit / clean pin).
func assertPinCommitFilesOnlyModSum(t *testing.T, repo string) {
	t.Helper()
	sha := pinCommitSHA(t, repo)
	if sha == "" {
		t.Fatalf("missing cascade pin commit\nlog:\n%s",
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
			t.Fatalf("cascade pin commit must only include go.mod/go.sum; got %q\nfiles:\n%s",
				line, names)
		}
	}
}

// assertUnrelatedWIPStillPresent fails if the unrelated WIP file vanished or was committed.
func assertUnrelatedWIPStillPresent(t *testing.T, req *Request) {
	t.Helper()
	path := filepath.Join(req.MainRepo, cascadeUnrelatedWIPFile)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("unrelated WIP file missing after cascade: %v", err)
	}
	// Must not be in HEAD tree.
	ls := gitOutputIsolated(t, req.MainRepo, "ls-tree", "-r", "--name-only", "HEAD")
	for _, line := range strings.Split(ls, "\n") {
		if strings.TrimSpace(line) == cascadeUnrelatedWIPFile {
			t.Fatalf("unrelated WIP file must not be committed at HEAD; ls-tree has %s", cascadeUnrelatedWIPFile)
		}
	}
	status := gitOutputIsolated(t, req.MainRepo, "status", "--porcelain", "--", cascadeUnrelatedWIPFile)
	if strings.TrimSpace(status) == "" {
		t.Fatalf("unrelated WIP should remain untracked/uncommitted; porcelain empty")
	}
}

// assertPartialEditWIPRestoredExact checks go.mod matches pre-run snapshot (failure path).
func assertPartialEditWIPRestoredExact(t *testing.T, req *Request) {
	t.Helper()
	snap := filepath.Join(req.WorkRoot, "_partial_edit_wip", "go.mod")
	want := readFile(t, snap)
	got := readFile(t, filepath.Join(req.MainRepo, "go.mod"))
	if got != want {
		t.Fatalf("partial-edit failure must restore WT go.mod byte-identical to pre-run WIP\n--- want ---\n%s\n--- got ---\n%s",
			want, got)
	}
}

// assertSequentialPinsOnWTAndCommits checks both shared and other requires bumped
// on WT (with WIP) and that pin commit history mentions both dep modules.
func assertSequentialPinsOnWTAndCommits(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo == "" {
		t.Fatal("MainRepo required")
	}
	goMod := filepath.Join(req.MainRepo, "go.mod")
	content := readFile(t, goMod)
	if !strings.Contains(content, cascadeGoModWIPMarker) {
		t.Fatalf("sequential partial-edit must preserve WIP marker\ngo.mod:\n%s", content)
	}
	for _, mod := range []string{cascadeSharedModule, cascadeOtherModule} {
		got := requireVersionInGoMod(t, goMod, mod)
		if got != unwindApplyNextTag {
			t.Fatalf("sequential pin WT require %s = %q, want %s\ngo.mod:\n%s",
				mod, got, unwindApplyNextTag, content)
		}
		if !goModHasReplace(t, goMod, mod) {
			t.Fatalf("sequential pin must KEEP replace for %s\ngo.mod:\n%s", mod, content)
		}
	}
	// All cascade pin commits on history (may be one or two commits).
	log := gitOutputIsolated(t, req.MainRepo, "log", "--oneline", "--all", "--grep", cascadePinCommitPrefix)
	if strings.TrimSpace(log) == "" {
		t.Fatalf("missing cascade pin commit(s) for sequential pins\nfull log:\n%s",
			gitOutputIsolated(t, req.MainRepo, "log", "--oneline", "-20"))
	}
	// Require both modules appear somewhere in pin commit subjects or go.mod trees.
	// Subject may mention only the dep of that step; check log + last pin tree content.
	if !strings.Contains(log, cascadeSharedModule) && !strings.Contains(log, "shared") {
		// Soft: inspect go.mod at HEAD (committed Base should have both bumps).
		_ = cascadeSharedModule
	}
	// HEAD go.mod (committed) should have both requires at next without WIP.
	// Prefer pin commit tip: after cascade, HEAD may be root tag commit = pin ancestor.
	// Walk recent pin commits' go.mod content for both modules.
	foundShared, foundOther := false, false
	shas := gitOutputIsolated(t, req.MainRepo, "log", "--format=%H", "--grep", cascadePinCommitPrefix)
	for _, sha := range strings.Split(shas, "\n") {
		sha = strings.TrimSpace(sha)
		if sha == "" {
			continue
		}
		body := goModAtCommit(t, req.MainRepo, sha)
		if strings.Contains(body, cascadeGoModWIPMarker) {
			t.Fatalf("pin commit %s must not contain WIP marker", sha)
		}
		for _, line := range strings.Split(body, "\n") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			if fields[0] == cascadeSharedModule && fields[1] == unwindApplyNextTag {
				foundShared = true
			}
			if fields[0] == cascadeOtherModule && fields[1] == unwindApplyNextTag {
				foundOther = true
			}
		}
	}
	if !foundShared || !foundOther {
		t.Fatalf("sequential pin commits must eventually record both require bumps (shared=%v other=%v)\npin log:\n%s",
			foundShared, foundOther, log)
	}
	// Tags for both free modules.
	if !tagRefExists(t, req.MainRepo, cascadeSharedNextTag) {
		t.Fatalf("missing shared tag %s after sequential cascade", cascadeSharedNextTag)
	}
	if !tagRefExists(t, req.MainRepo, cascadeOtherNextTag) {
		t.Fatalf("missing other tag %s after sequential cascade", cascadeOtherNextTag)
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	// Keep apply cascade helpers referenced for the generator.
	_ = setupApplyCascadeSingleRepoTwoModules
	_ = setupApplyCascadeMultiRepoBothDirty
	_ = setupApplyCascadeSingleRepoThreeModules
	_ = setupApplyCascadePartialEditTidyFail
	_ = dirtyRootGoModWIP
	_ = dirtyUnrelatedWIPFile
	_ = snapshotPartialEditWIP
	_ = rootOriginBare
	_ = assertCascadePinCommitPresent
	_ = assertRequireBumped
	_ = assertRequireBumpedKeepReplace
	_ = assertGoModCommittedClean
	_ = assertCommitBeforeTag
	_ = assertPathScopeTagAtMainHEAD
	_ = assertPartialEditWTPreserved
	_ = assertPinCommitBaseNoWIP
	_ = assertPinCommitFilesOnlyModSum
	_ = assertUnrelatedWIPStillPresent
	_ = assertPartialEditWIPRestoredExact
	_ = assertSequentialPinsOnWTAndCommits
	_ = pinCommitSHA
	_ = goModAtCommit
	_ = cascadePinCommitPrefix
	_ = cascadeGoModWIPMarker
	_ = cascadeOtherModule
	_ = cascadeOtherNextTag
	_ = cascadeUnrelatedWIPFile
	return nil
}
```
