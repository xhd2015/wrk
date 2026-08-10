package wrkcli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/wrk/wrkcli/storage"
	"golang.org/x/mod/modfile"
)

// UnwindCascadeStepKind is a free-module cascade plan step kind.
type UnwindCascadeStepKind string

const (
	// CascadeTagNext plans a tagscope next release for a free/pending module.
	CascadeTagNext UnwindCascadeStepKind = "tag-next"
	// CascadePin plans pinning a consumer module require onto a dep version.
	CascadePin UnwindCascadeStepKind = "pin"
)

// UnwindCascadeStep is one dry-run cascade action (tag-next or pin).
type UnwindCascadeStep struct {
	Kind UnwindCascadeStepKind
	// ModulePath is the tagged module (tag-next) or consumer module (pin).
	ModulePath string
	// DepModulePath is set for pin steps (dependency module path).
	DepModulePath string
	// TagOrVersion is the full next tag for tag-next, or go require version for pin.
	TagOrVersion string
}

// UnwindCascadePlan is the free-first module cascade over the stack module DAG.
type UnwindCascadePlan struct {
	Steps []UnwindCascadeStep
}

// PlanUnwindCascade builds a free-first module cascade from stack inventory:
// owned-changed modules (tagscope next tags) plus consumers that need pin /
// require-drift. Testdata / forever-skip scopes never emit tag-next steps.
// Does not mutate the workspace.
func PlanUnwindCascade(members []StackMember) (*UnwindCascadePlan, error) {
	if len(members) == 0 {
		return &UnwindCascadePlan{}, nil
	}
	byLabel := pickPeelMembersByLabel(members)
	nodes, edges, err := buildUnwindModuleGraph(members, byLabel)
	if err != nil {
		return nil, err
	}
	attachTagScopeToModules(nodes, members)
	return planUnwindCascadeFromGraph(nodes, edges)
}

// planUnwindCascadeFromGraph is the pure cascade planner over module nodes/edges.
func planUnwindCascadeFromGraph(nodes []UnwindGraphModuleNode, edges []UnwindGraphModuleEdge) (*UnwindCascadePlan, error) {
	out := &UnwindCascadePlan{}
	if len(nodes) == 0 {
		return out, nil
	}

	nodeByPath := make(map[string]UnwindGraphModuleNode, len(nodes))
	for _, n := range nodes {
		if n.Path == "" {
			continue
		}
		nodeByPath[n.Path] = n
	}

	// plannedPinVer: go require version consumers should pin to (prefer next, else latest).
	plannedPinVer := make(map[string]string, len(nodes))
	for _, n := range nodes {
		tag := n.NextTag
		if tag == "" {
			tag = n.LatestTag
		}
		if v := goRequireVersionFromTag(tag); v != "" {
			plannedPinVer[n.Path] = v
		}
	}

	// Droppable external replaces among stack-owned modules (C-DR7): same policy
	// as apply's classifyLocalReplace / localReplacePolicy — intra keep-local
	// (same repo label) does not force pin (C-DR8).
	// Key: consumer\x00dep
	droppableExternal := make(map[string]struct{})
	for _, e := range edges {
		if e.Kind != "replace" || e.From == "" || e.To == "" {
			continue
		}
		from, ok1 := nodeByPath[e.From]
		to, ok2 := nodeByPath[e.To]
		if !ok1 || !ok2 {
			continue
		}
		if isDroppableExternalStackReplace(from, to, e) {
			droppableExternal[e.From+"\x00"+e.To] = struct{}{}
		}
	}

	// Pending modules: owned-changed (taggable) ∪ require-drift ∪ needs-pin
	// (including droppable external replace of a stack dep).
	pending := make(map[string]struct{})
	for _, n := range nodes {
		if cascadeModuleShouldTag(n) {
			pending[n.Path] = struct{}{}
		}
	}
	for _, e := range edges {
		if e.Kind != "require" {
			continue
		}
		if e.From == "" || e.To == "" {
			continue
		}
		if _, ok := nodeByPath[e.From]; !ok {
			continue
		}
		dep, ok := nodeByPath[e.To]
		if !ok {
			continue
		}
		want := plannedPinVer[e.To]
		depWillTag := cascadeModuleShouldTag(dep)
		drift := want != "" && e.Version != "" && !versionsMatch(e.Version, want)
		if depWillTag || drift {
			pending[e.From] = struct{}{}
			if depWillTag {
				pending[e.To] = struct{}{}
			} else if drift {
				// Require-drift alone still needs the consumer; dep stays if already pending.
				_ = dep
			}
		}
	}
	// Droppable external replace alone ⇒ consumer needs-pin (even when require
	// already matches and the free dep will not tag).
	for pair := range droppableExternal {
		from, _, ok := strings.Cut(pair, "\x00")
		if !ok || from == "" {
			continue
		}
		if _, ok := nodeByPath[from]; ok {
			pending[from] = struct{}{}
		}
	}
	// Fixpoint: any module requiring a pending dep is pending (needs-pin).
	for {
		added := false
		for _, e := range edges {
			if e.Kind != "require" {
				continue
			}
			if _, toP := pending[e.To]; !toP {
				continue
			}
			if _, fromP := pending[e.From]; fromP {
				continue
			}
			if _, ok := nodeByPath[e.From]; !ok {
				continue
			}
			pending[e.From] = struct{}{}
			added = true
		}
		if !added {
			break
		}
	}

	if len(pending) == 0 {
		return out, nil
	}

	// Residual module edges among pending (require + replace: From depends on To).
	edgeSeen := make(map[string]struct{})
	var residual []RepoEdge
	for _, e := range edges {
		if e.From == "" || e.To == "" || e.From == e.To {
			continue
		}
		if _, ok := pending[e.From]; !ok {
			continue
		}
		if _, ok := pending[e.To]; !ok {
			continue
		}
		key := e.From + "\x00" + e.To
		if _, ok := edgeSeen[key]; ok {
			continue
		}
		edgeSeen[key] = struct{}{}
		residual = append(residual, RepoEdge{From: e.From, To: e.To})
	}

	labels := make([]string, 0, len(pending))
	for p := range pending {
		labels = append(labels, p)
	}
	sort.Strings(labels)

	order, err := peelOrderKahn(labels, residual)
	if err != nil {
		// Module cycle among pending: do not invent a cascade body.
		return out, fmt.Errorf("wrk: dependency cycle detected among stack modules")
	}

	// Emit free-first per module: pins first (commit-before-tag), then tag-next.
	// Earlier modules are fully processed (including tag) before later consumers pin.
	tagged := make(map[string]string) // module path → full next tag
	for _, modPath := range order {
		n := nodeByPath[modPath]

		// Stable pin order by dep path — before this module's own tag-next.
		var pinDeps []string
		pinVerByDep := make(map[string]string)
		for _, e := range edges {
			if e.Kind != "require" || e.From != modPath {
				continue
			}
			dep, ok := nodeByPath[e.To]
			if !ok {
				continue
			}
			pinVer := plannedPinVer[e.To]
			if pinVer == "" {
				// Fall back to version from cascade tag applied earlier.
				if t, ok := tagged[e.To]; ok {
					pinVer = goRequireVersionFromTag(t)
				}
			}
			// D3: when no tag/drift planned version, keep current require.
			if pinVer == "" && e.Version != "" {
				pinVer = e.Version
			}
			if pinVer == "" {
				continue
			}
			_, wasTagged := tagged[e.To]
			willTag := cascadeModuleShouldTag(dep)
			drift := e.Version != "" && !versionsMatch(e.Version, pinVer)
			_, dropExt := droppableExternal[e.From+"\x00"+e.To]
			// Pin when dep was/will be tagged, require drifts, or there is a
			// droppable external replace of this stack dep (replace-only pin).
			if !wasTagged && !willTag && !drift && !dropExt {
				continue
			}
			if _, seen := pinVerByDep[e.To]; seen {
				continue
			}
			pinVerByDep[e.To] = pinVer
			pinDeps = append(pinDeps, e.To)
		}
		sort.Strings(pinDeps)
		for _, depPath := range pinDeps {
			out.Steps = append(out.Steps, UnwindCascadeStep{
				Kind:          CascadePin,
				ModulePath:    modPath,
				DepModulePath: depPath,
				TagOrVersion:  pinVerByDep[depPath],
			})
		}

		if cascadeModuleShouldTag(n) {
			out.Steps = append(out.Steps, UnwindCascadeStep{
				Kind:         CascadeTagNext,
				ModulePath:   modPath,
				TagOrVersion: n.NextTag,
			})
			tagged[modPath] = n.NextTag
		}
	}
	return out, nil
}

