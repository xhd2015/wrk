package unwind

import "sort"

// JobPhase is one unwind preview/apply stage (snapshot, repos, modules, ship).
type JobPhase struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	// ActionGraph is set for phase 1 (cross-repo / repo land).
	ActionGraph *ActionGraph `json:"action_graph,omitempty"`
	// ByRepo holds per-repo graphs for phase 2 (intra pins) and ship (push+sync).
	ByRepo map[string]*ActionGraph `json:"by_repo,omitempty"`
}

// pinReposClassifies whether a pin's consumer and dep share a RepoLabel.
func pinIsIntraRepo(nodeByPath map[string]UnwindGraphModuleNode, consPath, depPath string) bool {
	c, okC := nodeByPath[consPath]
	d, okD := nodeByPath[depPath]
	if !okC || !okD {
		return false
	}
	if c.RepoLabel == "" || d.RepoLabel == "" {
		return false
	}
	return c.RepoLabel == d.RepoLabel
}

func moduleNodeIndex(nodes []UnwindGraphModuleNode) map[string]UnwindGraphModuleNode {
	out := make(map[string]UnwindGraphModuleNode, len(nodes))
	for _, n := range nodes {
		if n.Path != "" {
			out[n.Path] = n
		}
	}
	return out
}

const phase2EmptyNote = "no nested module requires a Phase 1 tag"

const (
	phaseTitleSnapshot = "snapshot"
	phaseTitleRepos    = "phase-1 · cross-repo unwind"
	phaseTitleModules  = "phase-2 · intra-repo update"
	phaseTitleShip     = "ship"
)

func emptyJobPhases(full *ActionGraph) []JobPhase {
	return []JobPhase{
		{ID: "snapshot", Title: phaseTitleSnapshot},
		{ID: "repos", Title: phaseTitleRepos, ActionGraph: full},
		{ID: "modules", Title: phaseTitleModules, ByRepo: map[string]*ActionGraph{}},
		{ID: "ship", Title: phaseTitleShip, ByRepo: map[string]*ActionGraph{}},
	}
}

func phaseByID(phases []JobPhase, id string) *JobPhase {
	for i := range phases {
		if phases[i].ID == id {
			return &phases[i]
		}
	}
	return nil
}

