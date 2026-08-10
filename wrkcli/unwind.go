package wrkcli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
	"github.com/xhd2015/wrk/wrkcli/storage"
	"golang.org/x/mod/modfile"
)

// StackMember is one git checkout in the consumer stack (primary or nested external).
type StackMember struct {
	// Path is the checkout toplevel path (linked worktree or main).
	Path string
	// MainRepo is the main repository directory for this checkout.
	MainRepo string
	// Label is filepath.Base(MainRepo) — DAG/pin short name (not peel display path).
	Label string
	// Dirty is true when porcelain status is non-empty (untracked counts).
	Dirty bool
	// Linked is true when Path is a linked worktree (not the main checkout).
	Linked bool
}

// RepoEdge is a depends-on edge in the stack repo DAG: From depends on To
// (module require/replace contracted to main-repo labels).
type RepoEdge struct {
	From string // consumer label
	To   string // dependency label
}

// UnwindFlags are ship/land modifiers composed with --unwind.
type UnwindFlags struct {
	DryRun         bool
	TagNext        bool
	Push           bool
	Force          bool // with Push: force-with-lease branch push
	Done           bool
	MergeBack      bool
	ReinstallLocal bool
	Color          bool
	NoColor        bool // --no-color: force plain stdout (mutually exclusive with Color)
	Sync           bool
	GenCommitMsg   bool
	GenCommitArgs  []string
	// AddAll stages all changes when cascade pin commits (and gen-commit when set).
	// Top-level --add-all is accepted with --unwind without requiring --commit.
	AddAll bool
	// ShowGraph is the read-only inspect path (--unwind --show-graph).
	ShowGraph bool
	// Verify is the read-only post-job audit path (--unwind --verify).
	Verify bool
	// JSON requests machine-readable show-graph / verify output.
	JSON bool
}

// UnwindApplyStats counts successful apply stages for the end summary line.
type UnwindApplyStats struct {
	HadPeels    bool // true when PeelOrder was non-empty
	Peeled      int
	Tagged      int
	Pinned      int
	Pushed      int
	Reinstalled int
}

// UnwindPlan is the free-first peel plan for dirty pending stack members.
type UnwindPlan struct {
	PeelOrder []string // labels, free-first
	// HasPendingEdges is true when residual pending graph has any edge.
	HasPendingEdges bool
	// NeedsLand is true when any pending member is a linked worktree.
	NeedsLand bool
}

// StackInventory is the expanded checkout stack for --unwind, including
// synthetic follow edges and soft warnings from local filesystem replaces.
type StackInventory struct {
	Members        []StackMember
	SyntheticEdges []RepoEdge // C→D when follow maps consumer C into dep checkout D
	Warnings       []string   // lines with warning: prefix (missing/non-git targets)
}

// BuildStackInventory discovers the checkout stack under workDir: primary git
// toplevel plus nested independent repos (status-like scan), expanded via local
// filesystem replace BFS, with dirtiness and stable main-repo labels.
func BuildStackInventory(workDir string) ([]StackMember, error) {
	inv, err := CollectStackInventory(workDir)
	if err != nil {
		return nil, err
	}
	return inv.Members, nil
}

// CollectStackInventory is BuildStackInventory plus synthetic follow edges and
// soft warnings from missing/non-git local replace targets.
func CollectStackInventory(workDir string) (StackInventory, error) {
	var empty StackInventory
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return empty, fmt.Errorf("resolve cwd: %w", err)
	}
	if !worktree.IsInsideWorkTree(cwd) {
		return empty, fmt.Errorf("%s is not a git repository", cwd)
	}
	checkoutRoot, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return empty, err
	}

	repos, err := discoverStatusRepos(context.Background(), checkoutRoot)
	if err != nil {
		return empty, err
	}

	// Seed: primary checkout + nested independent repos under it.
	seed := make([]string, 0, len(repos)+1)
	seedSeen := make(map[string]struct{}, len(repos)+1)
	addSeed := func(p string) {
		n := storage.NormalizePath(p)
		if _, ok := seedSeen[n]; ok {
			return
		}
		seedSeen[n] = struct{}{}
		seed = append(seed, n)
	}
	addSeed(checkoutRoot)
	for _, r := range repos {
		addSeed(r.Path)
	}

	paths, pathEdges, warnings, err := expandStackViaLocalReplaces(seed)
	if err != nil {
		return empty, err
	}

	members := make([]StackMember, 0, len(paths))
	for _, p := range paths {
		mainRepo, err := worktree.ResolveMainRepo(p)
		if err != nil {
			// Fall back to path itself when main cannot be resolved.
			mainRepo = p
		}
		mainRepo = storage.NormalizePath(mainRepo)
		label := filepath.Base(mainRepo)
		dirty, err := stackCheckoutDirty(p)
		if err != nil {
			return empty, err
		}
		members = append(members, StackMember{
			Path:     storage.NormalizePath(p),
			MainRepo: mainRepo,
			Label:    label,
			Dirty:    dirty,
			Linked:   worktree.IsLinked(p),
		})
	}

	return StackInventory{
		Members:        members,
		SyntheticEdges: pathEdgesToRepoEdges(members, pathEdges),
		Warnings:       warnings,
	}, nil
}

// stackPathEdge is a synthetic depend edge between checkout paths (pre-label).
type stackPathEdge struct {
	fromPath string // consumer checkout (normalized)
	toPath   string // dep checkout (normalized)
}

