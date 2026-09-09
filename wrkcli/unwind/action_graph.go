package unwind

import (
	"fmt"
	"sort"
	"strings"
)

// Action graph IR (Go build-action inspired). Preview + dry-run share this DAG;
// apply still bridges through existing peel/cascade helpers (M1/M2).

// ArtifactKind is a typed output flowing between actions.
type ArtifactKind string

const (
	ArtTag     ArtifactKind = "tag"
	ArtVersion ArtifactKind = "version"
	ArtTree    ArtifactKind = "tree"
)

// Artifact is produced or consumed by actions.
type Artifact struct {
	ID      string       `json:"id"`
	Kind    ArtifactKind `json:"kind"`
	Module  string       `json:"module,omitempty"`
	Repo    string       `json:"repo,omitempty"`
	Ref     string       `json:"ref,omitempty"`
	Version string       `json:"version,omitempty"`
}

// SubjectKind identifies what an action operates on.
type SubjectKind string

const (
	SubCheckout SubjectKind = "checkout"
	SubModule   SubjectKind = "module"
	SubRepo     SubjectKind = "repo"
)

// Subject is the generalized target of an action (Go's Package analogue).
type Subject struct {
	Kind      SubjectKind `json:"kind"`
	ID        string      `json:"id"`
	Display   string      `json:"display"`
	RepoLabel string      `json:"repo_label,omitempty"`
	Module    string      `json:"module,omitempty"`
	Linked    bool        `json:"linked,omitempty"`
}

// ActionMode is the primitive operation (Go's Mode analogue).
type ActionMode string

const (
	ModeGenCommit ActionMode = "gen-commit"
	ModeMergeBack ActionMode = "merge-back"
	ModeDone      ActionMode = "done"
	ModeTagNext   ActionMode = "tag-next"
	ModePin       ActionMode = "pin"
	// ModeDepUpdate is a cross-repo pin of a newly tagged (or drifted) dep.
	// Same payload as pin; Phase 1 preview uses this name.
	ModeDepUpdate ActionMode = "dep-update"
	ModeCommit    ActionMode = "commit"
	ModePush      ActionMode = "push"
	ModeSync      ActionMode = "sync"
	ModeReinstall ActionMode = "reinstall-local"
)

// ActionReason explains why the action was included.
type ActionReason string

const (
	ReasonLand         ActionReason = "land"
	ReasonOwnedRelease ActionReason = "owned-release"
	ReasonPropagate    ActionReason = "propagate"
	ReasonPinBeforeTag ActionReason = "pin-before-tag"
	ReasonLatestDrift  ActionReason = "latest-drift"
	ReasonDropReplace  ActionReason = "drop-replace"
	ReasonShip         ActionReason = "ship"
)

// Action is one node in the unwind action DAG.
type Action struct {
	ID      string       `json:"id"`
	Mode    ActionMode   `json:"mode"`
	Subject Subject      `json:"subject"`
	Reason  ActionReason `json:"reason"`
	Detail  string       `json:"detail,omitempty"`
	Why     string       `json:"why,omitempty"`

	Deps     []string `json:"deps"`
	Consumes []string `json:"consumes,omitempty"`
	Produces []string `json:"produces,omitempty"`

	DepModule  string `json:"dep_module,omitempty"`
	PinVersion string `json:"pin_version,omitempty"`

	// Rank is longest-path depth for layout (0 = roots).
	Rank int `json:"rank"`
	// Lane is swimlane key (usually repo label).
	Lane string `json:"lane,omitempty"`
}

// ActionWave is a derived band label for UI (not the IR source of truth).
type ActionWave struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	RankMin   int      `json:"rank_min"`
	RankMax   int      `json:"rank_max"`
	ActionIDs []string `json:"action_ids"`
}

// ActionGraph is the plan DAG for preview / debug / (later) apply.
type ActionGraph struct {
	WorkDir   string               `json:"work_dir"`
	Cleanup   bool                 `json:"cleanup"`
	Actions   []*Action            `json:"actions"`
	Artifacts map[string]*Artifact `json:"artifacts"`
	// LaneLevels is project topo level (0 = root deps) for band rank layout.
	LaneLevels map[string]int `json:"lane_levels,omitempty"`
	Waves      []ActionWave   `json:"waves,omitempty"`
	Excluded   int            `json:"excluded,omitempty"` // cleanup actions omitted
	FilterNote string         `json:"filter_note,omitempty"`
}

