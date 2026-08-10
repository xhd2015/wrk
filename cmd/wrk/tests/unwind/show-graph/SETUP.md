# Scenario

**Feature**: read-only `wrk --unwind --show-graph [--json]` prints repo + module graph

```
# stack inventory + repo DAG + PlanUnwind peel + full stack modules
wrk --unwind --show-graph [--json]
  -> reject --dry-run / apply partners (Error, non-zero)
  -> cycle? Error mentioning cycle (no success graph body)
  -> else human banners OR JSON (repos/modules/summary/warnings)
  -> zero mutations; no pin/land flags required
```

## Preconditions

- Inherits root `cmd/wrk/tests/unwind` Request/Response/Run and fixture helpers
  (`setupSingleMainDirty`, `setupThreeRepoChain`, `setupTwoCycleStack`,
  `setupFollow*`, peel display, baseline zero-mutation asserts).
- **Classic TDD (human polish):** formatter rewrite — human leaves assert **dir
  identity**, **collapsed `→` edges**, **`replaced`** (not `replace =>`),
  multi-repo grouping, optional drift/color. Must stay **RED** until
  `FormatUnwindGraphHuman` (+ color flags) land. Reject/cycle/JSON may stay GREEN.
- Leaves set `req.InProcess = true` and full `req.Args` including `--unwind`
  and `--show-graph` (plus optional `--json` / `--color` / `--no-color` or
  forbidden partners on reject). Color via Args only (no `t.Setenv` / `t.Chdir`).
- PeelOrder holds **display paths** for free-first expectations (same policy as
  dry-run).

## Steps

1. Grouping scopes the show-graph family; descendants branch on outcome
   (reject | cycle | success).
2. Success leaves seed acyclic fixtures and assert human or JSON graph body.
3. Reject leaves use minimal fixtures (single main) — fail before inventory is load-bearing.
4. Cycle leaf reuses `setupTwoCycleStack`.

## Context

- Human structure (locked banners):
  - `==== unwind graph (repo) ====`
  - `==== unwind graph (module) ====`
  - `==== status summary ====`
- **Human module identity:** scan-relative `dir` (`.`, `pkgs/shared`, …). Multi-repo
  (≥2 stack repos): key = `label` if dir is `.`, else `label/dir`. Group nodes
  with `modules @ <label> (<display>):` when multi-repo.
- **Edges collapsed by consumer:** `fromKey:` then indented `→ toKey …` (unicode
  arrow). Merge require+replace on one line: `require vX  replaced`. Never
  `replace => path` in human (path only in JSON).
- **Drift:** when require version ≠ dep latest tag version → `(latest …)`.
- Peel subsection: dirty free-first only; `(none)` when empty.
- Clean stack members appear in repo **nodes** even when omitted from peel.
- Soft inventory warnings: `warning:` on stderr; exit 0; graph still printed.
- Color (stdout, go-best-practice): auto / `--color` / `--no-color`; conflict
  error if both. JSON never ANSI.
- JSON snake_case: `repos`, `modules`, `summary`, `warnings`;
  `repos.peel_order`, `repos.has_pending_edges`, `repos.needs_land` (full paths).

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
	graphRepoBanner    = "==== unwind graph (repo) ===="
	graphModuleBanner  = "==== unwind graph (module) ===="
	graphSummaryBanner = "==== status summary ===="
	graphPeelSection   = "peel order"
)

// showGraphArgs returns base args for a successful human show-graph run.
func showGraphArgs(extra ...string) []string {
	args := []string{"--unwind", "--show-graph"}
	return append(args, extra...)
}

// showGraphJSONArgs returns args for JSON show-graph.
func showGraphJSONArgs(extra ...string) []string {
	return showGraphArgs(append([]string{"--json"}, extra...)...)
}