// expandStackViaLocalReplaces BFS-expands seed checkout paths by following local
// filesystem replaces on every Go module under each known checkout. Intra-repo
// targets (toplevel already in the set) are not re-added; missing/non-git
// targets yield soft warnings and are skipped.
func expandStackViaLocalReplaces(seed []string) (paths []string, pathEdges []stackPathEdge, warnings []string, err error) {
	seen := make(map[string]struct{}, len(seed))
	queue := make([]string, 0, len(seed))
	addPath := func(p string) bool {
		n := storage.NormalizePath(p)
		if _, ok := seen[n]; ok {
			return false
		}
		seen[n] = struct{}{}
		paths = append(paths, n)
		queue = append(queue, n)
		return true
	}
	for _, s := range seed {
		addPath(s)
	}

	edgeSeen := make(map[string]struct{})
	addPathEdge := func(from, to string) {
		from = storage.NormalizePath(from)
		to = storage.NormalizePath(to)
		if from == "" || to == "" || from == to {
			return
		}
		key := from + "\x00" + to
		if _, ok := edgeSeen[key]; ok {
			return
		}
		edgeSeen[key] = struct{}{}
		pathEdges = append(pathEdges, stackPathEdge{fromPath: from, toPath: to})
	}

	for qi := 0; qi < len(queue); qi++ {
		checkout := queue[qi]
		scanned, scanErr := scan.Scan(checkout, scan.Options{})
		if scanErr != nil {
			return nil, nil, nil, scanErr
		}
		for _, sm := range scanned {
			modDir := checkout
			if sm.Dir != "" && sm.Dir != "." {
				modDir = filepath.Join(checkout, filepath.FromSlash(sm.Dir))
			}
			for _, repl := range sm.LocalFilesystemReplaces() {
				resolved, resolveErr := resolveLocalReplacePath(modDir, repl.NewPath)
				if resolveErr != nil {
					warnings = append(warnings, fmt.Sprintf(
						"warning: local replace target %s (from %s): %v",
						repl.NewPath, modDir, resolveErr))
					continue
				}
				if _, statErr := os.Stat(resolved); statErr != nil {
					warnings = append(warnings, fmt.Sprintf(
						"warning: local replace target missing: %s (from %s)",
						resolved, modDir))
					continue
				}
				if !worktree.IsInsideWorkTree(resolved) {
					warnings = append(warnings, fmt.Sprintf(
						"warning: local replace target is not a git worktree: %s (from %s)",
						resolved, modDir))
					continue
				}
				toplevel, topErr := worktree.ShowToplevel(resolved)
				if topErr != nil {
					warnings = append(warnings, fmt.Sprintf(
						"warning: local replace target %s: %v", resolved, topErr))
					continue
				}
				toplevel = storage.NormalizePath(toplevel)
				// Always record synthetic edge C→D when follow resolves into D
				// (including when D is already a seed member). Same-checkout
				// (intra-repo) is filtered by from==to in addPathEdge.
				addPathEdge(checkout, toplevel)
				// Extra-repo: enqueue for BFS fixpoint. Intra-repo / already
				// seen: do not re-add as a separate member.
				addPath(toplevel)
			}
		}
	}
	return paths, pathEdges, warnings, nil
}

// resolveLocalReplacePath resolves a local filesystem replace NewPath against
// the owning module directory (not necessarily the repo root).
func resolveLocalReplacePath(modDir, newPath string) (string, error) {
	if newPath == "" {
		return "", fmt.Errorf("empty replace path")
	}
	if filepath.IsAbs(newPath) {
		return filepath.Clean(newPath), nil
	}
	return filepath.Clean(filepath.Join(modDir, newPath)), nil
}

// pathEdgesToRepoEdges maps checkout-path synthetic edges to peel labels.
func pathEdgesToRepoEdges(members []StackMember, pathEdges []stackPathEdge) []RepoEdge {
	labelByPath := make(map[string]string, len(members))
	for _, m := range members {
		labelByPath[m.Path] = m.Label
	}
	edgeSet := make(map[string]RepoEdge, len(pathEdges))
	for _, pe := range pathEdges {
		from := labelByPath[pe.fromPath]
		to := labelByPath[pe.toPath]
		if from == "" || to == "" || from == to {
			continue
		}
		key := from + "\x00" + to
		edgeSet[key] = RepoEdge{From: from, To: to}
	}
	out := make([]RepoEdge, 0, len(edgeSet))
	for _, e := range edgeSet {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}

// mergeRepoEdges unions two edge lists, deduping by From+To.
func mergeRepoEdges(a, b []RepoEdge) []RepoEdge {
	edgeSet := make(map[string]RepoEdge, len(a)+len(b))
	add := func(e RepoEdge) {
		if e.From == "" || e.To == "" || e.From == e.To {
			return
		}
		edgeSet[e.From+"\x00"+e.To] = e
	}
	for _, e := range a {
		add(e)
	}
	for _, e := range b {
		add(e)
	}
	out := make([]RepoEdge, 0, len(edgeSet))
	for _, e := range edgeSet {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}

// stackCheckoutDirty reports whether the checkout has uncommitted changes
// (including untracked files). Matches worktree.IsClean porcelain semantics.
func stackCheckoutDirty(path string) (bool, error) {
	err := worktree.IsClean(path)
	if err == nil {
		return false, nil
	}
	// IsClean returns a descriptive error for dirty; other failures are hard errors.
	if strings.Contains(err.Error(), "uncommitted changes") {
		return true, nil
	}
	// Also treat generic non-clean as dirty when git status ran.
	if strings.Contains(err.Error(), "worktree") {
		return true, nil
	}
	return false, err
}

// BuildRepoDAG contracts module require/replace edges among stack-owned modules
// into a repo-level DAG keyed by peel labels. Edge From→To means From depends on To.
func BuildRepoDAG(members []StackMember) ([]RepoEdge, error) {
	// module path → label (first owner wins; fixtures are single-module repos).
	modOwner := make(map[string]string)
	type modRec struct {
		label    string
		requires []string
		replaces []string // old paths (and new module paths when absolute)
	}
	var mods []modRec

	for _, m := range members {
		scanned, err := scan.Scan(m.Path, scan.Options{})
		if err != nil {
			return nil, err
		}
		for _, sm := range scanned {
			if sm.Path != "" {
				if _, ok := modOwner[sm.Path]; !ok {
					modOwner[sm.Path] = m.Label
				}
			}
			rec := modRec{label: m.Label}
			for _, req := range sm.Requires {
				if req.Path != "" {
					rec.requires = append(rec.requires, req.Path)
				}
			}
			// Tolerant require fallback when scan dropped invalid versions.
			if len(rec.requires) == 0 {
				modDir := m.Path
				if sm.Dir != "" && sm.Dir != "." {
					modDir = filepath.Join(m.Path, filepath.FromSlash(sm.Dir))
				}
				if reqs, err := parseRequiresTolerant(filepath.Join(modDir, "go.mod")); err == nil {
					for _, r := range reqs {
						rec.requires = append(rec.requires, r.Path)
					}
				}
			}
			for _, repl := range sm.Replaces {
				if repl.OldPath != "" {
					rec.replaces = append(rec.replaces, repl.OldPath)
				}
				if repl.NewPath != "" && !strings.HasPrefix(repl.NewPath, ".") && !filepath.IsAbs(repl.NewPath) {
					// Module-path replace target (not a relative path).
					rec.replaces = append(rec.replaces, repl.NewPath)
				}
			}
			mods = append(mods, rec)
		}
	}

	edgeSet := make(map[string]RepoEdge)
	addEdge := func(from, to string) {
		if from == "" || to == "" || from == to {
			return
		}
		key := from + "\x00" + to
		edgeSet[key] = RepoEdge{From: from, To: to}
	}

	for _, rec := range mods {
		for _, dep := range rec.requires {
			if owner, ok := modOwner[dep]; ok {
				addEdge(rec.label, owner)
			}
		}
		for _, dep := range rec.replaces {
			if owner, ok := modOwner[dep]; ok {
				addEdge(rec.label, owner)
			}
		}
	}

	edges := make([]RepoEdge, 0, len(edgeSet))
	for _, e := range edgeSet {
		edges = append(edges, e)
	}
	// Stable order for diagnostics.
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})
	return edges, nil
}

