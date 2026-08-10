package wrkcli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

// UnwindGraphReport is the read-only unwind stack graph (repo + module + summary).
type UnwindGraphReport struct {
	WorkDir  string
	Repos    UnwindGraphRepos
	Modules  UnwindGraphModules
	Summary  UnwindGraphSummary
	Warnings []string
}

// UnwindGraphRepos is the stack-repo layer of the show-graph report.
type UnwindGraphRepos struct {
	Nodes           []UnwindGraphRepoNode
	Edges           []RepoEdge
	PeelOrder       []string // display paths, free-first
	HasPendingEdges bool
	NeedsLand       bool
}

// UnwindGraphRepoNode is one stack checkout in the repo graph.
type UnwindGraphRepoNode struct {
	Display   string `json:"display"`
	Label     string `json:"label"`
	Linked    bool   `json:"linked"`
	Dirty     bool   `json:"dirty"`
	Branch    string `json:"branch,omitempty"`
	HeadShort string `json:"head_short,omitempty"`
	PeelIndex int    `json:"peel_index,omitempty"` // 1-based; 0 = not in peel
	Land      bool   `json:"land,omitempty"`
}

// UnwindGraphModules is the full-stack module layer.
type UnwindGraphModules struct {
	Nodes []UnwindGraphModuleNode
	Edges []UnwindGraphModuleEdge
}

// UnwindGraphModuleNode is one go.mod under the unwind stack.
type UnwindGraphModuleNode struct {
	Path         string `json:"path"`
	Dir          string `json:"dir"`
	RepoLabel    string `json:"repo_label"`
	LatestTag    string `json:"latest_tag,omitempty"`
	NextTag      string `json:"next_tag,omitempty"`
	OwnedChanged bool   `json:"owned_changed,omitempty"`
	SkipReason   string `json:"skip_reason,omitempty"`
}

// UnwindGraphModuleEdge is a require or replace edge among stack-owned modules.
type UnwindGraphModuleEdge struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Kind    string `json:"kind"` // require | replace
	Version string `json:"version,omitempty"`
	NewPath string `json:"new_path,omitempty"`
}

// UnwindGraphSummary aggregates counts and an optional apply hint.
type UnwindGraphSummary struct {
	Repos        int    `json:"repos"`
	DirtyRepos   int    `json:"dirty_repos"`
	Modules      int    `json:"modules"`
	RepoEdges    int    `json:"repo_edges"`
	ModuleEdges  int    `json:"module_edges"`
	PeelSteps    int    `json:"peel_steps"`
	Cycle        string `json:"cycle"`
	ApplyHint    string `json:"apply_hint,omitempty"`
}