// buildJobPhases splits the full action graph into four stages matching
// CLI --dry-run / web preview:
//
//	snapshot — counts + module graph (no DAG)
//	phase-1  — repo land, cross-repo pins, tags, reinstall-local
//	phase-2  — intra-repo dep-update onto Phase 1 tags (parallel per repo)
//	ship     — push then sync (parallel per repo)
func buildJobPhases(snap *Snapshot, flags UnwindFlags, opts ActionGraphOpts) []JobPhase {
	full := BuildActionGraph(snap, flags, opts)
	full = expandLandGraph(full, flags)
	if full == nil || len(full.Actions) == 0 {
		return emptyJobPhases(full)
	}
	nodes := moduleNodeIndex(nil)
	if snap != nil {
		nodes = moduleNodeIndex(snap.ModuleNodes)
	}

	// Phase 2 is "pin the new parent tag into a nested module". Intra
	// pin-before-tag stays in Phase 1 so it sits on the land chain before
	// tag-next(consumer).
	intraPin := map[string]bool{}
	for _, a := range full.Actions {
		if a == nil || !pinLike(a.Mode) {
			continue
		}
		if a.Reason != ReasonPropagate {
			continue
		}
		if pinIsIntraRepo(nodes, a.Subject.Module, a.DepModule) {
			intraPin[a.ID] = true
		}
	}

	shipIDs := map[string]bool{}
	for _, a := range full.Actions {
		if a != nil && isShipMode(a.Mode) {
			shipIDs[a.ID] = true
		}
	}
	drop := map[string]bool{}
	for id := range intraPin {
		drop[id] = true
	}
	for id := range shipIDs {
		drop[id] = true
	}
	p1 := filterActionGraph(full, func(a *Action) bool {
		return a != nil && !intraPin[a.ID] && !isShipMode(a.Mode)
	}, drop)
	restretchGraph(p1)

	byRepo := map[string]*ActionGraph{}
	for _, a := range full.Actions {
		if a == nil || !intraPin[a.ID] {
			continue
		}
		repo := a.Lane
		if repo == "" {
			repo = a.Subject.RepoLabel
		}
		if repo == "" {
			repo = "_"
		}
		g := byRepo[repo]
		if g == nil {
			g = &ActionGraph{
				WorkDir:   full.WorkDir,
				Cleanup:   full.Cleanup,
				Artifacts: map[string]*Artifact{},
			}
			byRepo[repo] = g
		}
		cp := cloneAction(a)
		cp.Mode = ModeDepUpdate
		// Drop deps on phase-1-only nodes so the small graph is self-contained
		// for layout; UI still shows pin detail (version / dep module).
		var deps []string
		for _, d := range cp.Deps {
			if intraPin[d] {
				deps = append(deps, d)
			}
		}
		cp.Deps = deps
		g.Actions = append(g.Actions, cp)
	}
	// Every Phase 1 repo stays visible, even with nothing to pin (e.g. spl).
	if p1 != nil {
		for _, a := range p1.Actions {
			if a == nil || a.Lane == "" || a.Lane == "_" {
				continue
			}
			if byRepo[a.Lane] != nil {
				continue
			}
			byRepo[a.Lane] = &ActionGraph{
				WorkDir:    p1.WorkDir,
				Cleanup:    p1.Cleanup,
				Artifacts:  map[string]*Artifact{},
				FilterNote: phase2EmptyNote,
			}
		}
	}
	repos := make([]string, 0, len(byRepo))
	for r := range byRepo {
		repos = append(repos, r)
	}
	sort.Strings(repos)
	for _, r := range repos {
		g := byRepo[r]
		// Module swimlanes: use module path as lane for intra pins.
		levels := map[string]int{}
		for _, a := range g.Actions {
			mod := a.Subject.Module
			if mod == "" {
				mod = a.Lane
			}
			a.Lane = mod
			levels[mod] = 0
		}
		g.LaneLevels = levels
		appendPhase2Commit(g, r, nodes)
		for _, a := range g.Actions {
			if a == nil || a.ID != "commit:"+r || a.Lane == "" {
				continue
			}
			if _, ok := levels[a.Lane]; !ok {
				levels[a.Lane] = 0
				g.LaneLevels = levels
			}
			break
		}
		assignRanksByBand(g.Actions, levels)
		stretchRanksFromBand(g.Actions)
		g.Waves = deriveWaves(g.Actions)
	}

	if byRepo == nil {
		byRepo = map[string]*ActionGraph{}
	}
	return []JobPhase{
		{ID: "snapshot", Title: phaseTitleSnapshot},
		{ID: "repos", Title: phaseTitleRepos, ActionGraph: p1},
		{ID: "modules", Title: phaseTitleModules, ByRepo: byRepo},
		{ID: "ship", Title: phaseTitleShip, ByRepo: splitShipByRepo(full)},
	}
}

func splitShipByRepo(full *ActionGraph) map[string]*ActionGraph {
	out := map[string]*ActionGraph{}
	if full == nil {
		return out
	}
	shipIDs := map[string]bool{}
	for _, a := range full.Actions {
		if a != nil && isShipMode(a.Mode) {
			shipIDs[a.ID] = true
		}
	}
	for _, a := range full.Actions {
		if a == nil || !isShipMode(a.Mode) {
			continue
		}
		repo := a.Lane
		if repo == "" {
			repo = a.Subject.RepoLabel
		}
		if repo == "" {
			repo = "_"
		}
		g := out[repo]
		if g == nil {
			g = &ActionGraph{
				WorkDir:   full.WorkDir,
				Cleanup:   full.Cleanup,
				Artifacts: map[string]*Artifact{},
			}
			out[repo] = g
		}
		cp := cloneAction(a)
		var deps []string
		for _, d := range cp.Deps {
			if shipIDs[d] {
				deps = append(deps, d)
			}
		}
		cp.Deps = deps
		g.Actions = append(g.Actions, cp)
	}
	for _, g := range out {
		levels := map[string]int{}
		for _, a := range g.Actions {
			if a != nil && a.Lane != "" {
				levels[a.Lane] = 0
			}
		}
		g.LaneLevels = levels
		assignRanksByBand(g.Actions, levels)
		stretchRanksFromBand(g.Actions)
		g.Waves = deriveWaves(g.Actions)
	}
	return out
}