// RepoDAGHasCycle reports whether the directed repo graph has a cycle.
func RepoDAGHasCycle(labels []string, edges []RepoEdge) bool {
	labelSet := make(map[string]struct{}, len(labels))
	for _, l := range labels {
		labelSet[l] = struct{}{}
	}
	// Restrict edges to known labels.
	var e2 []RepoEdge
	for _, e := range edges {
		if _, ok := labelSet[e.From]; !ok {
			continue
		}
		if _, ok := labelSet[e.To]; !ok {
			continue
		}
		e2 = append(e2, e)
	}
	order, err := peelOrderKahn(labels, e2)
	if err != nil {
		return true
	}
	return len(order) != len(labels)
}

// PeelOrder returns free-first (Kahn) peel order among pending labels.
// Edge From→To means From depends on To; free nodes have no deps to other pending.
func PeelOrder(pending []string, edges []RepoEdge) ([]string, error) {
	return peelOrderKahn(pending, edges)
}

// peelOrderKahn topo-sorts labels where edge From→To means From depends on To.
// Nodes with out-degree 0 (no remaining deps) peel first.
func peelOrderKahn(labels []string, edges []RepoEdge) ([]string, error) {
	if len(labels) == 0 {
		return nil, nil
	}
	pending := make(map[string]struct{}, len(labels))
	for _, l := range labels {
		pending[l] = struct{}{}
	}
	// deps[from] = set of To that from still depends on
	deps := make(map[string]map[string]struct{}, len(labels))
	// dependents[to] = set of From that depend on to
	dependents := make(map[string]map[string]struct{}, len(labels))
	for _, l := range labels {
		deps[l] = make(map[string]struct{})
		dependents[l] = make(map[string]struct{})
	}
	for _, e := range edges {
		if _, ok := pending[e.From]; !ok {
			continue
		}
		if _, ok := pending[e.To]; !ok {
			continue
		}
		if e.From == e.To {
			continue
		}
		deps[e.From][e.To] = struct{}{}
		dependents[e.To][e.From] = struct{}{}
	}

	var free []string
	for _, l := range labels {
		if len(deps[l]) == 0 {
			free = append(free, l)
		}
	}
	sort.Strings(free)

	order := make([]string, 0, len(labels))
	for len(free) > 0 {
		n := free[0]
		free = free[1:]
		if _, ok := pending[n]; !ok {
			continue
		}
		order = append(order, n)
		delete(pending, n)
		for dep := range dependents[n] {
			if _, ok := pending[dep]; !ok {
				continue
			}
			delete(deps[dep], n)
			if len(deps[dep]) == 0 {
				free = append(free, dep)
			}
		}
		sort.Strings(free)
	}
	if len(order) != len(labels) {
		return order, fmt.Errorf("cycle")
	}
	return order, nil
}

// PlanUnwind builds the free-first peel plan for dirty stack members.
// Cycle detection uses the full stack-repo DAG (not only dirty pending).
func PlanUnwind(members []StackMember, edges []RepoEdge) (*UnwindPlan, error) {
	labels := make([]string, 0, len(members))
	seenLabel := make(map[string]struct{}, len(members))
	for _, m := range members {
		if _, ok := seenLabel[m.Label]; ok {
			continue
		}
		seenLabel[m.Label] = struct{}{}
		labels = append(labels, m.Label)
	}
	if RepoDAGHasCycle(labels, edges) {
		return nil, fmt.Errorf("wrk: dependency cycle detected among stack repos")
	}

	var pending []string
	pendingSet := make(map[string]struct{})
	needsLand := false
	for _, m := range members {
		if !m.Dirty {
			continue
		}
		if _, ok := pendingSet[m.Label]; ok {
			continue
		}
		pendingSet[m.Label] = struct{}{}
		pending = append(pending, m.Label)
		if m.Linked {
			needsLand = true
		}
	}

	// Residual edges among pending only.
	var residual []RepoEdge
	hasPendingEdges := false
	for _, e := range edges {
		_, fromP := pendingSet[e.From]
		_, toP := pendingSet[e.To]
		if fromP && toP {
			residual = append(residual, e)
			hasPendingEdges = true
		}
	}

	order, err := PeelOrder(pending, residual)
	if err != nil {
		// Should not happen if full DAG is acyclic and residual is a subgraph.
		return nil, fmt.Errorf("wrk: dependency cycle detected among stack repos")
	}

	return &UnwindPlan{
		PeelOrder:       order,
		HasPendingEdges: hasPendingEdges,
		NeedsLand:       needsLand,
	}, nil
}