// isDroppableExternalStackReplace reports whether a module-graph replace edge
// should force needs-pin (apply would drop the replace). Aligns with
// classifyLocalReplace: module-path and cross-repo filesystem replaces are
// droppable; same-repo local filesystem replaces are keep-local (no force pin).
func isDroppableExternalStackReplace(from, to UnwindGraphModuleNode, e UnwindGraphModuleEdge) bool {
	if e.Kind != "replace" || e.From == "" || e.To == "" {
		return false
	}
	replNew := e.NewPath
	if replNew == "" {
		// Replace edge without NewPath still targets a stack-owned module; treat
		// cross-repo ownership as droppable external.
		return from.RepoLabel != "" && to.RepoLabel != "" && from.RepoLabel != to.RepoLabel
	}
	// Module-path replace (not filesystem): drop so pin can use a version.
	if !strings.HasPrefix(replNew, ".") && !filepath.IsAbs(replNew) {
		return true
	}
	// Filesystem replace: drop when consumer and dep live in different stack
	// repos (nested external worktree / multi-repo). Same RepoLabel = intra
	// keep-local (./pkgs/shared) — must not force pin alone (C-DR8).
	if from.RepoLabel != "" && to.RepoLabel != "" && from.RepoLabel != to.RepoLabel {
		return true
	}
	return false
}

// cascadeModuleShouldTag reports whether the cascade plans a would: tag-next line.
func cascadeModuleShouldTag(n UnwindGraphModuleNode) bool {
	if n.Path == "" || n.NextTag == "" {
		return false
	}
	if cascadeSkipTagScope(n) {
		return false
	}
	return true
}

// cascadeSkipTagScope excludes forever-skip / testdata scopes from cascade tags.
func cascadeSkipTagScope(n UnwindGraphModuleNode) bool {
	if strings.Contains(strings.ToLower(n.SkipReason), "forever") {
		return true
	}
	// Path-scope or module under testdata must never get tag-next.
	check := n.Dir + "\n" + n.Path + "\n" + n.NextTag + "\n" + n.LatestTag
	if strings.Contains(strings.ToLower(check), "testdata") {
		return true
	}
	return false
}

// goRequireVersionFromTag maps a full release tag to a go.mod require version
// (e.g. pkgs/shared/v0.0.2 → v0.0.2, v0.0.2 → v0.0.2).
func goRequireVersionFromTag(tag string) string {
	v := tagVersion(tag)
	if v == "" {
		return ""
	}
	// Prefer leading "v" for go module versions when the tag is numeric.
	if !strings.HasPrefix(v, "v") {
		// Keep non-v prefixes (rare); only add v when the remainder looks like X.Y.Z.
		if looksLikeSemver(v) {
			return "v" + v
		}
	}
	return v
}

