# Scenario

**Feature**: wrk compose multi-stage pipeline with fixed stage order and `activeRoot` switching

```
# Stage order (all optional; skip absent). CLI flag order free; execution order fixed.
# 1. gen-commit
# 2. --done | --merge-back   then activeRoot := main
# 3. sync
# 4. tag-next                (only when activeRoot is main)
# 5. push                    (follows activeRoot)
# 6. propagate-tags          (repo-level; OK from WT or main)
# 7. reinstall-local         (follows activeRoot)
# 8. exec                    (last; runs in final activeRoot)

# activeRoot starts as git toplevel of effective cwd.
# After successful --done/--merge-back: activeRoot := main for stages 3–8.
# --main + pipeline partners (no shell): resolve main for this checkout,
#   activeRoot := main at start, then stages 3–8 (no nested shell; no done/merge/remove).
# Without those switches: activeRoot stays cwd for the whole run.
# Bare wrk --main alone: nested shell at main (see main/ tree; not this compose model).

# Gates
# - --done/--merge-back: must run from linked worktree (main → Error naming flag + linked worktree)
# - --tag-next: only when that stage's activeRoot is main
#     with done/merge-back: after switch → OK
#     with --main + pipeline partners: activeRoot already main → OK from linked WT
#     without those: cwd must be main; linked WT → error naming --tag-next + main repository
# - Only --tag-next is main-only among pipeline stages
# - --json: bare --tag-next only; multi-stage + --json rejected
# - --done and --merge-back still exclusive of each other
# - --main is exclusive with --done / --merge-back / --gen-commit-msg
# - --main partners allowed: sync, tag-next, push, propagate-tags, reinstall-local, dry-run, exec
#     (+ status / reinstall-alone covered elsewhere)
# - Already on main + --main + pipeline: notice (--main not necessary; continuing); exit 0 if rest OK
# - --exec valid as last stage of any compose (not only with --done)
```

## Preconditions

- Reuses root `cmd/wrk/tests` harness (`Request` / `Response` / `Run`).
- Isolated `WRK_HOME` per leaf via root Setup.
- Classic TDD target model: multi-stage compose without primary is **allowed** under gates;
  `activeRoot` drives where post stages run. Existing `done-compose/still-exclusive/*-with-sync`
  leaves that asserted bare multi-stage mutex are updated to match this model.
- Helpers below mirror done-pipeline fixtures so this tree is self-contained (no inheritance
  from sibling trees).

## Steps

- Grouping only. Leaves call fixture helpers and set `req.Args` / `req.RepoDir`.

## Context

- Prefer dry-run for ordered multi-stage success paths (plan markers, zero mutations).
- Gate/error leaves assert non-zero exit and stderr naming the flag + requirement.
- Vocabulary: **stage**, **activeRoot** — not “primary” / “finish step” in product-facing asserts.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/gitops/git/git_isolated"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}

func createLightweightTagAP(t *testing.T, repo, name, ref string) {
	t.Helper()
	if ref == "" {
		ref = "HEAD"
	}
	runGitIsolated(t, repo, "tag", name, ref)
}

func tagRefExistsAP(t *testing.T, repo, name string) bool {
	t.Helper()
	err := git_isolated.Command(repo, "rev-parse", "--verify", "refs/tags/"+name).Run()
	return err == nil
}

func revParseRefAP(t *testing.T, repo, ref string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, repo, "rev-parse", ref))
}