// BuildUnwindGraphReport collects inventory, repo DAG, peel plan, module graph,
// and tagscope status for a read-only show-graph inspect. Does not mutate.
func BuildUnwindGraphReport(workDir string) (*UnwindGraphReport, error) {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve cwd: %w", err)
	}
	inv, err := CollectStackInventory(cwd)
	if err != nil {
		return nil, err
	}
	members := inv.Members
	edges, err := BuildRepoDAG(members)
	if err != nil {
		return nil, err
	}
	edges = mergeRepoEdges(edges, inv.SyntheticEdges)
	plan, err := PlanUnwind(members, edges)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		plan = &UnwindPlan{}
	}

	byLabel := pickPeelMembersByLabel(members)
	peelIndexByLabel := make(map[string]int, len(plan.PeelOrder))
	peelDisplays := make([]string, 0, len(plan.PeelOrder))
	for i, label := range plan.PeelOrder {
		peelIndexByLabel[label] = i + 1
		display := label
		if m, ok := byLabel[label]; ok {
			display = peelDisplayPath(cwd, m.Path)
		}
		peelDisplays = append(peelDisplays, display)
	}

	// Prefer peel member for display when multiple checkouts share a label.
	// List every inventory member so nested clean leaves stay visible.
	repoNodes := make([]UnwindGraphRepoNode, 0, len(members))
	dirtyRepos := 0
	seenDirtyLabel := make(map[string]struct{})
	for _, m := range members {
		display := peelDisplayPath(cwd, m.Path)
		branch, _ := gitOutputDir(m.Path, "rev-parse", "--abbrev-ref", "HEAD")
		headShort, _ := shortHEAD(m.Path)
		peelIdx := peelIndexByLabel[m.Label]
		// Only the chosen peel checkout gets peel# when label is dirty.
		if peelIdx > 0 {
			if chosen, ok := byLabel[m.Label]; ok && storage.NormalizePath(chosen.Path) != storage.NormalizePath(m.Path) {
				peelIdx = 0
			}
		}
		land := m.Linked && peelIdx > 0
		node := UnwindGraphRepoNode{
			Display:   display,
			Label:     m.Label,
			Linked:    m.Linked,
			Dirty:     m.Dirty,
			Branch:    strings.TrimSpace(branch),
			HeadShort: headShort,
			PeelIndex: peelIdx,
			Land:      land,
		}
		repoNodes = append(repoNodes, node)
		if m.Dirty {
			if _, ok := seenDirtyLabel[m.Label]; !ok {
				seenDirtyLabel[m.Label] = struct{}{}
				dirtyRepos++
			}
		}
	}
	// Stable: peel-index (in peel first, ascending), then label, then display.
	sort.SliceStable(repoNodes, func(i, j int) bool {
		ai, aj := repoNodes[i].PeelIndex, repoNodes[j].PeelIndex
		// 0 sorts after any positive peel index.
		if ai == 0 {
			ai = 1 << 30
		}
		if aj == 0 {
			aj = 1 << 30
		}
		if ai != aj {
			return ai < aj
		}
		if repoNodes[i].Label != repoNodes[j].Label {
			return repoNodes[i].Label < repoNodes[j].Label
		}
		return repoNodes[i].Display < repoNodes[j].Display
	})

	modNodes, modEdges, err := buildUnwindModuleGraph(members, byLabel)
	if err != nil {
		return nil, err
	}
	attachTagScopeToModules(modNodes, members, nil)

	applyHint := buildApplyHint(plan)
	summary := UnwindGraphSummary{
		Repos:       len(uniqueRepoLabels(members)),
		DirtyRepos:  dirtyRepos,
		Modules:     len(modNodes),
		RepoEdges:   len(edges),
		ModuleEdges: len(modEdges),
		PeelSteps:   len(peelDisplays),
		Cycle:       "none",
		ApplyHint:   applyHint,
	}

	return &UnwindGraphReport{
		WorkDir: cwd,
		Repos: UnwindGraphRepos{
			Nodes:           repoNodes,
			Edges:           edges,
			PeelOrder:       peelDisplays,
			HasPendingEdges: plan.HasPendingEdges,
			NeedsLand:       plan.NeedsLand,
		},
		Modules: UnwindGraphModules{
			Nodes: modNodes,
			Edges: modEdges,
		},
		Summary:  summary,
		Warnings: append([]string(nil), inv.Warnings...),
	}, nil
}

func uniqueRepoLabels(members []StackMember) []string {
	seen := make(map[string]struct{}, len(members))
	var labels []string
	for _, m := range members {
		if _, ok := seen[m.Label]; ok {
			continue
		}
		seen[m.Label] = struct{}{}
		labels = append(labels, m.Label)
	}
	return labels
}

func buildApplyHint(plan *UnwindPlan) string {
	if plan == nil {
		return ""
	}
	var parts []string
	if plan.NeedsLand {
		parts = append(parts, "--merge-back")
	}
	if plan.HasPendingEdges {
		parts = append(parts, "--tag-next", "--push")
	}
	if len(parts) == 0 {
		return ""
	}
	return "apply would need " + strings.Join(parts, " ")
}

// stackModule is a scanned module with owning stack member.
type stackModule struct {
	Path      string
	Dir       string // relative to checkout Path
	RepoLabel string
	MainRepo  string
	Requires  []scan.ModuleRequire
	Replaces  []scan.ModuleReplace
}