// ValidateUnwindFlags checks pin/land flags required by the plan.
func ValidateUnwindFlags(plan *UnwindPlan, flags UnwindFlags) error {
	if plan == nil {
		return nil
	}
	if plan.HasPendingEdges && (!flags.TagNext || !flags.Push) {
		return fmt.Errorf("wrk: --unwind with cross-repo edges requires --tag-next and --push")
	}
	if plan.NeedsLand && !flags.Done && !flags.MergeBack {
		return fmt.Errorf("wrk: --unwind for linked worktrees requires --done or --merge-back")
	}
	if flags.GenCommitMsg {
		if !genArgsHasFlag(flags.GenCommitArgs, "--commit") {
			return fmt.Errorf("wrk: --commit is required with --gen-commit-msg when used with --unwind")
		}
		if genArgsHasFlag(flags.GenCommitArgs, "--dir") {
			return fmt.Errorf("wrk: --dir is not valid with --gen-commit-msg when used with --unwind")
		}
	}
	return nil
}

// FormatUnwindDryRun returns the dry-run plan text (trailing newline).
// Peel lines use statusDirLine-style relative checkout paths vs workDir.
// When members is non-nil, gen-commit plans reflect --add-all and leave-N
// based on each peel checkout's porcelain (not flags alone).
func FormatUnwindDryRun(plan *UnwindPlan, members []StackMember, workDir string, flags ...UnwindFlags) string {
	var b strings.Builder
	var f UnwindFlags
	if len(flags) > 0 {
		f = flags[0]
	}
	byLabel := pickPeelMembersByLabel(members)
	addAll := genArgsHasFlag(f.GenCommitArgs, "--add-all")
	b.WriteString("==== unwind (dry-run) ====\n")
	if plan != nil {
		for _, label := range plan.PeelOrder {
			display := label
			var peelPath string
			if m, ok := byLabel[label]; ok {
				display = peelDisplayPath(workDir, m.Path)
				peelPath = m.Path
			}
			fmt.Fprintf(&b, "would: peel %s\n", display)
			if f.GenCommitMsg {
				if addAll {
					b.WriteString("  would: git add -A\n")
				} else if peelPath != "" {
					if n, err := countNotFullyStagedPaths(peelPath); err == nil && n > 0 {
						b.WriteString("  " + formatLeaveUncommittedLine(n) + "\n")
					}
				}
				b.WriteString("  would: generate commit message and commit staged changes\n")
			}
			if f.Done || f.MergeBack {
				b.WriteString("  would: merge-back linked worktree into main\n")
			}
			if f.Sync {
				b.WriteString("  would: sync linked worktrees\n")
			}
			if f.TagNext {
				b.WriteString("  would: create release tag\n")
			}
			if f.Push {
				b.WriteString("  would: push branch and created tag\n")
			}
			b.WriteString("  would: pin stack consumers\n")
		}
	}
	// Global free-module cascade after peels (only when --tag-next).
	// Errors are soft: peel plan still prints; empty cascade on failure.
	if f.TagNext && len(members) > 0 {
		if cascade, err := PlanUnwindCascade(members); err == nil {
			b.WriteString(formatUnwindCascadeDryRun(cascade))
		}
	}
	if f.ReinstallLocal {
		b.WriteString("would: reinstall local binaries\n")
	}
	return b.String()
}

// peelDisplayPath formats a peel checkout path for dry-run/apply banners.
// Same policy as statusDirLine (slash form; abs if Rel fails or leading ".." > 2).
func peelDisplayPath(workDir, checkoutPath string) string {
	return statusDirLine(workDir, checkoutPath)
}

// formatLeaveUncommittedLine is the locked dry-run leave-N vocabulary.
func formatLeaveUncommittedLine(n int) string {
	if n == 1 {
		return "would: leave 1 file uncommitted (use --add-all if necessary)"
	}
	return fmt.Sprintf("would: leave %d files uncommitted (use --add-all if necessary)", n)
}

// countNotFullyStagedPaths counts porcelain paths that are not fully staged
// (unstaged modifications and untracked). Fully staged (index change, clean
// worktree) paths are excluded so leave-N is omitted when only staged dirt remains.
func countNotFullyStagedPaths(repoPath string) (int, error) {
	out, err := gitOutputDir(repoPath, "status", "--porcelain")
	if err != nil {
		return 0, err
	}
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 2 {
			continue
		}
		x, y := line[0], line[1]
		// Fully staged: non-space/non-? index status and clean worktree.
		if x != ' ' && x != '?' && y == ' ' {
			continue
		}
		n++
	}
	return n, nil
}

// runUnwind implements wrk --unwind [flags]. Dry-run prints the free-first plan;
// apply peels free-first with explicit ship/land flags and pins consumers.
// --show-graph / --verify are read-only early paths (no ValidateUnwindFlags / ApplyUnwind).
func runUnwind(workDir string, flags UnwindFlags) error {
	if flags.ShowGraph {
		if flags.Color && flags.NoColor {
			return fmt.Errorf("wrk: --color and --no-color are mutually exclusive")
		}
		// JSON never colors; human uses three-mode stdout policy.
		colorOn := false
		if !flags.JSON {
			colorOn = resolveStdoutColor(flags.Color, flags.NoColor)
		}
		return runUnwindShowGraph(workDir, flags.JSON, colorOn)
	}
	if flags.Verify {
		if flags.Color && flags.NoColor {
			return fmt.Errorf("wrk: --color and --no-color are mutually exclusive")
		}
		colorOn := false
		if !flags.JSON {
			colorOn = resolveStdoutColor(flags.Color, flags.NoColor)
		}
		return runUnwindVerify(workDir, flags.JSON, colorOn)
	}
	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}
	inv, err := CollectStackInventory(workDir)
	if err != nil {
		return err
	}
	// Soft follow warnings (missing/non-git local replace targets) on stderr.
	for _, w := range inv.Warnings {
		msg := w
		if !strings.HasPrefix(msg, "warning:") && !strings.HasPrefix(msg, "Warning:") {
			msg = "warning: " + msg
		}
		fmt.Fprintln(os.Stderr, msg)
	}
	members := inv.Members
	edges, err := BuildRepoDAG(members)
	if err != nil {
		return err
	}
	edges = mergeRepoEdges(edges, inv.SyntheticEdges)
	plan, err := PlanUnwind(members, edges)
	if err != nil {
		return err
	}
	if err := ValidateUnwindFlags(plan, flags); err != nil {
		return err
	}
	if flags.DryRun {
		_, err = fmt.Fprint(os.Stdout, FormatUnwindDryRun(plan, members, workDir, flags))
		return err
	}
	return ApplyUnwind(workDir, wrkHome, members, edges, plan, flags)
}