func setupBareOriginAP(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

// seedMainRootBumpAP: main-gomod + lightweight v0.0.1 (tag-next plans v0.0.2 after owned change).
func seedMainRootBumpAP(t *testing.T, req *Request) string {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	cloneMainGoModFromSeed(t, mainRepo)
	mainRepo = compositionResolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	createLightweightTagAP(t, mainRepo, "v0.0.1", "")
	return mainRepo
}

// setupAPLinkedAhead: root-bump + wrk-managed linked wt ahead (no origin).
func setupAPLinkedAhead(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := seedMainRootBumpAP(t, req)
	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
}

// setupAPLinkedAheadOrigin: linked ahead + bare origin tracking main.
func setupAPLinkedAheadOrigin(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := seedMainRootBumpAP(t, req)
	bare := setupBareOriginAP(t, req.WorkRoot, "origin")
	runGitIsolated(t, mainRepo, "remote", "add", "origin", bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")
	runGitIsolated(t, mainRepo, "push", "origin", "v0.0.1")
	req.OriginBare = bare

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
}

// setupAPSyncWithOrigin: two worktrees + origin (wtA ahead; wtB stays).
func setupAPSyncWithOrigin(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := seedMainRootBumpAP(t, req)
	bare := setupBareOriginAP(t, req.WorkRoot, "origin")
	runGitIsolated(t, mainRepo, "remote", "add", "origin", bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")
	runGitIsolated(t, mainRepo, "push", "origin", "v0.0.1")
	req.OriginBare = bare

	wtA := runWrkFrom(t, req, mainRepo)
	wtA = compositionResolvePath(t, wtA)
	req.WtDir = wtA
	req.WtBranch = branchName("main", wrkDate, 0)

	wt2Path := filepath.Join(req.WorkRoot, "wt-stays")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "feature-stays", wt2Path)
	wt2Path = compositionResolvePath(t, wt2Path)
	req.Wt2Dir = wt2Path
	req.Wt2Branch = "feature-stays"

	commitAheadOnWorktree(t, wtA, "feature-work", "ahead of main")
	req.RepoDir = wtA
}

// setupAPMainOnOrigin: main checkout only (activeRoot stays main); origin + root-bump + ahead commit on main.
func setupAPMainOnOrigin(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := seedMainRootBumpAP(t, req)
	bare := setupBareOriginAP(t, req.WorkRoot, "origin")
	runGitIsolated(t, mainRepo, "remote", "add", "origin", bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")
	runGitIsolated(t, mainRepo, "push", "origin", "v0.0.1")
	req.OriginBare = bare

	// Owned change on main so tag-next can plan v0.0.2 without a worktree merge.
	writeFile(t, filepath.Join(mainRepo, "feature-work"), "on main\n")
	runGitIsolated(t, mainRepo, "add", "feature-work")
	runGitIsolated(t, mainRepo, "commit", "-m", "feature on main")
	req.RepoDir = mainRepo
	req.WtBranch = "main"
}

func seedAPReinstallPresent(t *testing.T, req *Request) string {
	t.Helper()
	if req.MainRepo == "" {
		t.Fatal("seedAPReinstallPresent: MainRepo empty")
	}
	src := fmt.Sprintf("package %s\n\nfunc main() {}\n", "main")
	cmdDir := filepath.Join(req.MainRepo, "cmd", "present")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatalf("mkdir cmd/present: %v", err)
	}
	writeFile(t, filepath.Join(cmdDir, "main.go"), src)
	// Commit present so reinstall scan sees it on the tip used by the stage.
	runGitIsolated(t, req.MainRepo, "add", "cmd/present")
	runGitIsolated(t, req.MainRepo, "commit", "-m", "add present cmd")

	binDir := filepath.Join(req.WorkRoot, "gobin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir gobin: %v", err)
	}
	writeFile(t, filepath.Join(binDir, "present"), "stub-binary\n")
	if err := os.Chmod(filepath.Join(binDir, "present"), 0o755); err != nil {
		t.Fatalf("chmod present stub: %v", err)
	}
	req.ExtraEnv = append(req.ExtraEnv, "GOBIN="+binDir)
	return binDir
}

func recordAPDryRunBaseline(t *testing.T, req *Request) {
	t.Helper()
	dir := filepath.Join(req.WorkRoot, "_compose_dry_run_baseline")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir baseline: %v", err)
	}
	writeFile(t, filepath.Join(dir, "main.sha"), revParseHEAD(t, req.MainRepo)+"\n")
	if req.WtDir != "" {
		writeFile(t, filepath.Join(dir, "wt.sha"), revParseHEAD(t, req.WtDir)+"\n")
	}
	if req.Wt2Dir != "" {
		writeFile(t, filepath.Join(dir, "wt2.sha"), revParseHEAD(t, req.Wt2Dir)+"\n")
	}
	if req.OriginBare != "" {
		writeFile(t, filepath.Join(dir, "origin-main.sha"), revParseRefAP(t, req.OriginBare, "refs/heads/main")+"\n")
	}
}