// ActionGraphOpts controls BuildActionGraph.
type ActionGraphOpts struct {
	// Cleanup includes latest-drift and drop-replace pins.
	Cleanup bool
}

func defaultReasons(cleanup bool) map[ActionReason]bool {
	m := map[ActionReason]bool{
		ReasonLand:         true,
		ReasonOwnedRelease: true,
		ReasonPropagate:    true,
		ReasonPinBeforeTag: true,
		ReasonShip:         true,
	}
	if cleanup {
		m[ReasonLatestDrift] = true
		m[ReasonDropReplace] = true
	}
	return m
}

func pinLike(mode ActionMode) bool {
	return mode == ModePin || mode == ModeDepUpdate
}

// BuildActionGraph builds the unwind action DAG from a snapshot and flags.
func BuildActionGraph(snap *Snapshot, flags UnwindFlags, opts ActionGraphOpts) *ActionGraph {
	g := &ActionGraph{
		Artifacts: map[string]*Artifact{},
		Cleanup:   opts.Cleanup,
	}
	if snap == nil {
		return g
	}
	g.WorkDir = snap.WorkDir
	allow := defaultReasons(opts.Cleanup)

	plan := snap.Peel
	if plan == nil {
		plan = &UnwindPlan{}
	}
	members := snap.Inv.Members
	byLabel := pickPeelMembersByLabel(members)
	cascade := snap.Cascade
	if cascade == nil {
		cascade = &UnwindCascadePlan{}
	}
	nodes := snap.ModuleNodes
	edges := snap.ModuleEdges
	nodeByPath := make(map[string]UnwindGraphModuleNode, len(nodes))
	for _, n := range nodes {
		if n.Path != "" {
			nodeByPath[n.Path] = n
		}
	}

	dirty := dirtyLabels(members)
	willTag := make(map[string]struct{})
	for _, n := range nodes {
		if !cascadeModuleShouldTag(n) {
			continue
		}
		if n.RepoLabel != "" && !dirty[n.RepoLabel] {
			continue
		}
		willTag[n.Path] = struct{}{}
	}
	droppable := droppableExternalPairs(nodes, edges)

	var early, deferred []string
	if flags.TagNext {
		early, deferred = splitPeelOrderB1(plan.PeelOrder, members, cascade, nodes, edges)
	} else {
		early = append([]string(nil), plan.PeelOrder...)
	}
	deferredSet := make(map[string]struct{}, len(deferred))
	for _, lab := range deferred {
		deferredSet[lab] = struct{}{}
	}

	actions := make([]*Action, 0, 32)
	byID := map[string]*Action{}
	add := func(a *Action) {
		if a == nil {
			return
		}
		if _, ok := byID[a.ID]; ok {
			return
		}
		byID[a.ID] = a
		actions = append(actions, a)
	}
	link := func(fromID, toID string) {
		to := byID[toID]
		if to == nil || fromID == "" {
			return
		}
		for _, d := range to.Deps {
			if d == fromID {
				return
			}
		}
		to.Deps = append(to.Deps, fromID)
	}

	// --- land actions (no peel banner) ---
	firstLandByLabel := map[string]string{} // repo label → first land action id
	lastLandByLabel := map[string]string{}  // repo label → last land action id
	emitLand := func(labels []string) {
		for _, label := range labels {
			m, ok := byLabel[label]
			display := label
			if ok {
				display = peelDisplayPath(snap.WorkDir, m.Path)
			}
			subj := Subject{
				Kind: SubCheckout, ID: "checkout:" + label, Display: display,
				RepoLabel: label, Linked: ok && m.Linked,
			}
			var prev string
			if flags.GenCommitMsg {
				id := "gen-commit:" + label
				add(&Action{
					ID: id, Mode: ModeGenCommit, Subject: subj,
					Reason: ReasonLand, Why: "feature work on dirty checkout",
					Lane: label,
				})
				prev = id
			}
			if ok && m.Linked && (flags.Done || flags.MergeBack) {
				id := "merge-back:" + label
				mode := ModeMergeBack
				detail := "keep worktree"
				if flags.Done {
					id = "done:" + label
					mode = ModeDone
					detail = "merge-back and remove worktree"
				}
				a := &Action{
					ID: id, Mode: mode, Subject: subj,
					Reason: ReasonLand, Detail: detail, Why: "linked worktree",
					Lane: label,
				}
				artID := "tree:" + label
				g.Artifacts[artID] = &Artifact{ID: artID, Kind: ArtTree, Repo: label}
				a.Produces = []string{artID}
				add(a)
				if prev != "" {
					link(prev, id)
				}
				prev = id
			}
			if prev != "" {
				if firstLandByLabel[label] == "" {
					// First emitted id is gen-commit when present, else merge/done.
					if flags.GenCommitMsg {
						firstLandByLabel[label] = "gen-commit:" + label
					} else {
						firstLandByLabel[label] = prev
					}
				}
				lastLandByLabel[label] = prev
			}
		}
	}
	emitLand(early)
	emitLand(deferred)

	// --- cascade: classify + filter ---
	excluded := 0
	tagActionByModule := map[string]string{}

	type classifiedPin struct {
		step   UnwindCascadeStep
		reason ActionReason
	}
	var pins []classifiedPin
	var tags []UnwindCascadeStep

	reqVer := map[string]string{} // consumer\x00dep → require version
	for _, e := range edges {
		if e.Kind == "require" && e.From != "" && e.To != "" {
			reqVer[e.From+"\x00"+e.To] = e.Version
		}
	}

	for _, s := range cascade.Steps {
		switch s.Kind {
		case CascadeTagNext:
			tags = append(tags, s)
		case CascadePin:
			_, depTags := willTag[s.DepModulePath]
			_, consTags := willTag[s.ModulePath]
			_, drop := droppable[s.ModulePath+"\x00"+s.DepModulePath]
			cur := reqVer[s.ModulePath+"\x00"+s.DepModulePath]
			drift := cur != "" && s.TagOrVersion != "" && !versionsMatch(cur, s.TagOrVersion)
			var reason ActionReason
			switch {
			case depTags:
				reason = ReasonPropagate
			case drop && !drift:
				reason = ReasonDropReplace
			case drift:
				// Catch-up to LatestTag of an untagged dep is latest-drift even
				// when the consumer itself is tagging (don't promote to pin-before-tag).
				reason = ReasonLatestDrift
			case consTags:
				reason = ReasonPinBeforeTag
			default:
				reason = ReasonLatestDrift
			}
			if !allow[reason] {
				excluded++
				continue
			}
			pins = append(pins, classifiedPin{step: s, reason: reason})
		}
	}
	if !flags.TagNext {
		tags = nil
		pins = nil
	}

	for _, s := range tags {
		if !allow[ReasonOwnedRelease] {
			excluded++
			continue
		}
		n := nodeByPath[s.ModulePath]
		if n.RepoLabel != "" && !dirty[n.RepoLabel] {
			continue
		}
		lane := n.RepoLabel
		if lane == "" {
			lane = s.ModulePath
		}
		id := "tag-next:" + s.ModulePath
		artTag := "tag:" + s.ModulePath
		artVer := "version:" + s.ModulePath
		ver := goRequireVersionFromTag(s.TagOrVersion)
		g.Artifacts[artTag] = &Artifact{ID: artTag, Kind: ArtTag, Module: s.ModulePath, Ref: s.TagOrVersion, Version: ver, Repo: lane}
		g.Artifacts[artVer] = &Artifact{ID: artVer, Kind: ArtVersion, Module: s.ModulePath, Version: ver, Repo: lane}
		a := &Action{
			ID: id, Mode: ModeTagNext,
			Subject:  Subject{Kind: SubModule, ID: "module:" + s.ModulePath, Display: s.ModulePath, Module: s.ModulePath, RepoLabel: lane},
			Reason:   ReasonOwnedRelease,
			Detail:   s.TagOrVersion,
			Why:      "owned changes",
			Produces: []string{artTag, artVer},
			Lane:     lane,
		}
		add(a)
		tagActionByModule[s.ModulePath] = id
		if land := lastLandByLabel[lane]; land != "" {
			link(land, id)
		}
	}

	emitPinAction := func(consPath, depPath, ver string, reason ActionReason) {
		if consPath == "" || depPath == "" {
			return
		}
		id := fmt.Sprintf("pin:%s<%s", consPath, depPath)
		if byID[id] != nil {
			return
		}
		n := nodeByPath[consPath]
		lane := n.RepoLabel
		if lane == "" {
			lane = consPath
		}
		mode := ModePin
		if !pinIsIntraRepo(nodeByPath, consPath, depPath) {
			mode = ModeDepUpdate
		}
		a := &Action{
			ID: id, Mode: mode,
			Subject:    Subject{Kind: SubModule, ID: "module:" + consPath, Display: consPath, Module: consPath, RepoLabel: lane},
			Reason:     reason,
			Detail:     fmt.Sprintf("<- %s @ %s", depPath, ver),
			Why:        string(reason),
			DepModule:  depPath,
			PinVersion: ver,
			Lane:       lane,
		}
		if _, ok := tagActionByModule[depPath]; ok {
			a.Consumes = append(a.Consumes, "version:"+depPath)
		}
		add(a)
		// Prefer dep tag-next as the release gate (not merge-back). Land edges
		// are added only when not already covered by that gate.
		if tagID, ok := tagActionByModule[depPath]; ok {
			link(tagID, id)
		}
		linkLandIfUncovered := func(landID string) {
			if landID == "" || depsReach(byID, id, landID) {
				return
			}
			link(landID, id)
		}
		if mode != ModeDepUpdate {
			if _, def := deferredSet[lane]; def {
				// Deferred consumer: wait on early peels not already implied by
				// the dep gate — never on this lane's own land (B1/D7). Pin-before-tag
				// of an untagged nested module does not need those peels; linking
				// them would make commit transitively wait on early merge-back.
				if reason != ReasonPinBeforeTag {
					for _, earlyLab := range early {
						linkLandIfUncovered(lastLandByLabel[earlyLab])
					}
				}
			} else {
				// Early lane: land this checkout before pin when not covered by gate.
				linkLandIfUncovered(lastLandByLabel[lane])
			}
		}
		// pin-before-tag: pins on M before tag-next(M). Skip the chord when
		// land already chains pin → firstLand → lastLand → tag-next (same
		// noise as merge-back → deferred commit).
		if tagID, ok := tagActionByModule[consPath]; ok {
			_, deferredLane := deferredSet[lane]
			willGateLand := firstLandByLabel[lane] != "" && (mode == ModeDepUpdate || deferredLane)
			if lastLandByLabel[lane] == "" || !willGateLand {
				link(id, tagID)
			}
		}
	}

	for _, cp := range pins {
		s := cp.step
		emitPinAction(s.ModulePath, s.DepModulePath, s.TagOrVersion, cp.reason)
	}
	// Tagged cross-repo require/replace with no cascade pin still needs
	// dep-update so consumer commit does not hang off tag-next directly.
	if flags.TagNext && allow[ReasonPropagate] {
		for _, e := range edges {
			if e.Kind != "require" && e.Kind != "replace" {
				continue
			}
			if tagActionByModule[e.To] == "" {
				continue
			}
			if pinIsIntraRepo(nodeByPath, e.From, e.To) {
				continue
			}
			from := nodeByPath[e.From]
			if firstLandByLabel[from.RepoLabel] == "" {
				continue
			}
			ver := nodeByPath[e.To].NextTag
			if ver == "" {
				ver = e.Version
			}
			emitPinAction(e.From, e.To, ver, ReasonPropagate)
		}
	}

	// Deferred gen-commit waits on pin/dep-update into that lane (UI expand
	// moves these onto the commit card). Do not also wait on early merge-back:
	// commit waits on the pin, not on the dep's land.
	for _, lab := range deferred {
		first := firstLandByLabel[lab]
		if first == "" {
			continue
		}
		for _, a := range actions {
			if !pinLike(a.Mode) || a.Lane != lab {
				continue
			}
			link(a.ID, first)
		}
	}
	// Cross-repo dep-update always gates that lane's gen-commit (including
	// early peels / missing cascade pins that B1 would not see).
	for _, a := range actions {
		if a.Mode != ModeDepUpdate {
			continue
		}
		if first := firstLandByLabel[a.Lane]; first != "" {
			link(a.ID, first)
		}
	}

	// Cross-repo release gate: consumer commit waits on dep-update (pin alias)
	// of a tagged dep. Direct tag-next → commit is only a fallback when no
	// pin-like covers that require. Pins wait on the same gate, except
	// same-lane lastLand (merge/done): that land waits on gen-commit which
	// (deferred) waits on the pin — a cycle.
	gateOfModule := func(modPath string) string {
		if id := tagActionByModule[modPath]; id != "" {
			return id
		}
		n := nodeByPath[modPath]
		lab := n.RepoLabel
		if lab == "" {
			return ""
		}
		return lastLandByLabel[lab]
	}
	for _, e := range edges {
		if e.Kind != "require" && e.Kind != "replace" {
			continue
		}
		from, okF := nodeByPath[e.From]
		to, okT := nodeByPath[e.To]
		if !okF || !okT {
			continue
		}
		consLab, depLab := from.RepoLabel, to.RepoLabel
		if consLab == "" || depLab == "" || consLab == depLab {
			continue
		}
		if firstLandByLabel[consLab] == "" || lastLandByLabel[depLab] == "" {
			continue
		}
		gate := gateOfModule(e.To)
		if gate == "" {
			continue
		}
		// Prefer dep-update/pin as the intermediate; skip tag-next → commit.
		if a := byID[fmt.Sprintf("pin:%s<%s", e.From, e.To)]; a != nil && pinLike(a.Mode) {
			continue
		}
		link(gate, firstLandByLabel[consLab])
	}
	for _, a := range actions {
		if !pinLike(a.Mode) || a.DepModule == "" {
			continue
		}
		gate := gateOfModule(a.DepModule)
		if gate == "" {
			continue
		}
		if own := lastLandByLabel[a.Lane]; own != "" && gate == own {
			continue
		}
		link(gate, a.ID)
	}

	// Per-project ship (slow band): after that project's fast release.
	peelLabels := append(append([]string{}, early...), deferred...)
	for _, lab := range peelLabels {
		last := lastLandByLabel[lab]
		if last == "" {
			continue
		}
		m, ok := byLabel[lab]
		display := lab
		if ok {
			display = peelDisplayPath(snap.WorkDir, m.Path)
		}
		subj := Subject{
			Kind: SubCheckout, ID: "checkout:" + lab, Display: display,
			RepoLabel: lab, Linked: ok && m.Linked,
		}
		var preds []string
		for mod, tid := range tagActionByModule {
			if nodeByPath[mod].RepoLabel == lab {
				preds = append(preds, tid)
			}
		}
		if len(preds) == 0 {
			preds = []string{last}
		}
		// reinstall-local stays on the cross-repo lane (after tag).
		// push/sync are ship-phase only; they still sit in the full graph
		// so buildJobPhases can slice them out.
		if flags.ReinstallLocal {
			id := "reinstall:" + lab
			add(&Action{
				ID: id, Mode: ModeReinstall, Subject: subj,
				Reason: ReasonShip, Why: "local binaries",
				Lane: lab,
			})
			for _, p := range preds {
				link(p, id)
			}
		}
		var prevShip string
		if flags.Push {
			id := "push:" + lab
			add(&Action{
				ID: id, Mode: ModePush, Subject: subj,
				Reason: ReasonShip, Why: "publish branch and tags",
				Lane: lab,
			})
			for _, p := range preds {
				link(p, id)
			}
			prevShip = id
		}
		if flags.Sync {
			id := "sync:" + lab
			add(&Action{
				ID: id, Mode: ModeSync, Subject: subj,
				Reason: ReasonShip, Why: "fast-forward after merge-back",
				Lane: lab,
			})
			if prevShip != "" {
				link(prevShip, id)
			} else {
				for _, p := range preds {
					link(p, id)
				}
			}
		}
	}

	levels := projectLevels(peelLabels, nodes, edges)
	g.LaneLevels = levels
	assignRanksByBand(actions, levels)
	stretchRanksFromBand(actions)
	g.Actions = actions
	g.Excluded = excluded
	if excluded > 0 && !opts.Cleanup {
		g.FilterNote = fmt.Sprintf("excluded %d latest-drift/drop-replace pins (enable cleanup to include)", excluded)
	}
	g.Waves = deriveWaves(actions)
	return g
}