func looksLikeSemver(s string) bool {
	// Minimal: digits and dots, at least one dot.
	if !strings.Contains(s, ".") {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && r != '.' {
			return false
		}
	}
	return true
}

// formatUnwindCascadeDryRun renders cascade plan lines (each ending with \n).
// Empty when plan is nil or has no steps.
func formatUnwindCascadeDryRun(plan *UnwindCascadePlan) string {
	if plan == nil || len(plan.Steps) == 0 {
		return ""
	}
	var b strings.Builder
	for _, s := range plan.Steps {
		switch s.Kind {
		case CascadeTagNext:
			if s.ModulePath == "" || s.TagOrVersion == "" {
				continue
			}
			fmt.Fprintf(&b, "would: tag-next %s @ %s\n", s.ModulePath, s.TagOrVersion)
		case CascadePin:
			if s.ModulePath == "" || s.DepModulePath == "" || s.TagOrVersion == "" {
				continue
			}
			fmt.Fprintf(&b, "would: pin %s <- %s @ %s\n", s.ModulePath, s.DepModulePath, s.TagOrVersion)
		}
	}
	return b.String()
}

// refreshStackMembersAfterLand remaps vanished linked Paths to MainRepo so
// post-land cascade planning scans and tags landed mains (not removed WTs).
func refreshStackMembersAfterLand(members []StackMember) []StackMember {
	if len(members) == 0 {
		return members
	}
	out := make([]StackMember, len(members))
	for i, m := range members {
		out[i] = m
		if m.Path == "" {
			continue
		}
		if st, err := os.Stat(m.Path); err == nil && st.IsDir() {
			continue
		}
		// Path gone (e.g. --done removed worktree): fall back to main checkout.
		if m.MainRepo != "" {
			out[i].Path = m.MainRepo
			out[i].Linked = false
		}
	}
	return out
}

// pinReadyExternalReplacesBeforeGenCommit applies cascade pins for **ready**
// droppable external stack replaces whose consumer modules live on checkout,
// as separate `wrk: cascade pin …` auto-commits, before feature gen-commit.
//
// Used on freeHost early peels (same-label free+consumer): B1 defers only pure
// pin-consumers, so external replace would otherwise still be present at
// gen-commit. Does not pin deps that still need tag-next (owned-changed free
// not yet landed/tagged) — those stay post-land cascade. Idempotent when the
// replace is already gone / require already matches.
func pinReadyExternalReplacesBeforeGenCommit(checkout string, members []StackMember, flags UnwindFlags, stats *UnwindApplyStats) error {
	if checkout == "" || len(members) == 0 {
		return nil
	}
	checkout = storage.NormalizePath(checkout)
	plan, err := PlanUnwindCascade(members)
	if err != nil {
		return err
	}
	if plan == nil || len(plan.Steps) == 0 {
		return nil
	}

	byLabel := pickPeelMembersByLabel(members)
	nodes, edges, err := buildUnwindModuleGraph(members, byLabel)
	if err != nil {
		return err
	}
	nodeByPath := make(map[string]UnwindGraphModuleNode, len(nodes))
	for _, n := range nodes {
		if n.Path != "" {
			nodeByPath[n.Path] = n
		}
	}

	// droppableExternal: same policy as cascade planner (cross-repo / module-path replaces).
	droppableExternal := make(map[string]struct{})
	for _, e := range edges {
		if e.Kind != "replace" || e.From == "" || e.To == "" {
			continue
		}
		from, ok1 := nodeByPath[e.From]
		to, ok2 := nodeByPath[e.To]
		if !ok1 || !ok2 {
			continue
		}
		if isDroppableExternalStackReplace(from, to, e) {
			droppableExternal[e.From+"\x00"+e.To] = struct{}{}
		}
	}

	// Modules hosted on this checkout (linked Path or main matching checkout).
	onCheckout := make(map[string]struct{})
	for _, n := range nodes {
		if n.Path == "" || n.RepoLabel == "" {
			continue
		}
		m, ok := byLabel[n.RepoLabel]
		if !ok {
			continue
		}
		if memberCheckoutPath(m) == checkout {
			onCheckout[n.Path] = struct{}{}
		}
	}
	if len(onCheckout) == 0 {
		return nil
	}

	addAll := flags.AddAll || genArgsHasFlag(flags.GenCommitArgs, "--add-all")

	for _, step := range plan.Steps {
		if step.Kind != CascadePin {
			continue
		}
		if step.ModulePath == "" || step.DepModulePath == "" || step.TagOrVersion == "" {
			continue
		}
		if _, ok := onCheckout[step.ModulePath]; !ok {
			continue
		}
		// Only droppable external stack replaces (not intra keep-local / require-only).
		if _, ok := droppableExternal[step.ModulePath+"\x00"+step.DepModulePath]; !ok {
			continue
		}
		depNode, ok := nodeByPath[step.DepModulePath]
		if !ok {
			continue
		}
		// Ready: dep is not waiting for a cascade tag-next (owned-changed free).
		// Clean free with LatestTag / current require is ready; dirty free that
		// still needs tag stays for post-land cascade.
		if cascadeModuleShouldTag(depNode) {
			continue
		}

		consumerNode, ok := nodeByPath[step.ModulePath]
		if !ok {
			continue
		}
		consumerLabel := consumerNode.RepoLabel
		depLabel := depNode.RepoLabel

		consumerModDir := moduleDirOnCheckout(checkout, consumerNode)
		if consumerModDir == "" {
			return fmt.Errorf("wrk: ready-external pin: empty consumer module dir for %s", step.ModulePath)
		}

		// Idempotent: replace already gone → nothing to pin for this path.
		if !goModHasLocalReplace(consumerModDir, step.DepModulePath) {
			continue
		}
		// Safety: if policy says keep (intra), do not force-drop.
		if keep, _ := localReplacePolicy(consumerModDir, step.DepModulePath); keep {
			continue
		}

		// Dirty go.mod/go.sum without --add-all → partial edit (same as cascade pin).
		usePartial := false
		var saved goModSumSnap
		if !addAll {
			dirty, err := goModSumUncommittedAt(checkout, consumerModDir)
			if err != nil {
				return err
			}
			if dirty {
				usePartial = true
				saved, err = saveGoModSumSnap(consumerModDir)
				if err != nil {
					return fmt.Errorf("wrk: ready-external pin save go.mod/go.sum in %s: %w", consumerModDir, err)
				}
				if err := writeBaseGoModSum(checkout, consumerModDir); err != nil {
					_ = restoreGoModSumSnap(consumerModDir, saved)
					return fmt.Errorf("wrk: ready-external pin restore Base go.mod/go.sum in %s: %w", consumerModDir, err)
				}
			}
		}

		logConsumer := consumerLabel
		if logConsumer == "" {
			logConsumer = filepath.Base(checkout)
		}
		logDep := depLabel
		if logDep == "" {
			logDep = filepath.Base(step.DepModulePath)
		}
		fmt.Printf("pin %s <- %s @ %s\n", logConsumer, logDep, step.TagOrVersion)

		pinFail := func(err error) error {
			if usePartial {
				_ = restoreGoModSumSnap(consumerModDir, saved)
			}
			return err
		}

		if err := cascadePinKeepLocalReplace(consumerModDir, step.DepModulePath, step.TagOrVersion, depNode, byLabel); err != nil {
			return pinFail(fmt.Errorf("wrk: ready-external pin %s <- %s: %w", step.ModulePath, step.DepModulePath, err))
		}
		if err := goModTidy(consumerModDir); err != nil {
			return pinFail(fmt.Errorf("wrk: go mod tidy in %s: %w", consumerModDir, err))
		}
		_ = expandGoModRequireBlocks(filepath.Join(consumerModDir, "go.mod"))

		// Selective pin commit only (D7); never scoop feature WIP into pin.
		if err := cascadeCommitPin(checkout, consumerModDir, step.DepModulePath, step.TagOrVersion, false); err != nil {
			return pinFail(err)
		}
		if stats != nil {
			stats.Pinned++
		}

		if usePartial {
			if err := restoreGoModSumSnap(consumerModDir, saved); err != nil {
				return fmt.Errorf("wrk: ready-external pin restore WIP go.mod/go.sum in %s: %w", consumerModDir, err)
			}
			if err := goModEditRequire(consumerModDir, step.DepModulePath, step.TagOrVersion); err != nil {
				_ = restoreGoModSumSnap(consumerModDir, saved)
				return fmt.Errorf("wrk: ready-external pin surgical require bump %s@%s in %s: %w",
					step.DepModulePath, step.TagOrVersion, consumerModDir, err)
			}
			_ = expandGoModRequireBlocks(filepath.Join(consumerModDir, "go.mod"))
		}
	}
	return nil
}