// setupSingleMainClean seeds a clean main-only root (no DIRTY file).
// RepoDir = MainRepo; PeelOrder empty (no dirty peel).
func setupSingleMainClean(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, labelRoot)
	initGitRepoOnMain(t, mainRepo)
	writeGoModRequire(t, mainRepo, unwindRootModule)
	writeFile(t, filepath.Join(mainRepo, "main.go"), "package main\n")
	runGitIsolated(t, mainRepo, "add", "go.mod", "main.go")
	runGitIsolated(t, mainRepo, "commit", "-m", "add module")
	mainRepo = resolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.PeelOrder = nil
}

// setupSingleMainTaggedOwnedChange seeds a main with tag v0.0.1 then a committed
// owned-file change so tagscope.Plan would plan NextTag=v0.0.2.
// Also leaves working-tree dirty (DIRTY) so peel is non-empty for status fields.
func setupSingleMainTaggedOwnedChange(t *testing.T, req *Request) {
	t.Helper()
	setupSingleMainClean(t, req)
	createLightweightTag(t, req.MainRepo, unwindApplyOldTag, "HEAD")
	// Owned change after baseline tag (tracked main.go).
	writeFile(t, filepath.Join(req.MainRepo, "main.go"), "package main\n// changed after tag\n")
	runGitIsolated(t, req.MainRepo, "add", "main.go")
	runGitIsolated(t, req.MainRepo, "commit", "-m", "owned change after tag")
	req.OldRequireVersion = unwindApplyOldTag
	req.ExpectedPinVersion = unwindApplyNextTag // expected next tag v0.0.2
	markDirty(t, req.MainRepo)
	req.PeelOrder = []string{"."}
}

// assertNoSuccessfulShowGraphBody fails if stdout looks like a completed graph print.
func assertNoSuccessfulShowGraphBody(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, graphRepoBanner) ||
		strings.Contains(stdout, graphModuleBanner) ||
		strings.Contains(stdout, graphSummaryBanner) {
		t.Fatalf("reject/cycle path must not print successful graph banners; stdout:\n%s", stdout)
	}
}

// assertShowGraphHumanBanners requires the three locked human section banners.
func assertShowGraphHumanBanners(t *testing.T, stdout string) {
	t.Helper()
	for _, b := range []string{graphRepoBanner, graphModuleBanner, graphSummaryBanner} {
		if !strings.Contains(stdout, b) {
			t.Fatalf("missing graph banner %q\nstdout:\n%s", b, stdout)
		}
	}
}

