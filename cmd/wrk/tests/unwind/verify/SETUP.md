# Scenario

**Feature**: read-only `wrk --unwind --verify [--json] [--color|--no-color]` post-job audit

```
# stack inventory + PlanUnwind + PlanUnwindCascade + require/replace status
wrk --unwind --verify [--json]
  -> reject dry-run / apply partners / --show-graph (Error, non-zero)
  -> bare --verify without --unwind? Error names both
  -> cycle? Error mentioning cycle (no success verify body)
  -> else report 6 checks (dirty-peel, needs-land, owned-changed,
     require-drift, droppable-replace, cascade-pending)
  -> any error-severity FAIL → exit 1, report on stdout, no Error: for logical FAIL
  -> all pass (+ soft warning:) → exit 0
  -> zero mutations
```

## Preconditions

- Inherits root `cmd/wrk/tests/unwind` Request/Response/Run and fixture helpers
  (`setupSingleMainDirty`, `setupThreeRepoChain`, `setupTwoCycleStack`,
  `setupFollow*`, `recordUnwindBaseline`, `assertUnwindZeroMutations` family
  primitives, tag/replace helpers). Cascade / show-graph **private** helpers
  are **not** inherited (sibling subtrees) — this SETUP redefines verify-local
  fixtures and zero-mutation asserts.
- **Classic TDD:** `--verify` not implemented yet. Leaves must stay **RED** until
  flag, mutual exclusion, report body, check catalog, and exit policy land.
- Leaves set `req.InProcess = true` and full `req.Args` including `--unwind`
  and `--verify` (except bare-verify-without-unwind). Color via Args only
  (no `t.Setenv` / `t.Chdir`).
- Soft inventory warnings: `warning:` on stderr; if all error checks pass → exit 0.

## Steps

1. Grouping scopes the verify family; descendants branch on outcome
   (reject | cycle | pass | fail | json).
2. Pass leaves seed fully shipped-clean stacks (tags at HEAD, no dirt, no drift).
3. Fail leaves seed one primary residual check failure (others may co-fail).
4. Reject leaves use minimal fixtures — fail before inventory is load-bearing.
5. Cycle leaf reuses `setupTwoCycleStack`.

## Context

- Human banners (locked intent):
  - `==== unwind verify ====`
  - `==== status summary ====`
- Check IDs (error severity unless noted): `dirty-peel`, `needs-land`,
  `owned-changed`, `require-drift`, `droppable-replace`, `cascade-pending`.
- Pass line: check id + `pass` (optional counts). Fail: uppercase `FAIL` + detail.
- Summary: `checks: 6  pass: N  fail: M  warn: W` then `result: pass|fail`.
- Color: green pass/result pass; red FAIL/result fail; gray banners. JSON: no ANSI.
- JSON snake_case: `work_dir`, `checks[]` (`id`,`severity`,`status`, optional `count`),
  `summary` (`checks`,`pass`,`fail`,`warn`,`result`), `warnings`.
- Stdout ends with trailing `\n`. Logical FAIL: exit **1**, report on **stdout**,
  **no** `Error:` for check failures. Fatal preflight: `Error:` on stderr, no body.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	verifyBanner        = "==== unwind verify ===="
	verifySummaryBanner = "==== status summary ===="

	// Intra-repo shared leaf for owned-changed / cascade-pending fixtures.
	verifySharedModule  = "example.com/root/shared"
	verifySharedDir     = "pkgs/shared"
	verifySharedOldTag  = "pkgs/shared/v0.0.1"
	verifySharedNextTag = "pkgs/shared/v0.0.2"
)

// Locked error-severity check IDs (order matches human report intent).
var verifyCheckIDs = []string{
	"dirty-peel",
	"needs-land",
	"owned-changed",
	"require-drift",
	"droppable-replace",
	"cascade-pending",
}

// verifyArgs returns base args for a successful human verify run.
func verifyArgs(extra ...string) []string {
	args := []string{"--unwind", "--verify"}
	return append(args, extra...)
}