func buildUnwindModuleGraph(members []StackMember, byLabel map[string]StackMember) ([]UnwindGraphModuleNode, []UnwindGraphModuleEdge, error) {
	_ = byLabel
	// Prefer linked/dirty checkout when the same main appears twice: scan each
	// inventory path but dedupe modules by module path (first wins after sort).
	type modKey struct {
		path string
	}
	modByPath := make(map[string]stackModule)
	for _, m := range members {
		scanned, err := scan.Scan(m.Path, scan.Options{})
		if err != nil {
			return nil, nil, err
		}
		for _, sm := range scanned {
			if sm.Path == "" {
				continue
			}
			// Tolerant require fallback when scan dropped invalid versions.
			requires := sm.Requires
			if len(requires) == 0 {
				modDir := m.Path
				if sm.Dir != "" && sm.Dir != "." {
					modDir = filepath.Join(m.Path, filepath.FromSlash(sm.Dir))
				}
				if reqs, err := parseRequiresTolerant(filepath.Join(modDir, "go.mod")); err == nil {
					for _, r := range reqs {
						requires = append(requires, scan.ModuleRequire{Path: r.Path, Version: r.Version})
					}
				}
			}
			rec := stackModule{
				Path:      sm.Path,
				Dir:       sm.Dir,
				RepoLabel: m.Label,
				MainRepo:  m.MainRepo,
				Requires:  requires,
				Replaces:  sm.Replaces,
			}
			if prev, ok := modByPath[sm.Path]; ok {
				// Prefer member that owns more edges / linked dirty; keep first otherwise.
				if prev.RepoLabel == rec.RepoLabel {
					continue
				}
				// Prefer label of peel-chosen member if either matches.
				continue
			}
			modByPath[sm.Path] = rec
		}
	}

	// Also ensure replace targets that resolve into stack checkouts are owned
	// (module path from scan of dep already present if dep is in inventory).
	paths := make([]string, 0, len(modByPath))
	for p := range modByPath {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	nodes := make([]UnwindGraphModuleNode, 0, len(paths))
	for _, p := range paths {
		rec := modByPath[p]
		dir := rec.Dir
		if dir == "" {
			dir = "."
		}
		nodes = append(nodes, UnwindGraphModuleNode{
			Path:      rec.Path,
			Dir:       dir,
			RepoLabel: rec.RepoLabel,
		})
	}

	edgeKey := make(map[string]UnwindGraphModuleEdge)
	addEdge := func(e UnwindGraphModuleEdge) {
		if e.From == "" || e.To == "" || e.From == e.To {
			return
		}
		if _, ok := modByPath[e.To]; !ok {
			return
		}
		key := e.From + "\x00" + e.To + "\x00" + e.Kind + "\x00" + e.Version + "\x00" + e.NewPath
		edgeKey[key] = e
	}

	for _, p := range paths {
		rec := modByPath[p]
		for _, req := range rec.Requires {
			if req.Path == "" {
				continue
			}
			if _, ok := modByPath[req.Path]; !ok {
				continue
			}
			addEdge(UnwindGraphModuleEdge{
				From:    rec.Path,
				To:      req.Path,
				Kind:    "require",
				Version: req.Version,
			})
		}
		for _, repl := range rec.Replaces {
			// Edge to the replaced module path when it is stack-owned.
			target := repl.OldPath
			if target == "" {
				continue
			}
			if _, ok := modByPath[target]; !ok {
				// Module-path replace: NewPath may be a module path (not filesystem).
				if repl.NewPath != "" && !strings.HasPrefix(repl.NewPath, ".") && !filepath.IsAbs(repl.NewPath) {
					if _, ok2 := modByPath[repl.NewPath]; ok2 {
						target = repl.NewPath
					} else {
						continue
					}
				} else {
					continue
				}
			}
			addEdge(UnwindGraphModuleEdge{
				From:    rec.Path,
				To:      target,
				Kind:    "replace",
				NewPath: repl.NewPath,
			})
		}
	}

	edges := make([]UnwindGraphModuleEdge, 0, len(edgeKey))
	for _, e := range edgeKey {
		edges = append(edges, e)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		if edges[i].To != edges[j].To {
			return edges[i].To < edges[j].To
		}
		if edges[i].Kind != edges[j].Kind {
			return edges[i].Kind < edges[j].Kind
		}
		return edges[i].NewPath < edges[j].NewPath
	})
	return nodes, edges, nil
}

// tagScopePlanEntry caches one tagscope.Plan result for a checkout root.
type tagScopePlanEntry struct {
	ok   bool
	plan tagscope.ChangePlan
}

// tagScopePlanCache is an optional per-ApplyUnwind cache of tagscope.Plan by
// normalized checkout root. tagscope.Plan is multi-second on large monorepos;
// pinReady + cascade + split would otherwise re-run it for every peel.
type tagScopePlanCache map[string]tagScopePlanEntry

