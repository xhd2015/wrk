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
	Done           bool
	MergeBack      bool
	ReinstallLocal bool
	Color          bool
	Sync           bool
	GenCommitMsg   bool
	GenCommitArgs  []string
}

// UnwindPlan is the free-first peel plan for dirty pending stack members.
type UnwindPlan struct {
	PeelOrder []string // labels, free-first
	// HasPendingEdges is true when residual pending graph has any edge.
	HasPendingEdges bool
	// NeedsLand is true when any pending member is a linked worktree.
	NeedsLand bool
}

// BuildStackInventory discovers the checkout stack under workDir: primary git
// toplevel plus nested independent repos (status-like scan), with dirtiness
// and stable main-repo labels.
func BuildStackInventory(workDir string) ([]StackMember, error) {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve cwd: %w", err)
	}
	if !worktree.IsInsideWorkTree(cwd) {
		return nil, fmt.Errorf("%s is not a git repository", cwd)
	}
	checkoutRoot, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return nil, err
	}

	repos, err := discoverStatusRepos(context.Background(), checkoutRoot)
	if err != nil {
		return nil, err
	}

	// Ensure primary checkout is always present even if scan omits it.
	seen := make(map[string]struct{}, len(repos)+1)
	paths := make([]string, 0, len(repos)+1)
	addPath := func(p string) {
		n := storage.NormalizePath(p)
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
		paths = append(paths, p)
	}
	addPath(checkoutRoot)
	for _, r := range repos {
		addPath(r.Path)
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
			return nil, err
		}
		members = append(members, StackMember{
			Path:     storage.NormalizePath(p),
			MainRepo: mainRepo,
			Label:    label,
			Dirty:    dirty,
			Linked:   worktree.IsLinked(p),
		})
	}
	return members, nil
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
func runUnwind(workDir string, flags UnwindFlags) error {
	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}
	members, err := BuildStackInventory(workDir)
	if err != nil {
		return err
	}
	edges, err := BuildRepoDAG(members)
	if err != nil {
		return err
	}
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

// ApplyUnwind peels free-first dirty stack repos with explicit ship/land flags,
// then pins stack consumers of each peeled dep to the new (or latest) tags.
// Preflight (cycle + flags) must already have passed; this mutates.
func ApplyUnwind(workDir, wrkHome string, members []StackMember, edges []RepoEdge, plan *UnwindPlan, flags UnwindFlags) error {
	_ = workDir
	if plan == nil {
		return nil
	}
	byLabel := pickPeelMembersByLabel(members)
	// Reinstall is a tail stage: collect each peeled main repository in the
	// deterministic free-first order, then run it only after every peel and pin
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

	for i, label := range plan.PeelOrder {
		m, ok := byLabel[label]
		if !ok {
			return fmt.Errorf("wrk: peel target %s not found in stack inventory", label)
		}
		if i > 0 {
			fmt.Println()
		}
		// Banner uses checkout Path relative to invocation workDir (statusDirLine policy).
		fmt.Printf("==== unwind: peel %s ====\n", peelDisplayPath(workDir, m.Path))

		mainPath := m.MainRepo
		if mainPath == "" {
			mainPath = m.Path
		}

		// Linked worktree: land with --done (remove) or --merge-back (keep).
		// Already-main: skip land; ship tag/push against main checkout.
		if m.Linked {
			if !flags.Done && !flags.MergeBack {
				return fmt.Errorf("wrk: --unwind for linked worktrees requires --done or --merge-back")
			}
			if flags.GenCommitMsg {
				fmt.Println("---- generate commit message ----")
				// Stage only when --add-all is in GenCommitArgs (library honors
				// it). Do not unconditional git add -A before gen-commit so
				// untracked dirt is not forced into the AI commit.
				if err := runGenCommitMsgStage(m.Path, flags.GenCommitArgs, false); err != nil {
					// Empty index without --add-all: AI gen-commit has nothing
					// to work on. Soft-skip and let auto-commit / land handle
					// remaining dirt (fixtures with only untracked dirt).
					if genArgsHasFlag(flags.GenCommitArgs, "--add-all") || !isNoStagedGenCommitErr(err) {
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
			if err := runSync(mainPath, false); err != nil {
				return err
			}
		}

		var createdTags []string
		if flags.TagNext {
			if err := requireMainActiveRoot(mainPath, "--tag-next"); err != nil {
				return err
			}
			fmt.Println()
			tagRes, err := runTagNextAtResult(mainPath, "HEAD", false, false, false)
			if err != nil {
				return err
			}
			createdTags = tagRes.Tags
		}
		if flags.Push {
			fmt.Println()
			// Publish branch + tags created by this peel's tag-next (if any).
			if err := runPushMain(mainPath, false, createdTags); err != nil {
				return err
			}
		}

		// After peeling U: Pin every stack consumer that depends on U, then tidy.
		version := pinVersionFromCreatedTags(createdTags)
		if err := pinConsumersOfPeeled(label, mainPath, version, members, edges); err != nil {
			return err
		}
		addReinstallMainPath(mainPath)
	}

	if flags.ReinstallLocal {
		for _, mainPath := range reinstallMainPaths {
			if err := runUnwindReinstallLocal(mainPath, flags.Color); err != nil {
				return err
			}
		}
	}
	return nil
}

// runUnwindReinstallLocal executes one unwind tail entry. A repository without
// modules is not an error for a successful unwind, matching the compose tail.
func runUnwindReinstallLocal(mainPath string, colorFlag bool) error {
	err := runReinstallLocal(mainPath, false, true, colorFlag)
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "no go.mod modules found") ||
		strings.Contains(err.Error(), "no go.mod found") {
		fmt.Fprintf(os.Stderr, "skip reinstall-local: %s\n", err.Error())
		return nil
	}
	return err
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

// pinConsumersOfPeeled pins each stack consumer repo that depends on peeledLabel
// to the given version (or latest tag when version is empty), then go mod tidy.
func pinConsumersOfPeeled(peeledLabel, depMainPath, version string, members []StackMember, edges []RepoEdge) error {
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

	// Prefer main-repo path for each consumer label (pin target / assert surface).
	mainByLabel := make(map[string]string, len(members))
	for _, m := range members {
		if m.MainRepo == "" {
			continue
		}
		if _, ok := mainByLabel[m.Label]; !ok {
			mainByLabel[m.Label] = m.MainRepo
		}
		// Prefer non-linked (main checkout) when available.
		if !m.Linked {
			mainByLabel[m.Label] = m.MainRepo
		}
	}

	// Dep modules under the peeled main: Pin each module dir (fixtures: one root).
	depModDirs, err := moduleDirsUnder(depMainPath)
	if err != nil {
		return err
	}
	if len(depModDirs) == 0 {
		depModDirs = []string{depMainPath}
	}

	for _, cl := range consumerLabels {
		consumerDir := mainByLabel[cl]
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
			for _, dMod := range depModDirs {
				fmt.Printf("pin %s <- %s", cl, peeledLabel)
				if version != "" {
					fmt.Printf(" @ %s", version)
				}
				fmt.Println()
				_, err := update.Pin(update.PinOptions{
					ConsumerDir: cMod,
					DepDir:      dMod,
					Version:     version,
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