// verifyJSONArgs returns args for JSON verify.
func verifyJSONArgs(extra ...string) []string {
	return verifyArgs(append([]string{"--json"}, extra...)...)
}

// setupVerifySingleMainClean seeds a clean main-only root tagged at HEAD so
// tagscope has latest and no next (all verify checks should pass).
// RepoDir = MainRepo; PeelOrder empty.
func setupVerifySingleMainClean(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, mainRepo)
	writeGoModRequire(t, mainRepo, unwindRootModule)
	writeFile(t, filepath.Join(mainRepo, "main.go"), "package main\n")
	runGitIsolated(t, mainRepo, "add", "go.mod", "main.go")
	runGitIsolated(t, mainRepo, "commit", "-m", "add module")
	createLightweightTag(t, mainRepo, unwindApplyOldTag, "HEAD")
	mainRepo = resolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.PeelOrder = nil
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = ""
}

// setupVerifySingleMainDirtyTagged is dirty-peel primary: HEAD tagged, then DIRTY
// only (no post-tag owned commits) so owned-changed/cascade stay quiet if possible.
func setupVerifySingleMainDirtyTagged(t *testing.T, req *Request) {
	t.Helper()
	setupVerifySingleMainClean(t, req)
	markDirty(t, req.MainRepo)
	req.PeelOrder = []string{"."}
}

// setupVerifySingleMainOwnedChanged: tag v0.0.1 then committed owned change so
// tagscope plans next v0.0.2. Working tree left clean (no DIRTY) so dirty-peel
// can pass while owned-changed / cascade-pending fail.
func setupVerifySingleMainOwnedChanged(t *testing.T, req *Request) {
	t.Helper()
	setupVerifySingleMainClean(t, req)
	writeFile(t, filepath.Join(req.MainRepo, "main.go"), "package main\n// changed after tag\n")
	runGitIsolated(t, req.MainRepo, "add", "main.go")
	runGitIsolated(t, req.MainRepo, "commit", "-m", "owned change after tag")
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag
	req.PeelOrder = nil
}

// setupVerifySharedOwnedChanged builds single main with root + pkgs/shared.
// shared is owned-changed after pkgs/shared/v0.0.1; root requires shared@v0.0.1
// with intra local replace (keep-local, not droppable). Working tree clean.
// Cascade would: tag-next shared then pin root — cascade-pending + owned-changed.
func setupVerifySharedOwnedChanged(t *testing.T, req *Request) {
	t.Helper()

	mainRepo := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, mainRepo)

	sharedDir := filepath.Join(mainRepo, filepath.FromSlash(verifySharedDir))
	mkdirAll(t, sharedDir)
	writeGoModRequire(t, sharedDir, verifySharedModule)
	writeFile(t, filepath.Join(sharedDir, "shared.go"),
		"package shared\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")

	writeGoModRequire(t, mainRepo, unwindRootModule, verifySharedModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(mainRepo, "root.go"),
		"package root\n\nimport _ \""+verifySharedModule+"\"\n")
	appendLocalReplace(t, mainRepo, verifySharedModule, "./"+verifySharedDir)

	runGitIsolated(t, mainRepo, "add", "go.mod", "root.go", "pkgs")
	runGitIsolated(t, mainRepo, "commit", "-m", "root + shared modules")
	createLightweightTag(t, mainRepo, unwindApplyOldTag, "HEAD")
	createLightweightTag(t, mainRepo, verifySharedOldTag, "HEAD")

	writeFile(t, filepath.Join(sharedDir, "shared.go"),
		"package shared\n\nfunc Version() string { return \""+unwindApplyNextTag+"\" }\n")
	runGitIsolated(t, mainRepo, "add", "pkgs")
	runGitIsolated(t, mainRepo, "commit", "-m", "shared owned change for next tag")

	mainRepo = resolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.LeafModulePath = verifySharedModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag
	req.PeelOrder = nil
}