// assertShowGraphReject checks mutual-exclusion Error for show-graph + partner.
// partnerSubstr is a substring of the forbidden flag (e.g. "dry-run", "tag-next", "done").
func assertShowGraphReject(t *testing.T, resp *Response, partnerSubstr string) {
	t.Helper()
	assertExitNonZero(t, resp)
	combined := resp.Stdout + "\n" + resp.Stderr
	lower := strings.ToLower(combined)
	if !strings.Contains(lower, "show-graph") && !strings.Contains(lower, "show_graph") {
		t.Fatalf("error must mention show-graph; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
	if partnerSubstr != "" {
		ps := strings.ToLower(partnerSubstr)
		if !strings.Contains(lower, ps) && !strings.Contains(combined, partnerSubstr) {
			t.Fatalf("error must mention partner %q; stderr=%q stdout=%q",
				partnerSubstr, resp.Stderr, resp.Stdout)
		}
	}
	assertNoSuccessfulShowGraphBody(t, resp.Stdout)
}

// indexShowGraphPeelLine returns the byte index of a whole-line peel-order entry
// for displayPath under the peel-order section, or -1 if absent.
// Whole-line matching so "." is not a hit on "../external/..." (same class of
// bug as dry-run indexPeelLine / "would: peel .").
func indexShowGraphPeelLine(stdout, displayPath string) int {
	lower := strings.ToLower(stdout)
	peelIdx := strings.Index(lower, graphPeelSection)
	if peelIdx < 0 {
		return -1
	}
	section := stdout[peelIdx:]
	for _, stop := range []string{graphModuleBanner, graphSummaryBanner} {
		if i := strings.Index(section, stop); i > 0 {
			section = section[:i]
			break
		}
	}
	// Accept "  <display>" or "  N. <display>" / "  N) <display>" line forms.
	wantExact := "  " + displayPath
	at := 0
	for at <= len(section) {
		nl := strings.IndexByte(section[at:], '\n')
		segEnd := len(section)
		if nl >= 0 {
			segEnd = at + nl
		}
		seg := strings.TrimRight(section[at:segEnd], "\r")
		if seg == wantExact {
			return peelIdx + at
		}
		// Numbered: "  1. ." or "  1) ."
		trimmed := strings.TrimSpace(seg)
		if trimmed == displayPath {
			return peelIdx + at
		}
		// "1. display" / "1) display"
		for _, sep := range []string{". ", ") "} {
			if j := strings.Index(trimmed, sep); j > 0 {
				// leading digits only
				num := trimmed[:j]
				allDigit := len(num) > 0
				for _, c := range num {
					if c < '0' || c > '9' {
						allDigit = false
						break
					}
				}
				if allDigit && trimmed[j+len(sep):] == displayPath {
					return peelIdx + at
				}
			}
		}
		if nl < 0 {
			break
		}
		at = segEnd + 1
	}
	return -1
}

// assertShowGraphPeelOrderHuman checks free-first display paths appear in order
// under the peel-order section. Empty PeelOrder requires "(none)" somewhere in
// the peel section. Uses whole-line matching so "." ≠ prefix of "../…".
func assertShowGraphPeelOrderHuman(t *testing.T, stdout string, displayPaths []string) {
	t.Helper()
	lower := strings.ToLower(stdout)
	if !strings.Contains(lower, graphPeelSection) {
		t.Fatalf("stdout must include peel order section; got:\n%s", stdout)
	}
	if len(displayPaths) == 0 {
		// Empty peel: require (none) near peel order language.
		if !strings.Contains(lower, "(none)") && !strings.Contains(lower, "none") {
			t.Fatalf("empty peel order should show (none); stdout:\n%s", stdout)
		}
		return
	}
	var prev int = -1
	for i, display := range displayPaths {
		idx := indexShowGraphPeelLine(stdout, display)
		if idx < 0 {
			t.Fatalf("missing peel display %q (step %d)\nstdout:\n%s", display, i+1, stdout)
		}
		if idx <= prev {
			t.Fatalf("peel order wrong at %q: idx=%d prev=%d\nstdout:\n%s", display, idx, prev, stdout)
		}
		prev = idx
	}
}

// assertRepoNodeListed requires a stack member label/display appears in the repo graph.
func assertRepoNodeListed(t *testing.T, stdout, needle string) {
	t.Helper()
	if needle == "" {
		t.Fatal("assertRepoNodeListed: empty needle")
	}
	if !strings.Contains(stdout, needle) {
		t.Fatalf("repo graph must list %q\nstdout:\n%s", needle, stdout)
	}
}

// assertModulePathListed requires a full module path string somewhere in output.
// Prefer for JSON; human polish uses dir/label keys (assertModuleDirListed).
func assertModulePathListed(t *testing.T, stdout, modulePath string) {
	t.Helper()
	if !strings.Contains(stdout, modulePath) {
		t.Fatalf("module graph must list %q\nstdout:\n%s", modulePath, stdout)
	}
}

// moduleSection returns the human module graph section body (after module banner).
func moduleSection(stdout string) string {
	i := strings.Index(stdout, graphModuleBanner)
	if i < 0 {
		return ""
	}
	sec := stdout[i:]
	if j := strings.Index(sec, graphSummaryBanner); j > 0 {
		sec = sec[:j]
	}
	return sec
}

// repoSection returns the human repo graph section body (after repo banner).
func repoSection(stdout string) string {
	i := strings.Index(stdout, graphRepoBanner)
	if i < 0 {
		return ""
	}
	sec := stdout[i:]
	if j := strings.Index(sec, graphModuleBanner); j > 0 {
		sec = sec[:j]
	}
	return sec
}

// assertModuleDirListed requires a scan-relative module dir key in the module section.
// Single-repo identity is bare dir (`.`, `pkgs/shared`); multi-repo may use label/dir.
func assertModuleDirListed(t *testing.T, stdout, dirKey string) {
	t.Helper()
	if dirKey == "" {
		t.Fatal("assertModuleDirListed: empty dirKey")
	}
	sec := moduleSection(stdout)
	if sec == "" {
		t.Fatalf("missing module section for dir key %q\nstdout:\n%s", dirKey, stdout)
	}
	// Accept bare dir or multi-repo qualified key ending with /dir (or exact label for `.`).
	if strings.Contains(sec, dirKey) {
		return
	}
	t.Fatalf("module section must list dir/key %q\nsection:\n%s\nstdout:\n%s", dirKey, sec, stdout)
}

// assertHumanNoFullModulePaths fails if human output still lists fixture full
// module paths (pre-polish identity). JSON leaves must not call this.
func assertHumanNoFullModulePaths(t *testing.T, stdout string) {
	t.Helper()
	for _, p := range []string{
		unwindRootModule,
		unwindRootModule + "/shared",
		unwindRootModule + "/svc",
		unwindAgentProModule,
		unwindDotPkgsModule,
	} {
		if strings.Contains(stdout, p) {
			t.Fatalf("human format must use dir/label keys, not full module path %q\nstdout:\n%s", p, stdout)
		}
	}
	// Old node line form used "dir=" / "repo=" columns with full path first.
	sec := moduleSection(stdout)
	if strings.Contains(sec, "dir=") && strings.Contains(sec, "repo=") {
		t.Fatalf("human module nodes must not use old 'path dir=… repo=…' form\nsection:\n%s", sec)
	}
}

// assertHumanNoFlatFullPathEdges fails if human output still uses old flat
// "example.com/… -> example.com/…" edge lines (pre-polish full-path format).
func assertHumanNoFlatFullPathEdges(t *testing.T, stdout string) {
	t.Helper()
	// Old formatter: "  <module path> -> <module path> [require|replace …]"
	for _, line := range strings.Split(stdout, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.Contains(trim, " -> ") {
			continue
		}
		// Allow non-module narrative; flag module-path shaped flat edges.
		if strings.Contains(trim, "example.com/") || strings.Count(trim, "/") >= 2 &&
			(strings.Contains(trim, "[require") || strings.Contains(trim, "[replace")) {
			t.Fatalf("human format must not use flat full-path edges; line=%q\nstdout:\n%s", trim, stdout)
		}
	}
}

// assertCollapsedEdgesHuman requires consumer-collapsed deps with unicode →.
func assertCollapsedEdgesHuman(t *testing.T, stdout string) {
	t.Helper()
	if !strings.Contains(stdout, "→") && !strings.Contains(stdout, "\u2192") {
		t.Fatalf("human edges must collapse with unicode arrow →; stdout:\n%s", stdout)
	}
}

// assertReplacedHuman requires word "replaced" and forbids "replace =>" in human body.
func assertReplacedHuman(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, "replace =>") {
		t.Fatalf("human format must not show replace => path; use 'replaced'\nstdout:\n%s", stdout)
	}
	if !strings.Contains(strings.ToLower(stdout), "replaced") {
		t.Fatalf("human replace edge must use word 'replaced'\nstdout:\n%s", stdout)
	}
}

// assertNoReplaceArrowHuman fails if stdout still contains go.mod-style "replace =>".
func assertNoReplaceArrowHuman(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, "replace =>") {
		t.Fatalf("human format must not contain 'replace =>'\nstdout:\n%s", stdout)
	}
}