// ApplyUnwind peels free-first dirty stack repos with explicit ship/land flags.
// With --tag-next (B1): peel free deps first → free-module cascade (tag + pin
// consumers, selective pin commits) → peel remaining dirty consumers (feature
// gen-commit sees pinned go.mod). Without --tag-next: land peels and legacy
// pinConsumersOfPeeled (latest tags). Preflight (cycle + flags) must already
// have passed; this mutates.
func ApplyUnwind(workDir, wrkHome string, members []StackMember, edges []RepoEdge, plan *UnwindPlan, flags UnwindFlags) error {
	if plan == nil {
		return nil
	}
	byLabel := pickPeelMembersByLabel(members)
	stdoutColor := resolveStdoutColor(flags.Color, flags.NoColor)
	var stats UnwindApplyStats
	stats.HadPeels = len(plan.PeelOrder) > 0

	// Reinstall is a tail stage: collect each peeled/cascade main repository in
	// deterministic free-first order, then run only after every peel and cascade
	// succeeds. This avoids rebuilding from intermediate stack states.
	reinstallMainPaths := make([]string, 0, len(plan.PeelOrder))
	seenReinstallMainPath := make(map[string]struct{}, len(plan.PeelOrder))
	addReinstallMainPath := func(path string) {
		path = storage.NormalizePath(path)
		if path == "" {
			return
		}
		if _, ok := seenReinstallMainPath[path]; ok {
			return
		}
		seenReinstallMainPath[path] = struct{}{}
		reinstallMainPaths = append(reinstallMainPaths, path)
	}

	peelCount := 0
	peelLabels := func(labels []string) error {
		for _, label := range labels {
			if err := applyUnwindPeelOne(workDir, wrkHome, label, byLabel, members, edges, flags, &stats, addReinstallMainPath, peelCount > 0); err != nil {
				return err
			}
			peelCount++
		}
		return nil
	}

	if !flags.TagNext {
		// Legacy: peel all free-first, pin consumers to latest per peel.
		if err := peelLabels(plan.PeelOrder); err != nil {
			return err
		}
	} else {
		// B1 interleaved free-first: free peels → cascade pin → consumer peels.
		// Consumers that only need pin of a free dep must not gen-commit while a
		// droppable external replace remains (D7 separate pin then feature commit).
		early, deferred := splitPeelOrderB1(plan.PeelOrder, members)
		if err := peelLabels(early); err != nil {
			return err
		}
		// Linked free paths may be removed by --done; remap for cascade graph.
		// Early peels already land into main: also force Path→MainRepo for those
		// labels so cascade pin/tag edits main when --merge-back keeps the WT
		// (else pin commits land only on the post-land linked branch).
		cascadeMembers := refreshStackMembersAfterLand(members)
		cascadeMembers = remapPeeledLabelsToMain(cascadeMembers, early)
		if err := applyUnwindCascade(cascadeMembers, flags, addReinstallMainPath, &stats); err != nil {
			return err
		}
		if err := peelLabels(deferred); err != nil {
			return err
		}
	}

	if flags.ReinstallLocal {
		for _, mainPath := range reinstallMainPaths {
			n, err := runUnwindReinstallLocal(mainPath, flags.Color, flags.NoColor)
			if err != nil {
				return err
			}
			stats.Reinstalled += n
		}
	}

	if line := formatUnwindSummaryLine(stats, flags, stdoutColor); line != "" {
		fmt.Println()
		fmt.Println(line)
	}
	return nil
}

// applyUnwindPeelOne peels one dirty stack label: optional gen-commit, land,
// sync, and (without --tag-next) legacy push/pin. blankBefore prints a blank
// line before the peel banner when prior peels already ran.
func applyUnwindPeelOne(
	workDir, wrkHome, label string,
	byLabel map[string]StackMember,
	members []StackMember,
	edges []RepoEdge,
	flags UnwindFlags,
	stats *UnwindApplyStats,
	addReinstallMainPath func(string),
	blankBefore bool,
) error {
	m, ok := byLabel[label]
	if !ok {
		return fmt.Errorf("wrk: peel target %s not found in stack inventory", label)
	}
	if blankBefore {
		fmt.Println()
	}
	// Banner uses checkout Path relative to invocation workDir (statusDirLine policy).
	fmt.Printf("==== unwind: peel %s ====\n", peelDisplayPath(workDir, m.Path))

	mainPath := m.MainRepo
	if mainPath == "" {
		mainPath = m.Path
	}

	// Linked worktree: land with --done (remove) or --merge-back (keep).
	// Already-main: skip land; ship/cascade against main checkout.
	if m.Linked {
		if !flags.Done && !flags.MergeBack {
			return fmt.Errorf("wrk: --unwind for linked worktrees requires --done or --merge-back")
		}
		if flags.GenCommitMsg {
			// B1 freeHost hole: same-label free+consumer peels early, so global
			// cascade pin has not run yet. Pin ready droppable external stack
			// replaces on this checkout first (separate pin auto-commit), then
			// feature gen-commit sees go.mod without external replace (D7).
			if flags.TagNext {
				if err := pinReadyExternalReplacesBeforeGenCommit(m.Path, members, flags, stats); err != nil {
					return err
				}
			}
			fmt.Println("---- generate commit message ----")
			// Stage only when --add-all is in GenCommitArgs (library honors
			// it). Do not unconditional git add -A before gen-commit so
			// untracked dirt is not forced into the AI commit.
			if err := runGenCommitMsgStage(m.Path, flags.GenCommitArgs, false); err != nil {
				// Empty index without --add-all: AI gen-commit has nothing
				// to work on. Soft-skip and let auto-commit / land handle
				// remaining dirt (fixtures with only untracked dirt).
				if genArgsHasFlag(flags.GenCommitArgs, "--add-all") || flags.AddAll || !isNoStagedGenCommitErr(err) {
					return err
				}
				if err := autoCommitIfDirty(m.Path); err != nil {
					return err
				}
			}
		}
		// --done (Remove) refuses dirty porcelain; auto-commit pending dirt
		// (including leftover untracked after a staged-only gen-commit).
		if flags.Done {
			if err := autoCommitIfDirty(m.Path); err != nil {
				return err
			}
		}
		result, err := worktree.MergeBack(worktree.MergeBackOptions{
			SourcePath: m.Path,
			TargetPath: "",
			Remove:     flags.Done, // --done removes; --merge-back alone keeps
			DryRun:     false,
			TmpDir:     filepath.Join(wrkHome, "worktrees"),
			StashLabel: "wrk-unwind",
			Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
				// Default auto-yes (same as CLI --done/--merge-back).
				return worktree.PromptConfirmPlan(plan, false, true)
			},
		})
		if err != nil {
			return mapMergeBackSharedError(err, "--unwind")
		}
		if result.Action == "aborted" {
			return fmt.Errorf("wrk: merge-back aborted during unwind peel %s", label)
		}
		fmt.Println(result.Message)
		if result.TargetPath != "" {
			mainPath = result.TargetPath
		}
	}

	if flags.Sync {
		fmt.Println("---- sync linked worktrees ----")
		if _, err := runSyncWithColor(mainPath, false, flags.Color, flags.NoColor); err != nil {
			return err
		}
	}

	// With --tag-next: land prelude only; tag/pin/push are global cascade.
	if !flags.TagNext {
		// Legacy: optional push without tag-next, pin consumers to latest.
		if flags.Push {
			fmt.Println()
			if err := runPushMain(mainPath, false, flags.Force, nil); err != nil {
				return err
			}
			if stats != nil {
				stats.Pushed++
			}
		}
		if err := pinConsumersOfPeeled(label, mainPath, nil, members, edges); err != nil {
			return err
		}
	}
	if addReinstallMainPath != nil {
		addReinstallMainPath(mainPath)
	}
	if stats != nil {
		stats.Peeled++
	}
	return nil
}