// memberCheckoutPath prefers still-present Path over MainRepo (same as cascade checkoutForEdit).
func memberCheckoutPath(m StackMember) string {
	if m.Path != "" {
		if st, err := os.Stat(m.Path); err == nil && st.IsDir() {
			return storage.NormalizePath(m.Path)
		}
	}
	main := m.MainRepo
	if main == "" {
		main = m.Path
	}
	return storage.NormalizePath(main)
}

// moduleDirOnCheckout joins module Dir under checkout (root module → checkout).
func moduleDirOnCheckout(checkout string, n UnwindGraphModuleNode) string {
	if checkout == "" {
		return ""
	}
	dir := n.Dir
	if dir == "" || dir == "." {
		return checkout
	}
	return filepath.Join(checkout, filepath.FromSlash(dir))
}

// applyUnwindCascade runs PlanUnwindCascade steps after the land prelude:
// one-scope tags, keep-local-replace pins, selective commits, push when a main
// has no remaining cascade modules. addReinstallMainPath records mains for the
// reinstall-local tail (may be nil). stats may be nil; when set, Tagged/Pinned/
// Pushed are incremented on successful steps.
func applyUnwindCascade(members []StackMember, flags UnwindFlags, addReinstallMainPath func(string), stats *UnwindApplyStats) error {
	if len(members) == 0 {
		return nil
	}
	plan, err := PlanUnwindCascade(members)
	if err != nil {
		return err
	}
	if plan == nil || len(plan.Steps) == 0 {
		return nil
	}

	byLabel := pickPeelMembersByLabel(members)
	nodes, _, err := buildUnwindModuleGraph(members, byLabel)
	if err != nil {
		return err
	}
	nodeByPath := make(map[string]UnwindGraphModuleNode, len(nodes))
	for _, n := range nodes {
		if n.Path != "" {
			nodeByPath[n.Path] = n
		}
	}

	// Tags created per main repo (for push).
	tagsByMain := make(map[string][]string)
	pushedMain := make(map[string]struct{})
	// addAll: top-level flag or gen-commit peeled --add-all.
	addAll := flags.AddAll || genArgsHasFlag(flags.GenCommitArgs, "--add-all")

	mainForModule := func(modPath string) (mainRepo string, label string, ok bool) {
		n, ok := nodeByPath[modPath]
		if !ok || n.RepoLabel == "" {
			return "", "", false
		}
		m, ok := byLabel[n.RepoLabel]
		if !ok {
			return "", n.RepoLabel, false
		}
		main := m.MainRepo
		if main == "" {
			main = m.Path
		}
		return storage.NormalizePath(main), n.RepoLabel, main != ""
	}

	// checkoutForEdit prefers still-present Path (linked consumer) over MainRepo.
	checkoutForEdit := func(label string) string {
		m, ok := byLabel[label]
		if !ok {
			return ""
		}
		if m.Path != "" {
			if st, err := os.Stat(m.Path); err == nil && st.IsDir() {
				return storage.NormalizePath(m.Path)
			}
		}
		main := m.MainRepo
		if main == "" {
			main = m.Path
		}
		return storage.NormalizePath(main)
	}

	moduleDirOn := func(checkout string, n UnwindGraphModuleNode) string {
		if checkout == "" {
			return ""
		}
		dir := n.Dir
		if dir == "" || dir == "." {
			return checkout
		}
		return filepath.Join(checkout, filepath.FromSlash(dir))
	}

	// remainingTouchesMain: later steps whose primary ModulePath lives on main
	// (tag-next host or pin consumer). Dep module path of a pin is not pending
	// work on the dep main once that dep has been tagged.
	remainingTouchesMain := func(from int, main string) bool {
		main = storage.NormalizePath(main)
		for _, s := range plan.Steps[from:] {
			if s.ModulePath == "" {
				continue
			}
			if m, _, ok := mainForModule(s.ModulePath); ok && m == main {
				return true
			}
		}
		return false
	}

	maybePushMain := func(main string, stepIdx int) error {
		if !flags.Push || main == "" {
			return nil
		}
		main = storage.NormalizePath(main)
		if _, ok := pushedMain[main]; ok {
			return nil
		}
		// Push when this main has no remaining cascade work and has tags to
		// publish. Pin-only consumer mains (no cascade tag) are left local —
		// matches pre-cascade peel push scope and fixtures without root origin.
		if remainingTouchesMain(stepIdx+1, main) {
			return nil
		}
		tags := tagsByMain[main]
		if len(tags) == 0 {
			return nil
		}
		fmt.Println()
		if err := runPushMain(main, false, flags.Force, tags); err != nil {
			return err
		}
		pushedMain[main] = struct{}{}
		if stats != nil {
			stats.Pushed++
		}
		return nil
	}

	recordTag := func(main, tag string) {
		if main == "" || tag == "" {
			return
		}
		main = storage.NormalizePath(main)
		for _, t := range tagsByMain[main] {
			if t == tag {
				return
			}
		}
		tagsByMain[main] = append(tagsByMain[main], tag)
	}

	for i, step := range plan.Steps {
		switch step.Kind {
		case CascadeTagNext:
			if step.ModulePath == "" || step.TagOrVersion == "" {
				continue
			}
			main, _, ok := mainForModule(step.ModulePath)
			if !ok {
				return fmt.Errorf("wrk: cascade tag-next %s: no stack main for module", step.ModulePath)
			}
			if err := requireMainActiveRoot(main, "--tag-next"); err != nil {
				return err
			}
			if err := cascadeCreateOneTag(main, step.TagOrVersion); err != nil {
				return err
			}
			fmt.Printf("tag-next %s @ %s\n", step.ModulePath, step.TagOrVersion)
			recordTag(main, step.TagOrVersion)
			if stats != nil {
				stats.Tagged++
			}
			if addReinstallMainPath != nil {
				addReinstallMainPath(main)
			}
			if err := maybePushMain(main, i); err != nil {
				return err
			}

		case CascadePin:
			if step.ModulePath == "" || step.DepModulePath == "" || step.TagOrVersion == "" {
				continue
			}
			consumerNode, ok := nodeByPath[step.ModulePath]
			if !ok {
				return fmt.Errorf("wrk: cascade pin: unknown consumer module %s", step.ModulePath)
			}
			depNode, ok := nodeByPath[step.DepModulePath]
			if !ok {
				return fmt.Errorf("wrk: cascade pin: unknown dep module %s", step.DepModulePath)
			}
			consumerMain, consumerLabel, ok := mainForModule(step.ModulePath)
			if !ok {
				return fmt.Errorf("wrk: cascade pin %s: no stack main", step.ModulePath)
			}
			_, depLabel, _ := mainForModule(step.DepModulePath)

			consumerCheckout := checkoutForEdit(consumerLabel)
			if consumerCheckout == "" {
				consumerCheckout = consumerMain
			}
			consumerModDir := moduleDirOn(consumerCheckout, consumerNode)
			if consumerModDir == "" {
				return fmt.Errorf("wrk: cascade pin: empty consumer module dir for %s", step.ModulePath)
			}

			// Dirty go.mod/go.sum without --add-all → partial edit (P3/D11):
			// save WIP → write Base → pin+tidy → selective commit → restore WIP
			// + surgical require bumps only. On failure after save: restore WIP.
			usePartial := false
			var saved goModSumSnap
			if !addAll {
				dirty, err := goModSumUncommittedAt(consumerCheckout, consumerModDir)
				if err != nil {
					return err
				}
				if dirty {
					usePartial = true
					saved, err = saveGoModSumSnap(consumerModDir)
					if err != nil {
						return fmt.Errorf("wrk: cascade pin save go.mod/go.sum in %s: %w", consumerModDir, err)
					}
					if err := writeBaseGoModSum(consumerCheckout, consumerModDir); err != nil {
						_ = restoreGoModSumSnap(consumerModDir, saved)
						return fmt.Errorf("wrk: cascade pin restore Base go.mod/go.sum in %s: %w", consumerModDir, err)
					}
				}
			}

			// Pin log: basename labels (legacy multi-module assert surface).
			logConsumer := consumerLabel
			if logConsumer == "" {
				logConsumer = filepath.Base(consumerCheckout)
			}
			logDep := depLabel
			if logDep == "" {
				logDep = filepath.Base(step.DepModulePath)
			}
			fmt.Printf("pin %s <- %s @ %s\n", logConsumer, logDep, step.TagOrVersion)

			pinFail := func(err error) error {
				if usePartial {
					_ = restoreGoModSumSnap(consumerModDir, saved)
				}
				return err
			}

			if err := cascadePinKeepLocalReplace(consumerModDir, step.DepModulePath, step.TagOrVersion, depNode, byLabel); err != nil {
				return pinFail(fmt.Errorf("wrk: cascade pin %s <- %s: %w", step.ModulePath, step.DepModulePath, err))
			}
			if err := goModTidy(consumerModDir); err != nil {
				return pinFail(fmt.Errorf("wrk: go mod tidy in %s: %w", consumerModDir, err))
			}
			_ = expandGoModRequireBlocks(filepath.Join(consumerModDir, "go.mod"))

			// Selective commit of this module's go.mod/go.sum only (D7): never
			// scoop feature WIP / --add-all extras into the pin auto-commit.
			// --add-all still gates partial-edit vs pin-on-dirty above; only
			// staging for the pin commit is forced selective.
			// Commit while WT still holds Base+pin (no WIP); restore WIP after.
			// Pass consumerModDir so nested modules stage tools/go.mod, not only root.
			if err := cascadeCommitPin(consumerCheckout, consumerModDir, step.DepModulePath, step.TagOrVersion, false); err != nil {
				return pinFail(err)
			}
			if stats != nil {
				stats.Pinned++
			}

			if usePartial {
				// Restore original WIP bytes, then surgical require bump only (no tidy).
				if err := restoreGoModSumSnap(consumerModDir, saved); err != nil {
					return fmt.Errorf("wrk: cascade pin restore WIP go.mod/go.sum in %s: %w", consumerModDir, err)
				}
				if err := goModEditRequire(consumerModDir, step.DepModulePath, step.TagOrVersion); err != nil {
					_ = restoreGoModSumSnap(consumerModDir, saved)
					return fmt.Errorf("wrk: cascade pin surgical require bump %s@%s in %s: %w",
						step.DepModulePath, step.TagOrVersion, consumerModDir, err)
				}
				// go mod edit may emit single-line require; expand to parenthesized
				// block for stable go.mod shape (matches pin path + consumers).
				_ = expandGoModRequireBlocks(filepath.Join(consumerModDir, "go.mod"))
			}

			if addReinstallMainPath != nil {
				addReinstallMainPath(consumerMain)
			}
			if err := maybePushMain(consumerMain, i); err != nil {
				// Push runs after successful pin; partial WT already restored.
				// Do not wipe surgical bump — push failure is independent of WIP.
				return err
			}
		}
	}

	// Push any tagged main not yet published (defensive).
	if flags.Push {
		for main, tags := range tagsByMain {
			if _, ok := pushedMain[main]; ok {
				continue
			}
			if len(tags) == 0 {
				continue
			}
			fmt.Println()
			if err := runPushMain(main, false, flags.Force, tags); err != nil {
				return err
			}
			pushedMain[main] = struct{}{}
			if stats != nil {
				stats.Pushed++
			}
		}
	}
	return nil
}