// attachTagScopeToModules fills latest/next/skip from tagscope.Plan per stack
// checkout. Prefers the peel-chosen Path (linked dirty worktree) so HEAD
// reflects owned changes still on the worktree, not only MainRepo tip.
// Soft-skips failures (leaves fields empty / skip_reason=unknown).
// When cache is non-nil, tagscope.Plan results are shared across calls (one
// ApplyUnwind run); pass nil for isolated one-shot use (show-graph, verify).
func attachTagScopeToModules(nodes []UnwindGraphModuleNode, members []StackMember, cache tagScopePlanCache) {
	// Peel-preferred checkout root per label (linked/dirty Path first).
	planRootByLabel := make(map[string]string, len(members))
	for label, m := range pickPeelMembersByLabel(members) {
		root := m.Path
		if root == "" {
			root = m.MainRepo
		}
		if root != "" {
			planRootByLabel[label] = storage.NormalizePath(root)
		}
	}
	for _, m := range members {
		if _, ok := planRootByLabel[m.Label]; ok {
			continue
		}
		root := m.Path
		if root == "" {
			root = m.MainRepo
		}
		if root != "" {
			planRootByLabel[m.Label] = storage.NormalizePath(root)
		}
	}
	// Local map when caller did not supply a shared ApplyUnwind cache.
	if cache == nil {
		cache = make(tagScopePlanCache)
	}

	getPlan := func(repoRoot string) (tagscope.ChangePlan, bool) {
		repoRoot = storage.NormalizePath(repoRoot)
		if c, ok := cache[repoRoot]; ok {
			return c.plan, c.ok
		}
		plan, _, err := tagscope.Plan(repoRoot, "HEAD")
		if err != nil {
			cache[repoRoot] = tagScopePlanEntry{ok: false}
			return tagscope.ChangePlan{}, false
		}
		cache[repoRoot] = tagScopePlanEntry{ok: true, plan: plan}
		return plan, true
	}

	for i := range nodes {
		repoRoot := planRootByLabel[nodes[i].RepoLabel]
		if repoRoot == "" {
			continue
		}
		plan, ok := getPlan(repoRoot)
		if !ok {
			nodes[i].SkipReason = "unknown"
			continue
		}
		// Match decision by path prefix of module dir when monorepo scopes exist.
		// Root scope (empty PathPrefix) is the default for fixtures.
		var best *tagscope.ScopeDecision
		modDir := nodes[i].Dir
		if modDir == "." {
			modDir = ""
		}
		if modDir != "" && !strings.HasSuffix(modDir, "/") {
			modDir += "/"
		}
		for j := range plan.Decisions {
			d := &plan.Decisions[j]
			pp := d.Scope.PathPrefix
			if pp == "" {
				if best == nil {
					best = d
				}
				continue
			}
			if strings.HasPrefix(modDir, pp) || strings.HasPrefix(nodes[i].Path, pp) {
				best = d
			}
		}
		if best == nil && len(plan.Decisions) > 0 {
			best = &plan.Decisions[0]
		}
		if best == nil {
			continue
		}
		nodes[i].LatestTag = best.LatestRelease
		nodes[i].NextTag = best.NextTag
		nodes[i].SkipReason = best.SkipReason
		if best.NextTag != "" {
			nodes[i].OwnedChanged = true
		}
	}
}

// FormatUnwindGraphHuman renders the human show-graph body (trailing newline).
// colorOn gates ANSI on stdout tokens (banners gray; dirty/replaced/drift orange;
// owned-changed/next green). Callers pass false for JSON / --no-color / non-TTY auto.
func FormatUnwindGraphHuman(report *UnwindGraphReport, colorOn bool) string {
	if report == nil {
		return ""
	}
	var b strings.Builder

	banner := func(s string) string {
		return paint(s, ansiGrey, colorOn)
	}

	// ---- repo section ----
	b.WriteString(banner("==== unwind graph (repo) ===="))
	b.WriteByte('\n')
	b.WriteString("nodes:\n")
	if len(report.Repos.Nodes) == 0 {
		b.WriteString("  (none)\n")
	} else {
		writeAlignedRepoNodes(&b, report.Repos.Nodes, colorOn)
	}
	b.WriteByte('\n')
	b.WriteString("edges (From depends on To):\n")
	if len(report.Repos.Edges) == 0 {
		b.WriteString("  (none)\n")
	} else {
		writeCollapsedRepoEdges(&b, report.Repos.Edges)
	}
	b.WriteByte('\n')
	b.WriteString("peel order (dirty, free-first):\n")
	if len(report.Repos.PeelOrder) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for _, p := range report.Repos.PeelOrder {
			fmt.Fprintf(&b, "  %s\n", p)
		}
	}
	b.WriteByte('\n')

	// ---- module section ----
	multiRepo := report.Summary.Repos >= 2
	// Prefer repo-node count when summary is zero but nodes present.
	if !multiRepo {
		labels := make(map[string]struct{})
		for _, n := range report.Repos.Nodes {
			labels[n.Label] = struct{}{}
		}
		for _, n := range report.Modules.Nodes {
			labels[n.RepoLabel] = struct{}{}
		}
		multiRepo = len(labels) >= 2
	}

	displayByLabel := make(map[string]string)
	for _, n := range report.Repos.Nodes {
		if n.Label == "" {
			continue
		}
		if _, ok := displayByLabel[n.Label]; !ok {
			displayByLabel[n.Label] = n.Display
		}
	}

	pathToNode := make(map[string]UnwindGraphModuleNode, len(report.Modules.Nodes))
	for _, n := range report.Modules.Nodes {
		pathToNode[n.Path] = n
	}

	b.WriteString(banner("==== unwind graph (module) ===="))
	b.WriteByte('\n')
	b.WriteString("nodes:\n")
	if len(report.Modules.Nodes) == 0 {
		b.WriteString("  (none)\n")
	} else if multiRepo {
		writeMultiRepoModuleNodes(&b, report.Modules.Nodes, displayByLabel, colorOn)
	} else {
		// Single-repo: identity is scan-relative dir only; aligned columns.
		writeAlignedModuleNodes(&b, sortedModuleNodesByDir(report.Modules.Nodes), "  ", colorOn)
	}
	b.WriteByte('\n')
	b.WriteString("edges (consumer → deps):\n")
	if len(report.Modules.Edges) == 0 {
		b.WriteString("  (none)\n")
	} else {
		writeCollapsedModuleEdges(&b, report.Modules.Edges, pathToNode, multiRepo, colorOn)
	}
	b.WriteByte('\n')

	// ---- summary ----
	b.WriteString(banner("==== status summary ===="))
	b.WriteByte('\n')
	s := report.Summary
	fmt.Fprintf(&b, "repos: %d  dirty: %d  modules: %d  repo-edges: %d  module-edges: %d  peel-steps: %d  cycle: %s\n",
		s.Repos, s.DirtyRepos, s.Modules, s.RepoEdges, s.ModuleEdges, s.PeelSteps, s.Cycle)
	if s.ApplyHint != "" {
		fmt.Fprintf(&b, "hint: %s\n", s.ApplyHint)
	}
	return b.String()
}