// splitPeelOrderB1 partitions PeelOrder for B1 free-first interleave:
//
//	early: free dep peels (and same-label free+consumer hosts) — peel before cascade
//	deferred: pure pin-consumer peels — peel after cascade pin drops external replace
//
// Same-repo free+consumer (shared label hosts both free dep and pin consumer)
// stays early so land/DIRTY prelude still runs before cascade tags on that main.
// Plan/cascade failures fall back to all-early (legacy peel-then-cascade).
func splitPeelOrderB1(peelOrder []string, members []StackMember) (early, deferred []string) {
	if len(peelOrder) == 0 {
		return nil, nil
	}
	if len(members) == 0 {
		return append([]string(nil), peelOrder...), nil
	}
	cascade, err := PlanUnwindCascade(members)
	if err != nil || cascade == nil || len(cascade.Steps) == 0 {
		return append([]string(nil), peelOrder...), nil
	}
	byLabel := pickPeelMembersByLabel(members)
	nodes, _, err := buildUnwindModuleGraph(members, byLabel)
	if err != nil {
		return append([]string(nil), peelOrder...), nil
	}
	labelOfMod := make(map[string]string, len(nodes))
	for _, n := range nodes {
		if n.Path != "" && n.RepoLabel != "" {
			labelOfMod[n.Path] = n.RepoLabel
		}
	}
	// freeHost: labels that host a free dep of a planned pin (or a free-only tag).
	// pinConsumer: labels that receive a cascade pin (feature peel deferred).
	freeHost := make(map[string]struct{})
	pinConsumer := make(map[string]struct{})
	for _, s := range cascade.Steps {
		if s.Kind != CascadePin {
			continue
		}
		if lab, ok := labelOfMod[s.ModulePath]; ok && lab != "" {
			pinConsumer[lab] = struct{}{}
		}
		if lab, ok := labelOfMod[s.DepModulePath]; ok && lab != "" {
			freeHost[lab] = struct{}{}
		}
	}
	for _, s := range cascade.Steps {
		if s.Kind != CascadeTagNext {
			continue
		}
		lab, ok := labelOfMod[s.ModulePath]
		if !ok || lab == "" {
			continue
		}
		// Tag module that is a pin dep is already freeHost. Pure free tags
		// (label not only a pin consumer) also keep peels early.
		if _, onlyConsumer := pinConsumer[lab]; !onlyConsumer {
			freeHost[lab] = struct{}{}
		}
	}

	early = make([]string, 0, len(peelOrder))
	deferred = make([]string, 0, len(peelOrder))
	for _, label := range peelOrder {
		_, isPinConsumer := pinConsumer[label]
		_, isFreeHost := freeHost[label]
		// Defer pure pin-consumer peels so cascade pin/commit runs first (B1/D7).
		// Same-label free+consumer hosts stay early (single-repo two modules).
		if isPinConsumer && !isFreeHost {
			deferred = append(deferred, label)
			continue
		}
		early = append(early, label)
	}
	return early, deferred
}

// runUnwindReinstallLocal executes one unwind tail entry. A repository without
// modules is not an error for a successful unwind, matching the compose tail.
// Returns the reinstalled binary count for the summary rollup.
func runUnwindReinstallLocal(mainPath string, colorFlag, noColor bool) (int, error) {
	st, err := runReinstallLocalEx(mainPath, false, true, colorFlag, noColor)
	if err == nil {
		return st.Reinstalled, nil
	}
	if strings.Contains(err.Error(), "no go.mod modules found") ||
		strings.Contains(err.Error(), "no go.mod found") {
		fmt.Fprintf(os.Stderr, "skip reinstall-local: %s\n", err.Error())
		return 0, nil
	}
	return 0, err
}

// pickPeelMembersByLabel chooses the checkout to peel for each label.
// Prefer linked (landable) over main, then dirty over clean.
func pickPeelMembersByLabel(members []StackMember) map[string]StackMember {
	out := make(map[string]StackMember, len(members))
	for _, m := range members {
		prev, ok := out[m.Label]
		if !ok {
			out[m.Label] = m
			continue
		}
		if m.Linked && !prev.Linked {
			out[m.Label] = m
			continue
		}
		if m.Linked == prev.Linked && m.Dirty && !prev.Dirty {
			out[m.Label] = m
		}
	}
	return out
}