func appendPhase2Commit(g *ActionGraph, repo string, nodes map[string]UnwindGraphModuleNode) {
	if g == nil {
		return
	}
	var pins []*Action
	for _, a := range g.Actions {
		if a != nil && pinLike(a.Mode) {
			pins = append(pins, a)
		}
	}
	if len(pins) == 0 {
		return
	}
	bumps := make([]DepUpdateBump, 0, len(pins))
	var deps []string
	lane := pins[0].Lane
	for _, a := range pins {
		deps = append(deps, a.ID)
		from, to := "", a.PinVersion
		if n, ok := nodes[a.DepModule]; ok {
			from = n.LatestTag
			if to == "" {
				to = n.NextTag
			}
		}
		bumps = append(bumps, DepUpdateBump{Module: a.DepModule, From: from, To: to})
		if lane == "" {
			lane = a.Lane
		}
	}
	msg := FormatDepUpdateCommitMsg(bumps)
	if msg == "" {
		return
	}
	g.Actions = append(g.Actions, &Action{
		ID:      "commit:" + repo,
		Mode:    ModeCommit,
		Subject: Subject{Kind: SubRepo, ID: "repo:" + repo, Display: repo, RepoLabel: repo},
		Reason:  ReasonPropagate,
		Detail:  msg,
		Why:     "dep-update",
		Deps:    deps,
		Lane:    lane,
	})
}

// expandLandGraph splits bundled gen-commit into add-all / gen-commit-msg / commit
// so JobPlan and apply match the web DAG.
func expandLandGraph(g *ActionGraph, flags UnwindFlags) *ActionGraph {
	if g == nil || len(g.Actions) == 0 {
		return g
	}
	wantCommit := genArgsHasFlag(flags.GenCommitArgs, "--commit")
	wantAddAll := flags.AddAll || genArgsHasFlag(flags.GenCommitArgs, "--add-all")
	expanded := ExpandLandActions(g.Actions, wantAddAll, wantCommit)
	if len(expanded) == len(g.Actions) {
		same := true
		for i := range expanded {
			if expanded[i] == nil || g.Actions[i] == nil || expanded[i].ID != g.Actions[i].ID || expanded[i].Mode != g.Actions[i].Mode {
				same = false
				break
			}
		}
		if same {
			return g
		}
	}
	out := &ActionGraph{
		WorkDir:    g.WorkDir,
		Cleanup:    g.Cleanup,
		Artifacts:  map[string]*Artifact{},
		LaneLevels: g.LaneLevels,
		Excluded:   g.Excluded,
		FilterNote: g.FilterNote,
	}
	for id, art := range g.Artifacts {
		out.Artifacts[id] = art
	}
	for _, a := range expanded {
		if a == nil || a.Mode != ModeGenCommitMsg {
			continue
		}
		artID := landMessageArtifactID(a.Lane)
		if out.Artifacts[artID] == nil {
			out.Artifacts[artID] = &Artifact{ID: artID, Kind: ArtMessage, Repo: a.Lane}
		}
	}
	out.Actions = expanded
	assignRanksByBand(out.Actions, out.LaneLevels)
	stretchRanksFromBand(out.Actions)
	out.Waves = deriveWaves(out.Actions)
	return out
}

func cloneAction(a *Action) *Action {
	if a == nil {
		return nil
	}
	cp := *a
	if a.Deps != nil {
		cp.Deps = append([]string(nil), a.Deps...)
	}
	if a.Consumes != nil {
		cp.Consumes = append([]string(nil), a.Consumes...)
	}
	if a.Produces != nil {
		cp.Produces = append([]string(nil), a.Produces...)
	}
	return &cp
}

func filterActionGraph(src *ActionGraph, keep func(*Action) bool, dropIDs map[string]bool) *ActionGraph {
	out := &ActionGraph{
		WorkDir:    src.WorkDir,
		Cleanup:    src.Cleanup,
		Artifacts:  map[string]*Artifact{},
		LaneLevels: src.LaneLevels,
		Excluded:   src.Excluded,
		FilterNote: src.FilterNote,
	}
	for id, art := range src.Artifacts {
		out.Artifacts[id] = art
	}
	kept := map[string]bool{}
	for _, a := range src.Actions {
		if !keep(a) {
			continue
		}
		cp := cloneAction(a)
		var deps []string
		for _, d := range cp.Deps {
			if dropIDs[d] {
				continue
			}
			deps = append(deps, d)
		}
		cp.Deps = deps
		out.Actions = append(out.Actions, cp)
		kept[cp.ID] = true
	}
	// Second pass: drop deps that were not kept (e.g. removed land).
	for _, a := range out.Actions {
		var deps []string
		for _, d := range a.Deps {
			if kept[d] {
				deps = append(deps, d)
			}
		}
		a.Deps = deps
	}
	return out
}

func restretchGraph(g *ActionGraph) {
	if g == nil {
		return
	}
	assignRanksByBand(g.Actions, g.LaneLevels)
	stretchRanksFromBand(g.Actions)
	g.Waves = deriveWaves(g.Actions)
}