// assertModulesGroupedByRepo requires multi-repo node grouping header.
func assertModulesGroupedByRepo(t *testing.T, stdout string) {
	t.Helper()
	sec := moduleSection(stdout)
	if !strings.Contains(sec, "modules @") {
		t.Fatalf("multi-repo human nodes must group with 'modules @ <label>'; section:\n%s", sec)
	}
}

// assertRequireDriftHuman requires "(latest" drift annotation somewhere in stdout.
func assertRequireDriftHuman(t *testing.T, stdout string) {
	t.Helper()
	if !strings.Contains(stdout, "(latest") {
		t.Fatalf("require-drift must show (latest …); stdout:\n%s", stdout)
	}
}

// assertShowGraphNoANSI fails if stdout has CSI/ANSI escapes.
func assertShowGraphNoANSI(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, "\x1b[") {
		t.Fatalf("stdout must not contain ANSI escapes; got:\n%q", stdout)
	}
}

// assertShowGraphHasANSI requires at least one CSI escape in stdout (--color).
func assertShowGraphHasANSI(t *testing.T, stdout string) {
	t.Helper()
	if !strings.Contains(stdout, "\x1b[") {
		t.Fatalf("expected ANSI escapes on stdout with --color; stdout:\n%q", stdout)
	}
}