// cascadeCreateOneTag creates a single lightweight tag at HEAD (one scope only).
// No-op when the tag already exists at HEAD; errors if it exists elsewhere.
func cascadeCreateOneTag(mainRepo, tag string) error {
	if mainRepo == "" || tag == "" {
		return fmt.Errorf("wrk: cascade tag requires main repo and tag name")
	}
	// Already present?
	out, err := gitOutputDir(mainRepo, "rev-parse", "--verify", "--quiet", "refs/tags/"+tag)
	if err == nil && strings.TrimSpace(out) != "" {
		head, herr := gitOutputDir(mainRepo, "rev-parse", "HEAD")
		if herr == nil && strings.TrimSpace(out) == strings.TrimSpace(head) {
			return nil
		}
		// Tag exists on another commit — leave as hard error from git tag.
	}
	if err := gitRunDir(mainRepo, "tag", tag, "HEAD"); err != nil {
		return fmt.Errorf("wrk: cascade tag %s: %w", tag, err)
	}
	return nil
}

// cascadePinKeepLocalReplace bumps require for depModule to version. Keeps an
// existing intra-repo local filesystem replace; drops external/out-of-repo
// replaces (multi-repo pin path) so tidy can resolve published versions.
func cascadePinKeepLocalReplace(consumerModDir, depModule, version string, depNode UnwindGraphModuleNode, byLabel map[string]StackMember) error {
	_ = depNode
	_ = byLabel
	if consumerModDir == "" || depModule == "" || version == "" {
		return fmt.Errorf("consumer dir, dep module, and version required")
	}
	// Snapshot whether an intra-repo local replace exists before edits.
	keepReplace, replaceNew := localReplacePolicy(consumerModDir, depModule)

	opts := &commands.GoModEditOptions{Dir: consumerModDir, Stderr: false, Stdout: false}
	if !keepReplace {
		// Drop external / non-intra replace so require can resolve via proxy/tags.
		_ = commands.GoModDropReplace(depModule, opts)
	}
	if err := commands.GoModEditRequire(depModule, version, opts); err != nil {
		return err
	}
	// If something stripped an intra-repo replace, restore it.
	if keepReplace && replaceNew != "" && !goModHasLocalReplace(consumerModDir, depModule) {
		if err := commands.GoModEditReplace(depModule, replaceNew, opts); err != nil {
			return fmt.Errorf("restore local replace %s => %s: %w", depModule, replaceNew, err)
		}
	}
	return nil
}