func readAPBaseline(t *testing.T, req *Request, name string) string {
	t.Helper()
	p := filepath.Join(req.WorkRoot, "_compose_dry_run_baseline", name)
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read baseline %s: %v", name, err)
	}
	return strings.TrimSpace(string(data))
}

func tagNextRootBumpPlanAP() string {
	return "v0.0.1        owned changed                  ->  v0.0.2\n1 tag planned\n"
}

func wouldPushMainOriginAP(tag string) string {
	if tag == "" {
		return "would: git push origin main\n"
	}
	return fmt.Sprintf("would: git push origin main\nwould: git push origin %s\n", tag)
}

func assertNoMutexReject(t *testing.T, se string) {
	t.Helper()
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("compose multi-stage still mutually exclusive; stderr=%q", se)
	}
}

func assertAPNoConfirmNoise(t *testing.T, resp *Response) {
	t.Helper()
	assertNotContains(t, resp.Stdout, "Proceed?")
	assertNotContains(t, resp.Stderr, "Proceed?")
	assertNotContains(t, resp.Stderr, "stdin is not a terminal")
	assertNotContains(t, resp.Stderr, "cannot prompt")
}

func assertAPDryRunZeroMutationsLinked(t *testing.T, req *Request) {
	t.Helper()
	if req.WtDir != "" {
		assertFileExists(t, req.WtDir)
		assertGitFileIsWorktreeLink(t, req.WtDir)
		if req.WtBranch != "" && req.WtBranch != "main" {
			assertBranchExists(t, req.MainRepo, req.WtBranch)
			assertWorktreeListContains(t, req.MainRepo, req.WtDir)
		}
	}
	mainSHA := revParseHEAD(t, req.MainRepo)
	if want := readAPBaseline(t, req, "main.sha"); mainSHA != want {
		t.Fatalf("main HEAD mutated under dry-run: got %s want %s", mainSHA, want)
	}
	if tagRefExistsAP(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 must not be created under dry-run")
	}
	if req.OriginBare != "" {
		originSHA := revParseRefAP(t, req.OriginBare, "refs/heads/main")
		if want := readAPBaseline(t, req, "origin-main.sha"); originSHA != want {
			t.Fatalf("origin/main mutated under dry-run: got %s want %s", originSHA, want)
		}
	}
}

func assertReinstallDryRunAP(t *testing.T, stdout string) {
	t.Helper()
	if !strings.Contains(stdout, "would: go install ./cmd/present") {
		t.Fatalf("stdout missing reinstall would: go install ./cmd/present\nfull:\n%s", stdout)
	}
}

func assertStubPresentAP(t *testing.T, binDir string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(binDir, "present"))
	if err != nil {
		t.Fatalf("read present stub: %v", err)
	}
	if string(data) != "stub-binary\n" {
		t.Fatalf("present bin mutated under dry-run: %q", string(data))
	}
}

var (
	_ = setupAPLinkedAhead
	_ = setupAPLinkedAheadOrigin
	_ = setupAPSyncWithOrigin
	_ = setupAPMainOnOrigin
	_ = seedAPReinstallPresent
	_ = recordAPDryRunBaseline
	_ = readAPBaseline
	_ = tagNextRootBumpPlanAP
	_ = wouldPushMainOriginAP
	_ = assertNoMutexReject
	_ = assertAPNoConfirmNoise
	_ = assertAPDryRunZeroMutationsLinked
	_ = assertReinstallDryRunAP
	_ = assertStubPresentAP
	_ = createLightweightTagAP
	_ = tagRefExistsAP
	_ = revParseRefAP
)
```