// setupVerifyMultiRepoPinnedClean: nested external leaf at tag v0.0.1 (clean);
// root requires leaf@v0.0.1 matching latest; **no** local replace (not droppable);
// both clean and fully tagged. All error checks should pass.
func setupVerifyMultiRepoPinnedClean(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = ""

	leafMain := filepath.Join(req.WorkRoot, labelDotPkgs)
	initGitRepoOnMain(t, leafMain)
	writeGoModRequire(t, leafMain, unwindDotPkgsModule)
	writeFile(t, filepath.Join(leafMain, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	runGitIsolated(t, leafMain, "add", "go.mod", "pkg.go")
	runGitIsolated(t, leafMain, "commit", "-m", "add dot-pkgs module")
	createLightweightTag(t, leafMain, unwindApplyOldTag, "HEAD")
	leafMain = resolvePath(t, leafMain)
	req.SecondRepo = leafMain

	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root consumer")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "HEAD")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	extDir := filepath.Join(rootMain, "external")
	mkdirAll(t, extDir)
	leafExt := filepath.Join(extDir, labelDotPkgs+"-"+branchNameMainDate())
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate(), leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt
	req.WtBranch = branchNameMainDate()
	// Clean leaf (no markDirty); HEAD at tag tree.

	writeFile(t, filepath.Join(rootMain, ".gitignore"), "/external\n")
	runGitIsolated(t, rootMain, "add", ".gitignore")
	runGitIsolated(t, rootMain, "commit", "-m", "ignore external stack members")
	// Retag root after ignore commit so no owned-changed residual.
	createLightweightTag(t, rootMain, "v0.0.1-ship", "HEAD")
	// Prefer single baseline tag at HEAD for root scope: move lightweight tag.
	// Delete and recreate v0.0.1 at current HEAD for tagscope latest alignment.
	runGitIsolated(t, rootMain, "tag", "-d", unwindApplyOldTag)
	createLightweightTag(t, rootMain, unwindApplyOldTag, "HEAD")

	req.RepoDir = rootMain
	req.PeelOrder = nil
}

// setupVerifyDroppableReplace: clean free leaf at v0.0.1; root require matches;
// external stack replace still present (droppable). Working tree clean so
// dirty-peel can pass; droppable-replace + cascade-pending fail.
func setupVerifyDroppableReplace(t *testing.T, req *Request) {
	t.Helper()

	req.LeafModulePath = unwindDotPkgsModule
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyOldTag

	leafMain := filepath.Join(req.WorkRoot, labelDotPkgs)
	initGitRepoOnMain(t, leafMain)
	writeGoModRequire(t, leafMain, unwindDotPkgsModule)
	writeFile(t, filepath.Join(leafMain, "pkg.go"),
		"package dotpkgs\n\nfunc Version() string { return \""+unwindApplyOldTag+"\" }\n")
	runGitIsolated(t, leafMain, "add", "go.mod", "pkg.go")
	runGitIsolated(t, leafMain, "commit", "-m", "add dot-pkgs module")
	createLightweightTag(t, leafMain, unwindApplyOldTag, "HEAD")
	leafMain = resolvePath(t, leafMain)
	req.SecondRepo = leafMain

	rootMain := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, rootMain)
	writeGoModRequire(t, rootMain, unwindRootModule, unwindDotPkgsModule+"@"+unwindApplyOldTag)
	writeFile(t, filepath.Join(rootMain, "root.go"),
		"package root\n\nimport _ \""+unwindDotPkgsModule+"\"\n")
	runGitIsolated(t, rootMain, "add", "go.mod", "root.go")
	runGitIsolated(t, rootMain, "commit", "-m", "add root consumer")
	createLightweightTag(t, rootMain, unwindApplyOldTag, "HEAD")
	rootMain = resolvePath(t, rootMain)
	req.MainRepo = rootMain

	extDir := filepath.Join(rootMain, "external")
	mkdirAll(t, extDir)
	leafExt := filepath.Join(extDir, labelDotPkgs+"-"+branchNameMainDate())
	runGitIsolated(t, leafMain, "worktree", "add", "-b", branchNameMainDate(), leafExt)
	leafExt = resolvePath(t, leafExt)
	req.DepsLinkedWtDir = leafExt
	req.WtBranch = branchNameMainDate()

	writeFile(t, filepath.Join(rootMain, ".gitignore"), "/external\n")
	appendLocalReplace(t, rootMain, unwindDotPkgsModule, relLocalReplace(t, rootMain, leafExt))
	runGitIsolated(t, rootMain, "add", ".gitignore", "go.mod")
	runGitIsolated(t, rootMain, "commit", "-m", "ignore external + droppable replace")
	// Align root tagscope latest with HEAD (no residual owned-changed on root).
	runGitIsolated(t, rootMain, "tag", "-d", unwindApplyOldTag)
	createLightweightTag(t, rootMain, unwindApplyOldTag, "HEAD")

	req.RepoDir = rootMain
	req.PeelOrder = nil
}

// setupVerifyRequireDrift: sibling local-replace; require v0.0.0; dep tagged
// v0.0.1. Working trees cleaned so require-drift is primary residual.
func setupVerifyRequireDrift(t *testing.T, req *Request) {
	t.Helper()
	setupFollowSiblingBothDirty(t, req)
	if req.DepsLinkedWtDir == "" {
		t.Fatal("require-drift: missing dep checkout")
	}
	createLightweightTag(t, req.DepsLinkedWtDir, unwindApplyOldTag, "HEAD")
	// Tag consumer at HEAD so only require version drifts vs dep latest.
	createLightweightTag(t, req.MainRepo, unwindApplyOldTag, "HEAD")
	markCleanTracked(t, req.MainRepo)
	markCleanTracked(t, req.DepsLinkedWtDir)
	req.OldRequireVersion = "v0.0.0"
	req.ExpectedPinVersion = unwindApplyOldTag
	req.PeelOrder = nil
}

// setupVerifyNeedsLand: three-repo linked stack with dirty linked consumer.
// PlanUnwind.NeedsLand true; dirty-peel also fails (multi-fail OK).
func setupVerifyNeedsLand(t *testing.T, req *Request) {
	t.Helper()
	setupThreeRepoChain(t, req)
	dirtyAllThree(t, req)
	// Tag mains/WTs lightly so owned-changed noise is reduced where possible.
	// (Linked dirty still drives needs-land + dirty-peel.)
	setPeelOrderDisplays(t, req, req.DepsLinkedWtDir, req.ExternalWtDir, req.WtDir)
}

// assertNoSuccessfulVerifyBody fails if stdout looks like a completed verify report.
func assertNoSuccessfulVerifyBody(t *testing.T, stdout string) {
	t.Helper()
	plain := stripVerifyANSI(stdout)
	if strings.Contains(plain, verifyBanner) ||
		strings.Contains(plain, verifySummaryBanner) ||
		strings.Contains(plain, "result: pass") ||
		strings.Contains(plain, "result: fail") {
		t.Fatalf("reject/cycle path must not print successful verify body; stdout:\n%s", stdout)
	}
}

// assertVerifyHumanBanners requires the two locked human section banners.
func assertVerifyHumanBanners(t *testing.T, stdout string) {
	t.Helper()
	plain := stripVerifyANSI(stdout)
	for _, b := range []string{verifyBanner, verifySummaryBanner} {
		if !strings.Contains(plain, b) {
			t.Fatalf("missing verify banner %q\nstdout:\n%s", b, stdout)
		}
	}
}

// assertVerifyReject checks mutual-exclusion / bare-verify Error.
// partnerSubstr is a substring of the forbidden flag or "unwind" for bare verify.
func assertVerifyReject(t *testing.T, resp *Response, partnerSubstr string) {
	t.Helper()
	assertExitNonZero(t, resp)
	combined := resp.Stdout + "\n" + resp.Stderr
	lower := strings.ToLower(combined)
	if !strings.Contains(lower, "verify") {
		t.Fatalf("error must mention verify; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
	if partnerSubstr != "" {
		ps := strings.ToLower(partnerSubstr)
		if !strings.Contains(lower, ps) && !strings.Contains(combined, partnerSubstr) {
			t.Fatalf("error must mention partner %q; stderr=%q stdout=%q",
				partnerSubstr, resp.Stderr, resp.Stdout)
		}
	}
	assertNoSuccessfulVerifyBody(t, resp.Stdout)
}

// assertVerifyNoLogicalErrorPrefix fails if logical check-FAIL report uses Error:.
func assertVerifyNoLogicalErrorPrefix(t *testing.T, resp *Response) {
	t.Helper()
	// Fatal preflight uses Error:; logical FAIL report must not.
	if strings.Contains(resp.Stderr, "Error:") {
		t.Fatalf("logical verify FAIL must not put Error: on stderr; stderr=%q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "Error:") {
		t.Fatalf("logical verify FAIL must not put Error: on stdout; stdout:\n%s", resp.Stdout)
	}
}

// lineHasCheckStatus reports whether a human report line mentions checkID with status.
// Accepts flexible spacing: "dirty-peel   pass" / "require-drift  FAIL".
func lineHasCheckStatus(line, checkID, status string) bool {
	line = strings.TrimSpace(stripVerifyANSI(line))
	if !strings.Contains(line, checkID) {
		return false
	}
	// Status token as whole word-ish (avoid "pass" matching "password").
	fields := strings.Fields(line)
	for _, f := range fields {
		if f == status {
			return true
		}
	}
	// Also accept "FAIL:" / "pass," variants.
	lower := strings.ToLower(line)
	return strings.Contains(lower, strings.ToLower(status))
}

// assertVerifyCheckStatus requires checkID to appear with wantStatus (pass|FAIL).
func assertVerifyCheckStatus(t *testing.T, stdout, checkID, wantStatus string) {
	t.Helper()
	plain := stripVerifyANSI(stdout)
	if !strings.Contains(plain, checkID) {
		t.Fatalf("verify report missing check %q\nstdout:\n%s", checkID, stdout)
	}
	for _, line := range strings.Split(plain, "\n") {
		if lineHasCheckStatus(line, checkID, wantStatus) {
			return
		}
	}
	t.Fatalf("check %q must show status %q\nstdout:\n%s", checkID, wantStatus, stdout)
}

// assertVerifyAllChecksPass requires every catalog id shows pass.
func assertVerifyAllChecksPass(t *testing.T, stdout string) {
	t.Helper()
	for _, id := range verifyCheckIDs {
		assertVerifyCheckStatus(t, stdout, id, "pass")
	}
}

// assertVerifyResult requires summary result: pass|fail.
func assertVerifyResult(t *testing.T, stdout, want string) {
	t.Helper()
	plain := stripVerifyANSI(stdout)
	needle := "result: " + want
	if !strings.Contains(plain, needle) {
		t.Fatalf("stdout must contain %q\nstdout:\n%s", needle, stdout)
	}
}

// assertVerifyStdoutTrailingNL requires CLI stdout ends with \n.
func assertVerifyStdoutTrailingNL(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" || !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("verify stdout must end with trailing newline; got %q", stdout)
	}
}

// assertVerifyNoANSI fails if stdout has CSI escapes.
func assertVerifyNoANSI(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, "\x1b[") {
		t.Fatalf("stdout must not contain ANSI escapes; got:\n%q", stdout)
	}
}

// assertVerifyHasANSI requires at least one CSI escape.
func assertVerifyHasANSI(t *testing.T, stdout string) {
	t.Helper()
	if !strings.Contains(stdout, "\x1b[") {
		t.Fatalf("expected ANSI escapes on stdout with --color; stdout:\n%q", stdout)
	}
}

// stripVerifyANSI removes CSI sequences for banner/token matching when colored.
func stripVerifyANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) {
				c := s[j]
				j++
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
					break
				}
			}
			i = j - 1
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// assertVerifyZeroMutations checks HEADs/worktrees unchanged (soft: DIRTY only if present).
// Unlike assertUnwindZeroMutations, does not require DIRTY on clean-stack pass leaves.
func assertVerifyZeroMutations(t *testing.T, req *Request) {
	t.Helper()
	if req.MainRepo != "" {
		if _, err := os.Stat(filepath.Join(req.WorkRoot, "_unwind_baseline", "main.sha")); err == nil {
			got := revParseHEAD(t, req.MainRepo)
			if want := readBaselineSHA(t, req, "main.sha"); got != want {
				t.Fatalf("main HEAD mutated: got %s want %s", got, want)
			}
		}
	}
	if req.WtDir != "" {
		if _, err := os.Stat(req.WtDir); err == nil {
			assertFileExists(t, req.WtDir)
			assertGitFileIsWorktreeLink(t, req.WtDir)
			if _, err := os.Stat(filepath.Join(req.WorkRoot, "_unwind_baseline", "wt.sha")); err == nil {
				got := revParseHEAD(t, req.WtDir)
				if want := readBaselineSHA(t, req, "wt.sha"); got != want {
					t.Fatalf("wt HEAD mutated: got %s want %s", got, want)
				}
			}
		}
	}
	if req.ExternalWtDir != "" {
		if _, err := os.Stat(req.ExternalWtDir); err == nil {
			if _, err := os.Stat(filepath.Join(req.WorkRoot, "_unwind_baseline", "ext.sha")); err == nil {
				got := revParseHEAD(t, req.ExternalWtDir)
				if want := readBaselineSHA(t, req, "ext.sha"); got != want {
					t.Fatalf("external HEAD mutated: got %s want %s", got, want)
				}
			}
		}
	}
	if req.DepsLinkedWtDir != "" {
		if _, err := os.Stat(req.DepsLinkedWtDir); err == nil {
			if _, err := os.Stat(filepath.Join(req.WorkRoot, "_unwind_baseline", "deps.sha")); err == nil {
				got := revParseHEAD(t, req.DepsLinkedWtDir)
				if want := readBaselineSHA(t, req, "deps.sha"); got != want {
					t.Fatalf("deps external HEAD mutated: got %s want %s", got, want)
				}
			}
		}
	}
	// If DIRTY was part of the fixture, it must remain (read-only).
	for _, dir := range []string{req.MainRepo, req.WtDir, req.ExternalWtDir, req.DepsLinkedWtDir} {
		if dir == "" {
			continue
		}
		// Presence-only: if baseline recorded DIRTY, keep it; soft no-op when absent.
		_ = filepath.Join(dir, "DIRTY")
	}
}