// localReplacePolicy reports whether depModule's replace should be kept and the
// replace NewPath when present. Intra-repo filesystem replaces are kept;
// external worktree / absolute-out-of-repo / module-path replaces are not.
func localReplacePolicy(consumerModDir, depModule string) (keep bool, newPath string) {
	goModPath := filepath.Join(consumerModDir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return false, ""
	}
	f, err := modfile.Parse(goModPath, data, nil)
	if err != nil || f == nil {
		// Tolerant scrape: replace lines with => .
		return scrapeLocalReplacePolicy(string(data), depModule, consumerModDir)
	}
	var replNew string
	found := false
	for _, r := range f.Replace {
		if r.Old.Path == depModule {
			replNew = r.New.Path
			found = true
			break
		}
	}
	if !found || replNew == "" {
		return false, ""
	}
	return classifyLocalReplace(consumerModDir, replNew)
}

func scrapeLocalReplacePolicy(content, depModule, consumerModDir string) (keep bool, newPath string) {
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if !strings.Contains(line, depModule) || !strings.Contains(line, "=>") {
			continue
		}
		// replace dep => path  OR  dep => path inside block
		parts := strings.SplitN(line, "=>", 2)
		if len(parts) != 2 {
			continue
		}
		left := strings.TrimSpace(parts[0])
		left = strings.TrimPrefix(left, "replace ")
		fields := strings.Fields(left)
		if len(fields) == 0 || fields[0] != depModule {
			continue
		}
		right := strings.Fields(strings.TrimSpace(parts[1]))
		if len(right) == 0 {
			continue
		}
		return classifyLocalReplace(consumerModDir, right[0])
	}
	return false, ""
}