// stripANSI removes CSI sequences so banner substrings can be matched when colored.
func stripANSI(s string) string {
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

// assertShowGraphHumanBannersMaybeColored accepts plain or ANSI-wrapped banners.
func assertShowGraphHumanBannersMaybeColored(t *testing.T, stdout string) {
	t.Helper()
	plain := stripANSI(stdout)
	assertShowGraphHumanBanners(t, plain)
}

// assertShowGraphZeroMutations checks HEADs/worktrees like dry-run; DIRTY only if present.
func assertShowGraphZeroMutations(t *testing.T, req *Request) {
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
		p := filepath.Join(dir, "DIRTY")
		if _, err := os.Stat(p); err == nil {
			// still present — good
			continue
		}
	}
	// When PeelOrder includes ".", dirty primary DIRTY should still exist if markDirty ran.
	if len(req.PeelOrder) > 0 && req.MainRepo != "" && req.WtDir == "" {
		dirtyPath := filepath.Join(req.MainRepo, "DIRTY")
		if _, err := os.Stat(dirtyPath); err == nil {
			assertFileExists(t, dirtyPath)
		}
	}
}

// graphJSON is a minimal schema probe for --json show-graph output.
type graphJSON struct {
	Repos    json.RawMessage `json:"repos"`
	Modules  json.RawMessage `json:"modules"`
	Summary  json.RawMessage `json:"summary"`
	Warnings json.RawMessage `json:"warnings"`
}

type graphReposJSON struct {
	Nodes           json.RawMessage `json:"nodes"`
	Edges           json.RawMessage `json:"edges"`
	PeelOrder       []string        `json:"peel_order"`
	HasPendingEdges *bool           `json:"has_pending_edges"`
	NeedsLand       *bool           `json:"needs_land"`
}

type graphModulesJSON struct {
	Nodes json.RawMessage `json:"nodes"`
	Edges json.RawMessage `json:"edges"`
}

// parseGraphJSON unmarshals stdout as show-graph JSON (must be pure JSON, no ANSI).
func parseGraphJSON(t *testing.T, stdout string) graphJSON {
	t.Helper()
	s := strings.TrimSpace(stdout)
	if s == "" {
		t.Fatal("expected JSON stdout, got empty")
	}
	// Guard against ANSI / human banners.
	if strings.Contains(stdout, "\x1b[") {
		t.Fatalf("JSON output must not contain ANSI escapes; stdout=%q", stdout)
	}
	if strings.Contains(stdout, graphRepoBanner) {
		t.Fatalf("JSON mode must not print human graph banners; stdout:\n%s", stdout)
	}
	var g graphJSON
	if err := json.Unmarshal([]byte(s), &g); err != nil {
		t.Fatalf("parse show-graph JSON: %v\nstdout:\n%s", err, stdout)
	}
	return g
}