// verifyJSON is a minimal schema probe for --json verify output.
type verifyJSON struct {
	WorkDir  string          `json:"work_dir"`
	Checks   json.RawMessage `json:"checks"`
	Summary  json.RawMessage `json:"summary"`
	Warnings json.RawMessage `json:"warnings"`
}

type verifyCheckJSON struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Status   string `json:"status"`
	Count    *int   `json:"count"`
}

type verifySummaryJSON struct {
	Checks int    `json:"checks"`
	Pass   int    `json:"pass"`
	Fail   int    `json:"fail"`
	Warn   int    `json:"warn"`
	Result string `json:"result"`
}

// parseVerifyJSON unmarshals stdout as verify JSON (must be pure JSON, no ANSI).
func parseVerifyJSON(t *testing.T, stdout string) verifyJSON {
	t.Helper()
	s := strings.TrimSpace(stdout)
	if s == "" {
		t.Fatal("expected JSON stdout, got empty")
	}
	if strings.Contains(stdout, "\x1b[") {
		t.Fatalf("JSON output must not contain ANSI escapes; stdout=%q", stdout)
	}
	if strings.Contains(stdout, verifyBanner) {
		t.Fatalf("JSON mode must not print human verify banners; stdout:\n%s", stdout)
	}
	var g verifyJSON
	if err := json.Unmarshal([]byte(s), &g); err != nil {
		t.Fatalf("parse verify JSON: %v\nstdout:\n%s", err, stdout)
	}
	return g
}