// depsReach reports whether action id transitively depends on targetID.
func depsReach(byID map[string]*Action, id, targetID string) bool {
	if id == "" || targetID == "" {
		return false
	}
	seen := map[string]bool{}
	var walk func(string) bool
	walk = func(cur string) bool {
		if cur == targetID {
			return true
		}
		if seen[cur] {
			return false
		}
		seen[cur] = true
		a := byID[cur]
		if a == nil {
			return false
		}
		for _, d := range a.Deps {
			if walk(d) {
				return true
			}
		}
		return false
	}
	return walk(id)
}

// projectLevels assigns topo levels on peel lanes from require/replace edges.
// Dep has lower level than consumer (roots = 0).
func projectLevels(peelLabels []string, nodes []UnwindGraphModuleNode, edges []UnwindGraphModuleEdge) map[string]int {
	out := make(map[string]int, len(peelLabels))
	for _, lab := range peelLabels {
		out[lab] = 0
	}
	if len(peelLabels) == 0 {
		return out
	}
	labelOf := make(map[string]string, len(nodes))
	for _, n := range nodes {
		if n.Path != "" && n.RepoLabel != "" {
			labelOf[n.Path] = n.RepoLabel
		}
	}
	// longest path: consumer level = max(dep level)+1
	type edge struct{ from, to string } // from=consumer, to=dep
	var es []edge
	for _, e := range edges {
		if e.Kind != "require" && e.Kind != "replace" {
			continue
		}
		cf, ct := labelOf[e.From], labelOf[e.To]
		if cf == "" || ct == "" || cf == ct {
			continue
		}
		if _, ok := out[cf]; !ok {
			continue
		}
		if _, ok := out[ct]; !ok {
			continue
		}
		es = append(es, edge{from: cf, to: ct})
	}
	changed := true
	for changed {
		changed = false
		for _, e := range es {
			need := out[e.to] + 1
			if need > out[e.from] {
				out[e.from] = need
				changed = true
			}
		}
	}
	return out
}