// writeAlignedRepoNodes prints path/label/kind/dirt/branch/head with padded columns.
func writeAlignedRepoNodes(b *strings.Builder, nodes []UnwindGraphRepoNode, colorOn bool) {
	type row struct {
		path, label, kind, dirt, branch, head, extra string
	}
	rows := make([]row, 0, len(nodes))
	pathW, labelW, kindW, dirtW, branchW, headW :=
		len("path"), len("label"), len("kind"), len("dirt"), len("branch"), len("head")
	for _, n := range nodes {
		kind := "main"
		if n.Linked {
			kind = "linked"
		}
		dirtPlain := "clean"
		dirt := dirtPlain
		if n.Dirty {
			dirtPlain = "dirty"
			dirt = paint("dirty", ansiOrange, colorOn)
		}
		var extraParts []string
		if n.PeelIndex > 0 {
			extraParts = append(extraParts, fmt.Sprintf("peel#%d", n.PeelIndex))
		}
		if n.Land {
			extraParts = append(extraParts, "land=yes")
		}
		r := row{
			path:   n.Display,
			label:  n.Label,
			kind:   kind,
			dirt:   dirt,
			branch: n.Branch,
			head:   n.HeadShort,
			extra:  strings.Join(extraParts, "  "),
		}
		rows = append(rows, r)
		if w := visibleLen(r.path); w > pathW {
			pathW = w
		}
		if w := visibleLen(r.label); w > labelW {
			labelW = w
		}
		if w := visibleLen(r.kind); w > kindW {
			kindW = w
		}
		if w := visibleLen(dirtPlain); w > dirtW {
			dirtW = w
		}
		if w := visibleLen(r.branch); w > branchW {
			branchW = w
		}
		if w := visibleLen(r.head); w > headW {
			headW = w
		}
	}
	hdr := "  " + padRightVisible("path", pathW) + "  " +
		padRightVisible("label", labelW) + "  " +
		padRightVisible("kind", kindW) + "  " +
		padRightVisible("dirt", dirtW) + "  " +
		padRightVisible("branch", branchW) + "  " +
		padRightVisible("head", headW)
	b.WriteString(paint(hdr, ansiGrey, colorOn))
	b.WriteByte('\n')
	for _, r := range rows {
		line := "  " + padRightVisible(r.path, pathW) + "  " +
			padRightVisible(r.label, labelW) + "  " +
			padRightVisible(r.kind, kindW) + "  " +
			padRightVisible(r.dirt, dirtW) + "  " +
			padRightVisible(r.branch, branchW) + "  " +
			padRightVisible(r.head, headW)
		if r.extra != "" {
			line += "  " + r.extra
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
}

// moduleDirKey returns the scan-relative dir identity (default ".").
func moduleDirKey(n UnwindGraphModuleNode) string {
	if n.Dir == "" {
		return "."
	}
	return n.Dir
}

// moduleHumanKey is the human edge/node key: bare dir (single-repo) or
// label / label/dir (multi-repo).
func moduleHumanKey(n UnwindGraphModuleNode, multiRepo bool) string {
	dir := moduleDirKey(n)
	if !multiRepo {
		return dir
	}
	if dir == "." {
		return n.RepoLabel
	}
	return n.RepoLabel + "/" + dir
}

func sortedModuleNodesByDir(nodes []UnwindGraphModuleNode) []UnwindGraphModuleNode {
	out := append([]UnwindGraphModuleNode(nil), nodes...)
	sort.SliceStable(out, func(i, j int) bool {
		di, dj := moduleDirKey(out[i]), moduleDirKey(out[j])
		if di != dj {
			return di < dj
		}
		return out[i].Path < out[j].Path
	})
	return out
}

// tagVersion extracts the trailing semver-ish segment from a full tag or version.
// Examples: "pkgs/log/v0.0.2" → "v0.0.2", "v0.0.279" → "v0.0.279", "0.0.1" → "0.0.1".
func tagVersion(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return ""
	}
	if i := strings.LastIndex(tag, "/"); i >= 0 && i+1 < len(tag) {
		tag = tag[i+1:]
	}
	return tag
}

// versionsMatch reports whether a require version and a latest tag refer to the
// same release (string equality after stripping path-style tag prefixes).
func versionsMatch(requireVer, latestTag string) bool {
	a := tagVersion(requireVer)
	b := tagVersion(latestTag)
	if a == "" || b == "" {
		return false
	}
	// Accept "v0.0.2" vs "0.0.2".
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")
	return a == b
}

// visibleLen returns rune length ignoring ANSI CSI sequences (…m).
func visibleLen(s string) int {
	n := 0
	inESC := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inESC {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				inESC = false
			}
			continue
		}
		if c == '\x1b' {
			inESC = true
			continue
		}
		n++
	}
	return n
}