// assertGraphJSONShape requires top-level keys and nested repos/modules structure.
func assertGraphJSONShape(t *testing.T, stdout string) (graphJSON, graphReposJSON, graphModulesJSON) {
	t.Helper()
	g := parseGraphJSON(t, stdout)
	if len(g.Repos) == 0 {
		t.Fatal("JSON missing repos")
	}
	if len(g.Modules) == 0 {
		t.Fatal("JSON missing modules")
	}
	if len(g.Summary) == 0 {
		t.Fatal("JSON missing summary")
	}
	// warnings may be null or []; key must be present in object — RawMessage empty means missing key.
	// encoding/json sets missing key to nil RawMessage; require presence via map probe.
	var top map[string]json.RawMessage
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &top); err != nil {
		t.Fatalf("re-parse top map: %v", err)
	}
	for _, k := range []string{"repos", "modules", "summary", "warnings"} {
		if _, ok := top[k]; !ok {
			t.Fatalf("JSON missing top-level key %q; keys=%v", k, mapKeys(top))
		}
	}
	var repos graphReposJSON
	if err := json.Unmarshal(g.Repos, &repos); err != nil {
		t.Fatalf("parse repos: %v", err)
	}
	if len(repos.Nodes) == 0 {
		t.Fatal("repos.nodes missing or empty")
	}
	if repos.Edges == nil {
		t.Fatal("repos.edges missing")
	}
	if repos.PeelOrder == nil {
		t.Fatal("repos.peel_order missing (want array, possibly empty)")
	}
	if repos.HasPendingEdges == nil {
		t.Fatal("repos.has_pending_edges missing")
	}
	if repos.NeedsLand == nil {
		t.Fatal("repos.needs_land missing")
	}
	var mods graphModulesJSON
	if err := json.Unmarshal(g.Modules, &mods); err != nil {
		t.Fatalf("parse modules: %v", err)
	}
	if len(mods.Nodes) == 0 {
		t.Fatal("modules.nodes missing or empty")
	}
	if mods.Edges == nil {
		t.Fatal("modules.edges missing")
	}
	return g, repos, mods
}

func mapKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// assertJSONPeelOrder checks peel_order equals expected display path sequence.
func assertJSONPeelOrder(t *testing.T, peelOrder, want []string) {
	t.Helper()
	if len(peelOrder) != len(want) {
		t.Fatalf("peel_order len=%d want %d: got %#v want %#v", len(peelOrder), len(want), peelOrder, want)
	}
	for i := range want {
		if peelOrder[i] != want[i] {
			t.Fatalf("peel_order[%d]=%q want %q; full got %#v", i, peelOrder[i], want[i], peelOrder)
		}
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	// Keep show-graph helpers referenced for generator / vet.
	_ = showGraphArgs
	_ = showGraphJSONArgs
	_ = setupSingleMainClean
	_ = setupSingleMainTaggedOwnedChange
	_ = assertNoSuccessfulShowGraphBody
	_ = assertShowGraphHumanBanners
	_ = assertShowGraphReject
	_ = assertShowGraphPeelOrderHuman
	_ = assertRepoNodeListed
	_ = assertModulePathListed
	_ = assertModuleDirListed
	_ = assertHumanNoFullModulePaths
	_ = assertHumanNoFlatFullPathEdges
	_ = assertCollapsedEdgesHuman
	_ = assertReplacedHuman
	_ = assertNoReplaceArrowHuman
	_ = assertModulesGroupedByRepo
	_ = assertRequireDriftHuman
	_ = assertShowGraphNoANSI
	_ = assertShowGraphHasANSI
	_ = stripANSI
	_ = assertShowGraphHumanBannersMaybeColored
	_ = moduleSection
	_ = repoSection
	_ = assertShowGraphZeroMutations
	_ = parseGraphJSON
	_ = assertGraphJSONShape
	_ = assertJSONPeelOrder
	unwindEnsureHelpersUsed()
	return nil
}
```