// assignRanksByBand sets layout rank floors: prep=0, fast release=1+L, slow ship=2+L.
// ModeGenCommit is prep (UI expands commit into the release band).
func assignRanksByBand(actions []*Action, levels map[string]int) {
	for _, a := range actions {
		L := 0
		if levels != nil {
			L = levels[a.Lane]
		}
		switch a.Mode {
		case ModeGenCommit:
			a.Rank = 0
		case ModeMergeBack, ModeDone, ModePin, ModeDepUpdate, ModeTagNext:
			a.Rank = 1 + L
		case ModePush, ModeSync, ModeReinstall:
			a.Rank = 2 + L
		default:
			a.Rank = 1 + L
		}
	}
}

// stretchRanksFromBand raises non-prep ranks so each action sits after its
// non-prep deps (max(floor, max(dep)+1)). Prep stays at 0 for UI expand.
// Iteration is capped so a dependency cycle cannot spin forever.
func stretchRanksFromBand(actions []*Action) {
	byID := make(map[string]*Action, len(actions))
	for _, a := range actions {
		byID[a.ID] = a
	}
	isPrep := func(a *Action) bool {
		return a != nil && a.Mode == ModeGenCommit
	}
	maxIters := len(actions) + 2
	if maxIters < 4 {
		maxIters = 4
	}
	changed := true
	for iters := 0; changed && iters < maxIters; iters++ {
		changed = false
		for _, a := range actions {
			if isPrep(a) {
				continue
			}
			best := a.Rank
			for _, d := range a.Deps {
				dep := byID[d]
				if isPrep(dep) || dep == nil {
					continue
				}
				if need := dep.Rank + 1; need > best {
					best = need
				}
			}
			if best > a.Rank {
				a.Rank = best
				changed = true
			}
		}
	}
}