// isNoStagedGenCommitErr reports the library "no staged changes" failure so
// unwind can soft-skip AI gen-commit when the index is empty without --add-all.
func isNoStagedGenCommitErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "no staged")
}

// autoCommitIfDirty stages and commits porcelain dirt so --done MergeBack can
// remove the worktree. No-op when clean.
func autoCommitIfDirty(path string) error {
	if err := worktree.IsClean(path); err == nil {
		return nil
	}
	if err := gitRunDir(path, "add", "-A"); err != nil {
		return fmt.Errorf("wrk: auto-commit before unwind land (git add): %w", err)
	}
	// Allow empty in case status flipped; prefer a real commit of dirt.
	if err := gitRunDir(path, "commit", "-m", "wrk: auto-commit before unwind land"); err != nil {
		// If still dirty after failed commit, surface the error.
		if cleanErr := worktree.IsClean(path); cleanErr != nil {
			return fmt.Errorf("wrk: auto-commit before unwind land: %w", err)
		}
	}
	return nil
}

// pinVersionFromCreatedTags picks a go require version from tag-next results.
// Prefers root tags (vN.N.N without path prefix); else strips scope prefix.
// Kept for single-version fallbacks; prefer pinVersionsFromCreatedTags for
// multi-module peels so nested modules do not inherit the root tag version.
func pinVersionFromCreatedTags(tags []string) string {
	for _, t := range tags {
		if t == "" {
			continue
		}
		if !strings.Contains(t, "/") && strings.HasPrefix(t, "v") {
			return t
		}
	}
	for _, t := range tags {
		if t == "" {
			continue
		}
		if i := strings.LastIndex(t, "/"); i >= 0 && i+1 < len(t) {
			return t[i+1:]
		}
		return t
	}
	return ""
}

// pinVersionsFromCreatedTags maps dep module paths to go require versions from
// this peel's created tags. Root tags (vN.N.N, no path prefix) apply only to the
// root module under depMainPath. Path tags (nested/vN.N.N) apply only to the
// module whose directory relative to depMainPath matches the tag path prefix.
// Modules without a matching created tag are omitted (Pin falls back to latest).
func pinVersionsFromCreatedTags(tags []string, depMainPath string, depModByPath map[string]string) map[string]string {
	if len(tags) == 0 || len(depModByPath) == 0 {
		return nil
	}
	// relDir (slash, "" for root) → module path
	byRel := make(map[string]string, len(depModByPath))
	for modPath, dir := range depModByPath {
		rel, err := filepath.Rel(depMainPath, dir)
		if err != nil {
			continue
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			rel = ""
		}
		byRel[rel] = modPath
	}
	out := make(map[string]string)
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		if !strings.Contains(tag, "/") {
			if strings.HasPrefix(tag, "v") {
				if mp, ok := byRel[""]; ok {
					out[mp] = tag
				}
			}
			continue
		}
		// Path-scoped tag: <rel>/vN.N.N (possibly multi-segment rel).
		i := strings.LastIndex(tag, "/")
		if i < 0 || i+1 >= len(tag) {
			continue
		}
		rel := tag[:i]
		ver := tag[i+1:]
		if !strings.HasPrefix(ver, "v") || strings.Contains(ver, "/") {
			continue
		}
		if mp, ok := byRel[rel]; ok {
			out[mp] = ver
		}
	}
	return out
}

// depModuleDirsByPath scans peeled main for go.mod modules: module path → dir.
func depModuleDirsByPath(depMainPath string) (map[string]string, error) {
	scanned, err := scan.Scan(depMainPath, scan.Options{})
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, sm := range scanned {
		if sm.Path == "" {
			continue
		}
		dir := depMainPath
		if sm.Dir != "" && sm.Dir != "." {
			dir = filepath.Join(depMainPath, filepath.FromSlash(sm.Dir))
		}
		out[sm.Path] = dir
	}
	if len(out) == 0 {
		if p, err := readModulePath(depMainPath); err == nil && p != "" {
			out[p] = depMainPath
		} else {
			// Last resort: anonymous root so single-module fixtures still pin.
			out[""] = depMainPath
		}
	}
	return out, nil
}

// consumerReferencedModulePaths returns require + replace old-paths from cMod's
// go.mod. Only those that match a peeled dep module are candidates for Pin
// (avoids force-adding modules the consumer never needed).
func consumerReferencedModulePaths(cMod string) (map[string]struct{}, error) {
	goModPath := filepath.Join(cMod, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{})
	if f, err := modfile.Parse(goModPath, data, nil); err == nil {
		for _, r := range f.Require {
			if r.Mod.Path != "" {
				out[r.Mod.Path] = struct{}{}
			}
		}
		for _, r := range f.Replace {
			if r.Old.Path != "" {
				out[r.Old.Path] = struct{}{}
			}
		}
		return out, nil
	}
	// Tolerant require fallback when strict parse rejects versions.
	if reqs, err := parseRequiresTolerant(goModPath); err == nil {
		for _, r := range reqs {
			if r.Path != "" {
				out[r.Path] = struct{}{}
			}
		}
	}
	// Light replace scrape for OldPath when full parse failed.
	scrapeReplaceOldPaths(string(data), out)
	return out, nil
}

// scrapeReplaceOldPaths adds replace old-paths from go.mod text into out.
func scrapeReplaceOldPaths(content string, out map[string]struct{}) {
	inBlock := false
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		if !inBlock {
			if line == "replace (" {
				inBlock = true
				continue
			}
			if strings.HasPrefix(line, "replace ") && !strings.HasPrefix(line, "replace (") {
				rest := strings.TrimSpace(strings.TrimPrefix(line, "replace "))
				if old := firstPathToken(rest); old != "" {
					out[old] = struct{}{}
				}
			}
			continue
		}
		if line == ")" {
			inBlock = false
			continue
		}
		if old := firstPathToken(line); old != "" {
			out[old] = struct{}{}
		}
	}
}