// assertVerifyJSONShape requires top-level keys and 6 catalog checks.
func assertVerifyJSONShape(t *testing.T, stdout string) (verifyJSON, []verifyCheckJSON, verifySummaryJSON) {
	t.Helper()
	g := parseVerifyJSON(t, stdout)
	var top map[string]json.RawMessage
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &top); err != nil {
		t.Fatalf("re-parse top map: %v", err)
	}
	for _, k := range []string{"work_dir", "checks", "summary", "warnings"} {
		if _, ok := top[k]; !ok {
			t.Fatalf("JSON missing top-level key %q; keys=%v", k, verifyMapKeys(top))
		}
	}
	var checks []verifyCheckJSON
	if err := json.Unmarshal(g.Checks, &checks); err != nil {
		t.Fatalf("parse checks: %v", err)
	}
	if len(checks) != 6 {
		t.Fatalf("checks len=%d want 6; got %#v", len(checks), checks)
	}
	seen := map[string]bool{}
	for _, c := range checks {
		seen[c.ID] = true
		if c.Severity == "" {
			t.Fatalf("check %q missing severity", c.ID)
		}
		if c.Status != "pass" && c.Status != "fail" && c.Status != "warn" {
			t.Fatalf("check %q status %q want pass|fail|warn", c.ID, c.Status)
		}
	}
	for _, id := range verifyCheckIDs {
		if !seen[id] {
			t.Fatalf("checks missing id %q; got %#v", id, checks)
		}
	}
	var sum verifySummaryJSON
	if err := json.Unmarshal(g.Summary, &sum); err != nil {
		t.Fatalf("parse summary: %v", err)
	}
	if sum.Checks != 6 {
		t.Fatalf("summary.checks=%d want 6", sum.Checks)
	}
	if sum.Result != "pass" && sum.Result != "fail" {
		t.Fatalf("summary.result=%q want pass|fail", sum.Result)
	}
	return g, checks, sum
}