func classifyLocalReplace(consumerModDir, replNew string) (keep bool, newPath string) {
	if replNew == "" {
		return false, ""
	}
	// Module-path replace (not filesystem): drop so pin can use version.
	if !strings.HasPrefix(replNew, ".") && !filepath.IsAbs(replNew) {
		return false, replNew
	}
	abs, err := resolveLocalReplacePath(consumerModDir, replNew)
	if err != nil {
		return false, replNew
	}
	// Target removed (e.g. --done dropped nested external worktree) → drop replace.
	if st, err := os.Stat(abs); err != nil || !st.IsDir() {
		return false, replNew
	}
	consumerTop, err := worktree.ShowToplevel(consumerModDir)
	if err != nil {
		return strings.HasPrefix(abs, consumerModDir), replNew
	}
	consumerTop = storage.NormalizePath(consumerTop)
	// Same git toplevel as consumer → intra-repo (./pkgs/shared) → keep.
	if targetTop, err := worktree.ShowToplevel(abs); err == nil {
		if storage.NormalizePath(targetTop) == consumerTop {
			return true, replNew
		}
		// Different toplevel (nested external leaf checkout) → drop for pin.
		return false, replNew
	}
	// Non-git target under consumer tree → keep.
	absN := storage.NormalizePath(abs)
	if absN == consumerTop || strings.HasPrefix(absN, consumerTop+string(filepath.Separator)) {
		return true, replNew
	}
	return false, replNew
}

// goModHasLocalReplace reports a replace directive for modulePath in dir's go.mod.
func goModHasLocalReplace(modDir, modulePath string) bool {
	data, err := os.ReadFile(filepath.Join(modDir, "go.mod"))
	if err != nil {
		return false
	}
	content := string(data)
	if strings.Contains(content, "replace "+modulePath) {
		return true
	}
	return strings.Contains(content, modulePath+" =>")
}