func droppableExternalPairs(nodes []UnwindGraphModuleNode, edges []UnwindGraphModuleEdge) map[string]struct{} {
	nodeByPath := make(map[string]UnwindGraphModuleNode, len(nodes))
	for _, n := range nodes {
		if n.Path != "" {
			nodeByPath[n.Path] = n
		}
	}
	out := make(map[string]struct{})
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
			out[e.From+"\x00"+e.To] = struct{}{}
		}
	}
	return out
}

func assignRanks(actions []*Action) {
	byID := make(map[string]*Action, len(actions))
	for _, a := range actions {
		byID[a.ID] = a
		a.Rank = -1
	}
	var rankOf func(string) int
	rankOf = func(id string) int {
		a := byID[id]
		if a == nil {
			return 0
		}
		if a.Rank >= 0 {
			return a.Rank
		}
		best := 0
		for _, d := range a.Deps {
			r := rankOf(d) + 1
			if r > best {
				best = r
			}
		}
		a.Rank = best
		return best
	}
	for _, a := range actions {
		_ = rankOf(a.ID)
	}
}

func deriveWaves(actions []*Action) []ActionWave {
	if len(actions) == 0 {
		return nil
	}
	type bucket struct {
		id, title string
		min, max  int
		ids       []string
	}
	// Classify by mode into coarse bands.
	land := bucket{id: "land", title: "land", min: 1 << 30, max: -1}
	release := bucket{id: "release", title: "release", min: 1 << 30, max: -1}
	ship := bucket{id: "ship", title: "ship", min: 1 << 30, max: -1}
	add := func(b *bucket, a *Action) {
		b.ids = append(b.ids, a.ID)
		if a.Rank < b.min {
			b.min = a.Rank
		}
		if a.Rank > b.max {
			b.max = a.Rank
		}
	}
	for _, a := range actions {
		switch a.Mode {
		case ModeGenCommit, ModeMergeBack, ModeDone:
			add(&land, a)
		case ModeTagNext, ModePin, ModeDepUpdate:
			add(&release, a)
		default:
			add(&ship, a)
		}
	}
	var out []ActionWave
	for _, b := range []bucket{land, release, ship} {
		if len(b.ids) == 0 {
			continue
		}
		sort.Strings(b.ids)
		out = append(out, ActionWave{
			ID: b.id, Title: b.title, RankMin: b.min, RankMax: b.max, ActionIDs: b.ids,
		})
	}
	return out
}