func padRightVisible(s string, width int) string {
	pad := width - visibleLen(s)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(" ", pad)
}

// writeAlignedModuleNodes prints a dir/latest/status table with a header.
func writeAlignedModuleNodes(b *strings.Builder, nodes []UnwindGraphModuleNode, indent string, colorOn bool) {
	if len(nodes) == 0 {
		fmt.Fprintf(b, "%snodes:\n%s  (none)\n", indent, indent)
		return
	}
	type row struct {
		dir, latest, status string
	}
	rows := make([]row, 0, len(nodes))
	dirW, latestW := len("dir"), len("latest")
	for _, n := range nodes {
		dir := moduleDirKey(n)
		latest := tagVersion(n.LatestTag)
		if latest == "" && n.LatestTag != "" {
			latest = n.LatestTag
		}
		var statusParts []string
		if n.OwnedChanged || n.NextTag != "" {
			statusParts = append(statusParts, paint("owned-changed", ansiGreen, colorOn))
		}
		if n.NextTag != "" {
			statusParts = append(statusParts, paint("next="+tagVersion(n.NextTag), ansiGreen, colorOn))
		}
		if n.SkipReason != "" && n.NextTag == "" && n.LatestTag == "" {
			statusParts = append(statusParts, "skip="+n.SkipReason)
		}
		status := strings.Join(statusParts, "  ")
		rows = append(rows, row{dir: dir, latest: latest, status: status})
		if w := visibleLen(dir); w > dirW {
			dirW = w
		}
		if w := visibleLen(latest); w > latestW {
			latestW = w
		}
	}
	// Header (gray when color on).
	hdr := indent + padRightVisible("dir", dirW) + "  " + padRightVisible("latest", latestW) + "  status"
	b.WriteString(paint(hdr, ansiGrey, colorOn))
	b.WriteByte('\n')
	for _, r := range rows {
		line := indent + padRightVisible(r.dir, dirW) + "  " + padRightVisible(r.latest, latestW)
		if r.status != "" {
			line += "  " + r.status
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
}

func writeMultiRepoModuleNodes(b *strings.Builder, nodes []UnwindGraphModuleNode, displayByLabel map[string]string, colorOn bool) {
	// Group by repo label; stable label order, dirs within group.
	byLabel := make(map[string][]UnwindGraphModuleNode)
	for _, n := range nodes {
		byLabel[n.RepoLabel] = append(byLabel[n.RepoLabel], n)
	}
	labels := make([]string, 0, len(byLabel))
	for l := range byLabel {
		labels = append(labels, l)
	}
	sort.Strings(labels)
	for _, label := range labels {
		display := displayByLabel[label]
		if display == "" {
			display = label
		}
		fmt.Fprintf(b, "  modules @ %s (%s):\n", label, display)
		group := sortedModuleNodesByDir(byLabel[label])
		// Within a multi-repo group, list bare dir (group already scopes label).
		writeAlignedModuleNodes(b, group, "    ", colorOn)
	}
}

func writeCollapsedRepoEdges(b *strings.Builder, edges []RepoEdge) {
	// Group by From; stable sort of from keys and to keys.
	byFrom := make(map[string][]string)
	for _, e := range edges {
		byFrom[e.From] = append(byFrom[e.From], e.To)
	}
	froms := make([]string, 0, len(byFrom))
	for f := range byFrom {
		froms = append(froms, f)
	}
	sort.Strings(froms)
	for _, from := range froms {
		tos := byFrom[from]
		sort.Strings(tos)
		// Dedupe tos while preserving sorted order.
		prev := ""
		fmt.Fprintf(b, "  %s:\n", from)
		for _, to := range tos {
			if to == prev {
				continue
			}
			prev = to
			fmt.Fprintf(b, "    → %s\n", to)
		}
	}
}

// moduleEdgeBundle merges require+replace for the same (from,to) into one human line.
type moduleEdgeBundle struct {
	fromKey    string
	toKey      string
	requireVer string
	hasRequire bool
	hasReplace bool
	latest     string
}

func writeCollapsedModuleEdges(b *strings.Builder, edges []UnwindGraphModuleEdge, pathToNode map[string]UnwindGraphModuleNode, multiRepo bool, colorOn bool) {
	// Keyed by fromPath\0toPath then rendered with human keys.
	type pair struct{ from, to string }
	bundles := make(map[pair]*moduleEdgeBundle)
	for _, e := range edges {
		fromNode, okF := pathToNode[e.From]
		toNode, okT := pathToNode[e.To]
		if !okF || !okT {
			continue
		}
		p := pair{from: e.From, to: e.To}
		bun, ok := bundles[p]
		if !ok {
			bun = &moduleEdgeBundle{
				fromKey: moduleHumanKey(fromNode, multiRepo),
				toKey:   moduleHumanKey(toNode, multiRepo),
				latest:  toNode.LatestTag,
			}
			bundles[p] = bun
		}
		switch e.Kind {
		case "replace":
			bun.hasReplace = true
		default: // require
			bun.hasRequire = true
			if e.Version != "" {
				bun.requireVer = e.Version
			}
		}
	}
	// Collapse by fromKey for display order.
	byFrom := make(map[string][]*moduleEdgeBundle)
	for _, bun := range bundles {
		byFrom[bun.fromKey] = append(byFrom[bun.fromKey], bun)
	}
	fromKeys := make([]string, 0, len(byFrom))
	for f := range byFrom {
		fromKeys = append(fromKeys, f)
	}
	// Global column width for dep keys across all groups (scannable columns).
	toKeyW := 0
	for _, list := range byFrom {
		for _, bun := range list {
			if w := visibleLen(bun.toKey); w > toKeyW {
				toKeyW = w
			}
		}
	}
	annotW := 0
	type rendered struct {
		toKey, annot, drift string
	}
	// Pre-render annot/drift (with color) and measure plain annot width for pad.
	type groupLines struct {
		from  string
		lines []rendered
	}
	var groups []groupLines
	sort.Strings(fromKeys)
	for _, from := range fromKeys {
		list := byFrom[from]
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].toKey < list[j].toKey
		})
		gl := groupLines{from: from}
		for _, bun := range list {
			var annotParts []string
			// Plain annot for width (no ANSI).
			var plainAnnotParts []string
			if bun.hasRequire {
				if bun.requireVer != "" {
					annotParts = append(annotParts, "require "+bun.requireVer)
					plainAnnotParts = append(plainAnnotParts, "require "+bun.requireVer)
				} else {
					annotParts = append(annotParts, "require")
					plainAnnotParts = append(plainAnnotParts, "require")
				}
			}
			if bun.hasReplace {
				annotParts = append(annotParts, paint("replaced", ansiOrange, colorOn))
				plainAnnotParts = append(plainAnnotParts, "replaced")
			}
			annot := strings.Join(annotParts, "  ")
			plainAnnot := strings.Join(plainAnnotParts, "  ")
			if w := visibleLen(plainAnnot); w > annotW {
				annotW = w
			}
			drift := ""
			// Drift only when require version differs from dep latest (normalized).
			if bun.hasRequire && bun.latest != "" && bun.requireVer != "" && !versionsMatch(bun.requireVer, bun.latest) {
				// Prefer short version in annotation for scan; tests accept vX.Y.Z.
				lv := tagVersion(bun.latest)
				if lv == "" {
					lv = bun.latest
				}
				drift = paint("(latest "+lv+")", ansiOrange, colorOn)
			}
			gl.lines = append(gl.lines, rendered{toKey: bun.toKey, annot: annot, drift: drift})
		}
		groups = append(groups, gl)
	}
	anyDrift := false
	for _, g := range groups {
		for _, ln := range g.lines {
			if ln.drift != "" {
				anyDrift = true
				break
			}
		}
		if anyDrift {
			break
		}
	}
	for _, g := range groups {
		fmt.Fprintf(b, "  %s:\n", g.from)
		for _, ln := range g.lines {
			// "    → <toKey>  <annot>  <drift>"
			line := "    → " + padRightVisible(ln.toKey, toKeyW)
			if ln.annot != "" {
				if anyDrift {
					// Pad annot so (latest …) lines up across rows.
					line += "  " + padRightVisible(ln.annot, annotW)
				} else {
					line += "  " + ln.annot
				}
			}
			if ln.drift != "" {
				line += "  " + ln.drift
			}
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
}

// FormatUnwindGraphJSON renders the show-graph report as pure JSON (trailing newline).
func FormatUnwindGraphJSON(report *UnwindGraphReport) ([]byte, error) {
	if report == nil {
		return []byte("{}\n"), nil
	}
	// Ensure empty slices encode as [] not null.
	reposNodes := report.Repos.Nodes
	if reposNodes == nil {
		reposNodes = []UnwindGraphRepoNode{}
	}
	reposEdges := report.Repos.Edges
	if reposEdges == nil {
		reposEdges = []RepoEdge{}
	}
	peelOrder := report.Repos.PeelOrder
	if peelOrder == nil {
		peelOrder = []string{}
	}
	modNodes := report.Modules.Nodes
	if modNodes == nil {
		modNodes = []UnwindGraphModuleNode{}
	}
	modEdges := report.Modules.Edges
	if modEdges == nil {
		modEdges = []UnwindGraphModuleEdge{}
	}
	warnings := report.Warnings
	if warnings == nil {
		warnings = []string{}
	}

	type edgeJSON struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	repoEdgeJSON := make([]edgeJSON, 0, len(reposEdges))
	for _, e := range reposEdges {
		repoEdgeJSON = append(repoEdgeJSON, edgeJSON{From: e.From, To: e.To})
	}

	out := struct {
		Repos struct {
			Nodes           []UnwindGraphRepoNode `json:"nodes"`
			Edges           []edgeJSON            `json:"edges"`
			PeelOrder       []string              `json:"peel_order"`
			HasPendingEdges bool                  `json:"has_pending_edges"`
			NeedsLand       bool                  `json:"needs_land"`
		} `json:"repos"`
		Modules struct {
			Nodes []UnwindGraphModuleNode `json:"nodes"`
			Edges []UnwindGraphModuleEdge `json:"edges"`
		} `json:"modules"`
		Summary  UnwindGraphSummary `json:"summary"`
		Warnings []string           `json:"warnings"`
	}{}
	out.Repos.Nodes = reposNodes
	out.Repos.Edges = repoEdgeJSON
	out.Repos.PeelOrder = peelOrder
	out.Repos.HasPendingEdges = report.Repos.HasPendingEdges
	out.Repos.NeedsLand = report.Repos.NeedsLand
	out.Modules.Nodes = modNodes
	out.Modules.Edges = modEdges
	out.Summary = report.Summary
	out.Warnings = warnings

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	return data, nil
}

// runUnwindShowGraph is the read-only show-graph path: report only, no apply.
// colorOn applies only to human stdout; JSON is always plain.
func runUnwindShowGraph(workDir string, jsonOut bool, colorOn bool) error {
	report, err := BuildUnwindGraphReport(workDir)
	if err != nil {
		return err
	}
	// Soft inventory warnings on stderr (same prefix policy as runUnwind).
	for _, w := range report.Warnings {
		msg := w
		if !strings.HasPrefix(msg, "warning:") && !strings.HasPrefix(msg, "Warning:") {
			msg = "warning: " + msg
		}
		fmt.Fprintln(os.Stderr, msg)
	}
	if jsonOut {
		data, err := FormatUnwindGraphJSON(report)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(data)
		return err
	}
	_, err = fmt.Fprint(os.Stdout, FormatUnwindGraphHuman(report, colorOn))
	return err
}