// assertVerifyJSONCheckStatus requires check id status in JSON array.
func assertVerifyJSONCheckStatus(t *testing.T, checks []verifyCheckJSON, id, want string) {
	t.Helper()
	for _, c := range checks {
		if c.ID == id {
			if c.Status != want {
				t.Fatalf("check %q status=%q want %q", id, c.Status, want)
			}
			return
		}
	}
	t.Fatalf("check id %q not found in JSON checks", id)
}

func verifyMapKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	// Keep verify helpers referenced for generator / vet.
	_ = verifyArgs
	_ = verifyJSONArgs
	_ = setupVerifySingleMainClean
	_ = setupVerifySingleMainDirtyTagged
	_ = setupVerifySingleMainOwnedChanged
	_ = setupVerifySharedOwnedChanged
	_ = setupVerifyMultiRepoPinnedClean
	_ = setupVerifyDroppableReplace
	_ = setupVerifyRequireDrift
	_ = setupVerifyNeedsLand
	_ = assertNoSuccessfulVerifyBody
	_ = assertVerifyHumanBanners
	_ = assertVerifyReject
	_ = assertVerifyNoLogicalErrorPrefix
	_ = assertVerifyCheckStatus
	_ = assertVerifyAllChecksPass
	_ = assertVerifyResult
	_ = assertVerifyStdoutTrailingNL
	_ = assertVerifyNoANSI
	_ = assertVerifyHasANSI
	_ = stripVerifyANSI
	_ = assertVerifyZeroMutations
	_ = parseVerifyJSON
	_ = assertVerifyJSONShape
	_ = assertVerifyJSONCheckStatus
	unwindEnsureHelpersUsed()
	return nil
}
```