// goModSumUncommitted reports uncommitted changes to go.mod or go.sum under repo root.
func goModSumUncommitted(repo string) (bool, error) {
	out, err := gitOutputDir(repo, "status", "--porcelain", "--", "go.mod", "go.sum")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// goModSumUncommittedAt reports uncommitted go.mod/go.sum for a module dir under checkout.
// Paths are relative to the git toplevel (supports nested modules).
func goModSumUncommittedAt(checkout, modDir string) (bool, error) {
	if checkout == "" || modDir == "" {
		return goModSumUncommitted(checkout)
	}
	modRel, sumRel, err := goModSumRelPaths(checkout, modDir)
	if err != nil {
		// Fall back to root paths when Rel fails.
		return goModSumUncommitted(checkout)
	}
	out, err := gitOutputDir(checkout, "status", "--porcelain", "--", modRel, sumRel)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// goModSumSnap is a byte snapshot of module go.mod (+ optional go.sum) for partial edit.
type goModSumSnap struct {
	mod    []byte
	sum    []byte
	hasSum bool
}

// saveGoModSumSnap reads current WT go.mod and go.sum (if present) from modDir.
func saveGoModSumSnap(modDir string) (goModSumSnap, error) {
	var s goModSumSnap
	mod, err := os.ReadFile(filepath.Join(modDir, "go.mod"))
	if err != nil {
		return s, err
	}
	s.mod = mod
	sumPath := filepath.Join(modDir, "go.sum")
	if b, err := os.ReadFile(sumPath); err == nil {
		s.sum = b
		s.hasSum = true
	}
	return s, nil
}

// restoreGoModSumSnap writes saved go.mod/go.sum bytes back to modDir.
// When go.sum was absent at save time, removes go.sum if present.
func restoreGoModSumSnap(modDir string, s goModSumSnap) error {
	if err := os.WriteFile(filepath.Join(modDir, "go.mod"), s.mod, 0o644); err != nil {
		return err
	}
	sumPath := filepath.Join(modDir, "go.sum")
	if s.hasSum {
		return os.WriteFile(sumPath, s.sum, 0o644)
	}
	_ = os.Remove(sumPath)
	return nil
}

// goModSumRelPaths returns checkout-relative paths for go.mod and go.sum under modDir.
func goModSumRelPaths(checkout, modDir string) (modRel, sumRel string, err error) {
	modAbs := filepath.Join(modDir, "go.mod")
	modRel, err = filepath.Rel(checkout, modAbs)
	if err != nil {
		return "", "", err
	}
	sumRel = filepath.Join(filepath.Dir(modRel), "go.sum")
	// git pathspecs prefer slash form; filepath is fine on macOS/Linux.
	return modRel, sumRel, nil
}

// writeBaseGoModSum writes Base (index if staged, else HEAD) go.mod/go.sum into WT.
// go.mod is required at Base. go.sum is optional — missing Base go.sum removes WT go.sum.
func writeBaseGoModSum(checkout, modDir string) error {
	modRel, sumRel, err := goModSumRelPaths(checkout, modDir)
	if err != nil {
		return err
	}
	if err := writeBaseBlobToFile(checkout, modRel, filepath.Join(modDir, "go.mod")); err != nil {
		return err
	}
	sumDest := filepath.Join(modDir, "go.sum")
	if err := writeBaseBlobToFile(checkout, sumRel, sumDest); err != nil {
		// go.sum often absent on minimal fixtures; drop any WIP go.sum for true Base.
		_ = os.Remove(sumDest)
		return nil
	}
	return nil
}

// writeBaseBlobToFile writes index-or-HEAD content of rel into dest.
func writeBaseBlobToFile(checkout, rel, dest string) error {
	data, err := readBaseBlob(checkout, rel)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o644)
}

// readBaseBlob returns file bytes from the index when rel is staged, else HEAD.
func readBaseBlob(checkout, rel string) ([]byte, error) {
	relSlash := filepath.ToSlash(rel)
	// Staged?
	cached, err := gitOutputDir(checkout, "diff", "--cached", "--name-only", "--", rel)
	src := "HEAD:" + relSlash
	if err == nil && strings.TrimSpace(cached) != "" {
		// Index stage 0: :path
		src = ":" + relSlash
	}
	// Use raw Output so trailing newlines on go.mod/go.sum are preserved.
	cmd := gitCommandDir(checkout, "show", src)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git show %s: %w", src, err)
	}
	return out, nil
}


// cascadeCommitPin stages the consumer module's go.mod/go.sum (paths relative to
// checkout — nested e.g. tools/go.mod) or -A with addAll, then commits with locked
// subject prefix "wrk: cascade pin <mod> @ <ver>". No-op when nothing to commit.
//
// Selective path (addAll=false) uses `git commit --only` so a pre-staged index of
// feature WIP is never scooped into the pin commit (D7). Plain `git add go.mod`
// + `git commit` would commit the entire index.
func cascadeCommitPin(repo, modDir, depModule, ver string, addAll bool) error {
	if repo == "" {
		return nil
	}
	msg := "wrk: cascade pin " + depModule + " @ " + ver

	if addAll {
		if err := gitRunDir(repo, "add", "-A"); err != nil {
			return fmt.Errorf("wrk: cascade pin git add -A: %w", err)
		}
		staged, err := gitOutputDir(repo, "diff", "--cached", "--name-only")
		if err != nil {
			return fmt.Errorf("wrk: cascade pin check staged: %w", err)
		}
		if strings.TrimSpace(staged) == "" {
			return nil
		}
		// --no-verify: cascade pin is a tool-authored deps commit; user hooks that
		// scan removed external worktrees must not block free-module cascade.
		if err := gitRunDir(repo, "commit", "--no-verify", "-m", msg); err != nil {
			return fmt.Errorf("wrk: cascade pin commit: %w", err)
		}
		return nil
	}

	// Selective: only this consumer module's go.mod/go.sum.
	// Paths are checkout-relative so nested modules (tools/) stage correctly.
	modRel, sumRel := "go.mod", "go.sum"
	if modDir != "" {
		if r, s, err := goModSumRelPaths(repo, modDir); err == nil {
			modRel, sumRel = r, s
		}
	}
	paths := []string{modRel}
	sumPath := filepath.Join(repo, sumRel)
	if modDir != "" {
		sumPath = filepath.Join(modDir, "go.sum")
	}
	if _, err := os.Stat(sumPath); err == nil {
		paths = append(paths, sumRel)
	}
	// Stage pin paths so the index reflects go.mod/go.sum edits for --only.
	for _, p := range paths {
		_ = gitRunDir(repo, "add", "--", p)
	}
	diffArgs := append([]string{"diff", "--cached", "--name-only", "--"}, paths...)
	staged, err := gitOutputDir(repo, diffArgs...)
	if err != nil {
		return fmt.Errorf("wrk: cascade pin check staged: %w", err)
	}
	if strings.TrimSpace(staged) == "" {
		return nil
	}
	// --only: commit these paths only; ignore any other already-staged WIP.
	// --no-verify: tool-authored deps commit (see addAll branch).
	commitArgs := append([]string{"commit", "--only", "--no-verify", "-m", msg, "--"}, paths...)
	if err := gitRunDir(repo, commitArgs...); err != nil {
		return fmt.Errorf("wrk: cascade pin commit: %w", err)
	}
	return nil
}