func firstPathToken(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// pinConsumersOfPeeled pins each stack consumer that depends on peeledLabel.
// For each consumer module, only dep modules that module requires or replaces
// are pinned (no Cartesian force-add of multi-module dep trees). Per-module
// version comes from this peel's created tags when available; empty Version
// lets update.Pin resolve the latest matching tag.
func pinConsumersOfPeeled(peeledLabel, depMainPath string, createdTags []string, members []StackMember, edges []RepoEdge) error {
	consumerLabels := make([]string, 0)
	seen := make(map[string]struct{})
	for _, e := range edges {
		if e.To != peeledLabel || e.From == "" || e.From == peeledLabel {
			continue
		}
		if _, ok := seen[e.From]; ok {
			continue
		}
		seen[e.From] = struct{}{}
		consumerLabels = append(consumerLabels, e.From)
	}
	if len(consumerLabels) == 0 {
		return nil
	}
	sort.Strings(consumerLabels)

	// Prefer in-scope Path for each consumer label (linked over main, dirty over clean).
	// Scope inventory already limits members to primary + nested under cwd; pin that
	// checkout, not out-of-scope MainRepo when a linked Path is present.
	byLabel := pickPeelMembersByLabel(members)
	pathByLabel := make(map[string]string, len(byLabel))
	for label, m := range byLabel {
		if m.Path != "" {
			pathByLabel[label] = m.Path
		}
	}

	depModByPath, err := depModuleDirsByPath(depMainPath)
	if err != nil {
		return err
	}
	versions := pinVersionsFromCreatedTags(createdTags, depMainPath, depModByPath)

	for _, cl := range consumerLabels {
		consumerDir := pathByLabel[cl]
		if consumerDir == "" {
			return fmt.Errorf("wrk: pin consumer %s: no stack path", cl)
		}
		// Consumer module dirs (fixtures: single go.mod at main root).
		consumerMods, err := moduleDirsUnder(consumerDir)
		if err != nil {
			return err
		}
		if len(consumerMods) == 0 {
			consumerMods = []string{consumerDir}
		}
		for _, cMod := range consumerMods {
			wanted, err := consumerReferencedModulePaths(cMod)
			if err != nil {
				// No go.mod / unreadable — skip this consumer module.
				continue
			}
			// Stable pin order by module path.
			var toPin []string
			for mp := range depModByPath {
				if mp == "" {
					// Anonymous fallback when go.mod path could not be read:
					// pin only for single-module dep trees.
					if len(depModByPath) == 1 {
						toPin = append(toPin, mp)
					}
					continue
				}
				if _, ok := wanted[mp]; ok {
					toPin = append(toPin, mp)
				}
			}
			sort.Strings(toPin)
			if len(toPin) == 0 {
				continue
			}
			for _, mp := range toPin {
				dMod := depModByPath[mp]
				ver := versions[mp]
				fmt.Printf("pin %s <- %s", cl, peeledLabel)
				if ver != "" {
					fmt.Printf(" @ %s", ver)
				}
				fmt.Println()
				_, err := update.Pin(update.PinOptions{
					ConsumerDir: cMod,
					DepDir:      dMod,
					Version:     ver,
				})
				if err != nil {
					return fmt.Errorf("wrk: pin %s <- %s: %w", cl, peeledLabel, err)
				}
			}
			if err := goModTidy(cMod); err != nil {
				return fmt.Errorf("wrk: go mod tidy in %s: %w", cMod, err)
			}
			// go mod edit often leaves single-line `require path ver`; expand to a
			// parenthesized block so consumers retain multi-require style.
			if err := expandGoModRequireBlocks(filepath.Join(cMod, "go.mod")); err != nil {
				return fmt.Errorf("wrk: reformat go.mod in %s: %w", cMod, err)
			}
		}
	}
	return nil
}

// expandGoModRequireBlocks rewrites single-line require directives into a
// parenthesized require ( ... ) block. Idempotent when already blocked.
func expandGoModRequireBlocks(goModPath string) error {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}
	f, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return err
	}
	if len(f.Require) == 0 {
		return nil
	}
	// Detect whether any require is already inside a parenthesized block.
	// When go mod edit wrote `require path ver` only, Syntax has single-line stmts.
	needsExpand := false
	for _, stmt := range f.Syntax.Stmt {
		line, ok := stmt.(*modfile.Line)
		if !ok || len(line.Token) < 3 {
			continue
		}
		if line.Token[0] == "require" {
			needsExpand = true
			break
		}
	}
	if !needsExpand {
		return nil
	}

	// Rebuild require section as a single parenthesized block while preserving
	// module/go and other directives via a light rewrite of require lines only.
	var b strings.Builder
	lines := strings.Split(string(data), "\n")
	// Drop existing require lines/blocks; re-append one block from f.Require.
	skipBlock := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if skipBlock {
			if trim == ")" {
				skipBlock = false
			}
			continue
		}
		if strings.HasPrefix(trim, "require (") {
			skipBlock = true
			continue
		}
		if strings.HasPrefix(trim, "require ") && !strings.HasPrefix(trim, "require (") {
			// single-line require — drop; re-emitted below
			continue
		}
		// Keep non-require content, but avoid trailing blank spam — write as-is.
		b.WriteString(line)
		b.WriteByte('\n')
	}
	// Trim trailing blank lines before appending require block.
	out := strings.TrimRight(b.String(), "\n") + "\n\n"
	var rb strings.Builder
	rb.WriteString("require (\n")
	for _, r := range f.Require {
		if r.Indirect {
			fmt.Fprintf(&rb, "\t%s %s // indirect\n", r.Mod.Path, r.Mod.Version)
		} else {
			fmt.Fprintf(&rb, "\t%s %s\n", r.Mod.Path, r.Mod.Version)
		}
	}
	rb.WriteString(")\n")
	out += rb.String()
	return os.WriteFile(goModPath, []byte(out), 0o644)
}

// moduleDirsUnder returns directories under root that contain a go.mod (scan).
func moduleDirsUnder(root string) ([]string, error) {
	scanned, err := scan.Scan(root, scan.Options{})
	if err != nil {
		return nil, err
	}
	var dirs []string
	seen := make(map[string]struct{})
	for _, sm := range scanned {
		dir := root
		if sm.Dir != "" && sm.Dir != "." {
			dir = filepath.Join(root, filepath.FromSlash(sm.Dir))
		}
		n := storage.NormalizePath(dir)
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	return dirs, nil
}
