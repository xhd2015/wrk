package unwind

import (
	"fmt"
	"path/filepath"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
)

// scanModulesHook, when set (tests), replaces scan.Scan so call counts are
// observable. Production leaves this nil.
var scanModulesHook func(checkout string) ([]scan.Module, error)

func scanModules(checkout string) ([]scan.Module, error) {
	if scanModulesHook != nil {
		return scanModulesHook(checkout)
	}
	return scan.Scan(checkout, scan.Options{})
}

func scanModulesCached(checkout string, scans map[string][]scan.Module) ([]scan.Module, error) {
	if scans != nil {
		if mods, ok := scans[checkout]; ok {
			return mods, nil
		}
	}
	mods, err := scanModules(checkout)
	if err != nil {
		return nil, err
	}
	if scans != nil {
		scans[checkout] = mods
	}
	return mods, nil
}

// Snapshot is one inventory + DAG + optional cascade plan, with module scans
// reused (no second scan.Scan per checkout).
type Snapshot struct {
	WorkDir     string
	Inv         StackInventory
	RepoEdges   []RepoEdge
	Peel        *UnwindPlan
	ModuleNodes []UnwindGraphModuleNode
	ModuleEdges []UnwindGraphModuleEdge
	Cascade     *UnwindCascadePlan
	TagCache    tagScopePlanCache
}

// SnapshotOpts selects extra planning work on CollectSnapshot.
type SnapshotOpts struct {
	// Cascade fills module graph, tagscope, and cascade plan (show-graph,
	// verify, --tag-next dry-run, web). Plain peel dry-run leaves this false.
	Cascade bool
}

// CollectSnapshot gathers stack inventory, repo DAG, peel plan, and optionally
// one module-graph + tagscope + cascade pass.
func CollectSnapshot(workDir string, opts SnapshotOpts) (*Snapshot, error) {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve cwd: %w", err)
	}
	inv, err := CollectStackInventory(cwd)
	if err != nil {
		return nil, err
	}
	edges, err := buildRepoDAG(inv.Members, inv.scans)
	if err != nil {
		return nil, err
	}
	edges = mergeRepoEdges(edges, inv.SyntheticEdges)
	plan, err := PlanUnwind(inv.Members, edges)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		plan = &UnwindPlan{}
	}
	snap := &Snapshot{
		WorkDir:   cwd,
		Inv:       inv,
		RepoEdges: edges,
		Peel:      plan,
		TagCache:  make(tagScopePlanCache),
	}
	if opts.Cascade {
		cascade, nodes, modEdges := planCascadeOnceScans(inv.Members, inv.scans, snap.TagCache)
		snap.Cascade = cascade
		snap.ModuleNodes = nodes
		snap.ModuleEdges = modEdges
	}
	return snap, nil
}

func planCascadeOnce(members []StackMember) (*UnwindCascadePlan, []UnwindGraphModuleNode, []UnwindGraphModuleEdge) {
	return planCascadeOnceScans(members, nil, nil)
}

func planCascadeOnceCached(members []StackMember, tagCache tagScopePlanCache) (*UnwindCascadePlan, []UnwindGraphModuleNode, []UnwindGraphModuleEdge) {
	return planCascadeOnceScans(members, nil, tagCache)
}

func planCascadeOnceScans(members []StackMember, scans map[string][]scan.Module, tagCache tagScopePlanCache) (*UnwindCascadePlan, []UnwindGraphModuleNode, []UnwindGraphModuleEdge) {
	empty := &UnwindCascadePlan{}
	if len(members) == 0 {
		return empty, nil, nil
	}
	byLabel := pickPeelMembersByLabel(members)
	nodes, edges, err := buildUnwindModuleGraphScans(members, byLabel, scans)
	if err != nil {
		return empty, nil, nil
	}
	if tagCache == nil {
		tagCache = make(tagScopePlanCache)
	}
	attachTagScopeToModules(nodes, members, tagCache)
	cascade, err := planUnwindCascadeFromGraph(nodes, edges)
	if err != nil || cascade == nil {
		return empty, nodes, edges
	}
	return cascade, nodes, edges
}