// epochsFromActionGraph builds legacy JobEpoch columns from the DAG for
// apply-bridge / older UI paths. Omits peel banner steps.
func epochsFromActionGraph(g *ActionGraph) []JobEpoch {
	if g == nil || len(g.Actions) == 0 {
		return nil
	}
	byID := make(map[string]*Action, len(g.Actions))
	for _, a := range g.Actions {
		byID[a.ID] = a
	}
	stepOf := func(a *Action) JobStep {
		st := JobStep{Kind: string(a.Mode), Target: a.Subject.Display, Detail: a.Detail, Why: a.Why}
		if st.Why == "" {
			st.Why = string(a.Reason)
		}
		return st
	}
	var out []JobEpoch
	for _, w := range g.Waves {
		steps := make([]JobStep, 0, len(w.ActionIDs))
		// Stable: sort by rank then id
		ids := append([]string(nil), w.ActionIDs...)
		sort.Slice(ids, func(i, j int) bool {
			ai, aj := byID[ids[i]], byID[ids[j]]
			if ai == nil || aj == nil {
				return ids[i] < ids[j]
			}
			if ai.Rank != aj.Rank {
				return ai.Rank < aj.Rank
			}
			if ai.Lane != aj.Lane {
				return ai.Lane < aj.Lane
			}
			return ai.ID < aj.ID
		})
		for _, id := range ids {
			if a := byID[id]; a != nil {
				steps = append(steps, stepOf(a))
			}
		}
		title := w.Title
		switch w.ID {
		case "land":
			title = "land"
		case "release":
			title = "release"
		}
		out = append(out, JobEpoch{ID: w.ID, Title: title, Steps: steps})
	}
	return out
}

// FormatActionGraphDebug returns a compact human debug listing.
func FormatActionGraphDebug(g *ActionGraph) string {
	if g == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "action graph: %d actions", len(g.Actions))
	if g.Excluded > 0 {
		fmt.Fprintf(&b, " (excluded %d)", g.Excluded)
	}
	b.WriteByte('\n')
	acts := append([]*Action(nil), g.Actions...)
	sort.Slice(acts, func(i, j int) bool {
		if acts[i].Rank != acts[j].Rank {
			return acts[i].Rank < acts[j].Rank
		}
		return acts[i].ID < acts[j].ID
	})
	for _, a := range acts {
		fmt.Fprintf(&b, "  [r%d] %s %s", a.Rank, a.Mode, a.Subject.Display)
		if a.Detail != "" {
			fmt.Fprintf(&b, " %s", a.Detail)
		}
		fmt.Fprintf(&b, " reason=%s", a.Reason)
		if len(a.Deps) > 0 {
			fmt.Fprintf(&b, " deps=%d", len(a.Deps))
		}
		b.WriteByte('\n')
	}
	if g.FilterNote != "" {
		fmt.Fprintf(&b, "note: %s\n", g.FilterNote)
	}
	return b.String()
}
