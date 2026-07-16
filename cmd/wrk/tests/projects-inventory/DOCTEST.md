# wrk projects inventory — pure graph helpers + source release resolution (P1)

## Version
0.0.2

Decision tree for **Phase 1** pure helpers that build an in-process model of
registered projects → modules → cross/intra require edges, and resolve source
module release tags for go `require` versions.

**Classic TDD (RED):** `wrkcli.BuildInventory`, inventory edge methods, and
`wrkcli.ResolveSourceReleases` do **not** exist yet. Leaves must fail compile or
assert until the implementer lands the public API below. **No CLI flags** in
this tree (`--projects-dep-graph`, `--propagate-tags` are P2+).

# DSN (Domain Specific Notion)

- **WRK_HOME registry** — `{WRK_HOME}/projects.json` lists main-repo absolute
  paths (same schema as `storage.ListProjects`). Paths are soft-skipped when
  missing on disk; present paths are scanned.
- **Inventory builder** — pure package API in `github.com/xhd2015/wrk/wrkcli`:
  `BuildInventory(wrkHome string) (Inventory, error)`. Loads registered
  projects, scans each with `gotool/mod/scan`-equivalent multi-module walk
  (root + nested `go.mod`), builds ownership map module-path → project path.
- **Project / Module** — each project has `Path` and `Modules[]` with `Dir`
  (relative to project root, `"."` for root), `Path` (module path), and
  `Requires[{Path, Version}]` (and optional local replaces for later phases).
- **Ownership** — `Inventory.FindOwner(modulePath) (projectPath string, ok bool)`.
- **Edges** — a require from consumer module A to dep module path D is an
  **Edge** `{ConsumerProject, ConsumerModule, DepPath, DepVersion, OwnerProject}`.
  `CrossEdges()` returns edges where owner project is known and **differs** from
  consumer project. `IntraEdges()` returns edges where owner project is known
  and **equals** consumer project (monorepo sibling requires). Unknown owners
  are neither cross nor intra for P1 assertions (or omitted from both).
- **Source release resolution** — `ResolveSourceReleases(sourceMain string)
  (SourceReleasesResult, error)` scans modules under a single source main repo
  and maps **latest numeric** git release tags to go require versions:

  | Module location | Git tag | Version for require |
  |-----------------|---------|---------------------|
  | root `go.mod`   | `v1.2.3` | `v1.2.3` |
  | nested `sub/`   | `sub/v0.1.0` | `v0.1.0` |

  Only **numeric** release tags (no prerelease suffix). Per-module missing
  tags: omit from `Releases`, collect module path in `Missing` — still return
  other modules. Prefer no hard error when at least one module resolves
  (zero modules / zero tags still returns empty lists, not fatal).
- **No CLI / no writes** — this tree never runs the wrk binary, never writes
  go.mod, never tags, never commits. Fixtures use real git + go.mod files under
  isolated temp dirs.

## Expected public API (implementer must match)

```go
package wrkcli

// BuildInventory loads WRK_HOME projects, soft-skips missing paths, scans
// modules, and builds ownership.
func BuildInventory(wrkHome string) (Inventory, error)

type Inventory struct {
	Projects     []ProjectEntry // included (existing) projects only
	SkippedPaths []string       // registry paths soft-skipped (missing disk)
}

type ProjectEntry struct {
	Path    string
	Modules []ModuleEntry
}

type ModuleEntry struct {
	Dir      string // relative to project Path; "." for root
	Path     string // module path from go.mod
	Requires []RequireEntry
}

type RequireEntry struct {
	Path    string
	Version string
}

// FindOwner returns the registered project path that owns modulePath.
func (inv Inventory) FindOwner(modulePath string) (projectPath string, ok bool)

// CrossEdges: require edges where consumer and owner projects both known and differ.
func (inv Inventory) CrossEdges() []Edge

// IntraEdges: require edges where consumer and owner projects both known and equal.
func (inv Inventory) IntraEdges() []Edge

type Edge struct {
	ConsumerProject string
	ConsumerModule  string // consumer module path
	DepPath         string
	DepVersion      string
	OwnerProject    string
}

// ResolveSourceReleases scans sourceMain for modules and maps numeric tags.
func ResolveSourceReleases(sourceMain string) (SourceReleasesResult, error)

type SourceReleasesResult struct {
	Releases []SourceRelease
	Missing  []string // module paths with no numeric release tag
}

type SourceRelease struct {
	ModulePath string
	Tag        string // full git tag, e.g. "v1.2.3" or "sub/v0.1.0"
	Version    string // go require version, e.g. "v1.2.3" or "v0.1.0"
}
```

Naming may live on `wrkcli` (preferred, matches `PartitionStatusPaths`) or a
small subpackage re-exported from `wrkcli`. Tests import `github.com/xhd2015/wrk/wrkcli`.

## Tree Overview

```
projects-inventory/
├── inventory/                         # Op = build inventory / edges
│   ├── empty-registry/                # empty projects.json → empty inventory
│   ├── multi-module-scan/             # two projects; root + nested sub modules
│   ├── cross-edge/                    # app requires lib → one cross edge
│   ├── intra-edge-not-cross/          # monorepo root requires sub → intra only
│   └── skip-missing-project-path/     # missing path soft-skip + good project
└── source-releases/                   # Op = ResolveSourceReleases
    ├── root-and-sub/                  # v1.2.3 + sub/v0.1.0 → both versions
    └── missing-tag/                   # root tagged, sub untagged → Missing
```

Split factor (MECE, significance-first):

1. **API under test** — inventory graph vs source release resolution.
2. Within inventory: registry emptiness / multi-module discovery / edge class /
   soft-skip (MECE scenarios for P1 exit criteria).
3. Within source-releases: all modules tagged vs partial missing.

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| I1 | inventory/empty-registry | empty registry → 0 projects, 0 modules, 0 cross/intra edges, 0 skipped |
| I2 | inventory/multi-module-scan | lib (root+sub) + app registered → modules with correct Dir/Path |
| I3 | inventory/cross-edge | app requires `example.com/lib@v1.0.0`; lib owned → one CrossEdge; Intra empty for that require |
| I4 | inventory/intra-edge-not-cross | monorepo root requires nested sub → IntraEdges contains it; CrossEdges empty |
| I5 | inventory/skip-missing-project-path | missing path + good project → SkippedPaths has missing; good project modules present |
| S1 | source-releases/root-and-sub | tags `v1.2.3` + `sub/v0.1.0` → Releases both with correct Version |
| S2 | source-releases/missing-tag | root `v1.0.0` only → root in Releases; sub module path in Missing |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/projects-inventory
doctest test -v ./cmd/wrk/tests/projects-inventory
doctest test ./cmd/wrk/tests/projects-inventory/inventory/empty-registry
doctest test ./cmd/wrk/tests/projects-inventory/inventory/cross-edge
doctest test ./cmd/wrk/tests/projects-inventory/source-releases/root-and-sub
```

Expect **RED** until implementer lands the public API (compile failure on
missing symbols is the Classic TDD red bar).

```go
import (
	"fmt"
	"sort"
	"testing"

	"github.com/xhd2015/wrk/wrkcli"
)

// Op selects which pure helper Run exercises.
const (
	OpInventory      = "inventory"
	OpSourceReleases = "source-releases"
)

// Request is filled by Setup chain (WorkRoot/WrkHome at root; Op + fixtures
// at grouping/leaf). Assert compares Response to Want* fields.
type Request struct {
	WorkRoot string
	WrkHome  string

	// Op is OpInventory or OpSourceReleases.
	Op string

	// SourceMain is the absolute main-repo path for OpSourceReleases.
	SourceMain string

	// Want* expected values (raw fixture paths/module paths; Assert normalizes
	// filesystem paths via storage.NormalizePath where applicable).
	WantProjectPaths []string
	WantModules      []WantModule
	WantCrossEdges   []WantEdge
	WantIntraEdges   []WantEdge
	WantSkippedPaths []string

	WantReleases []WantRelease
	WantMissing  []string
}

type WantModule struct {
	ProjectPath string // owning project absolute path
	Dir         string // "." or "sub" etc.
	Path        string // module path
}

type WantEdge struct {
	ConsumerProject string
	ConsumerModule  string
	DepPath         string
	// DepVersion optional; empty means do not assert version string.
	DepVersion   string
	OwnerProject string
}

type WantRelease struct {
	ModulePath string
	Tag        string
	Version    string
}

type Response struct {
	// Inventory side (OpInventory)
	ProjectPaths []string
	Modules      []ModuleSnap
	CrossEdges   []EdgeSnap
	IntraEdges   []EdgeSnap
	SkippedPaths []string

	// Source-release side (OpSourceReleases)
	Releases []ReleaseSnap
	Missing  []string
}

type ModuleSnap struct {
	ProjectPath string
	Dir         string
	Path        string
}

type EdgeSnap struct {
	ConsumerProject string
	ConsumerModule  string
	DepPath         string
	DepVersion      string
	OwnerProject    string
}

type ReleaseSnap struct {
	ModulePath string
	Tag        string
	Version    string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Op {
	case OpInventory:
		// Classic TDD: BuildInventory + CrossEdges/IntraEdges under design.
		inv, err := wrkcli.BuildInventory(req.WrkHome)
		if err != nil {
			return nil, err
		}
		resp := &Response{
			ProjectPaths: make([]string, 0, len(inv.Projects)),
			SkippedPaths: append([]string(nil), inv.SkippedPaths...),
		}
		for _, p := range inv.Projects {
			resp.ProjectPaths = append(resp.ProjectPaths, p.Path)
			for _, m := range p.Modules {
				resp.Modules = append(resp.Modules, ModuleSnap{
					ProjectPath: p.Path,
					Dir:         m.Dir,
					Path:        m.Path,
				})
			}
		}
		for _, e := range inv.CrossEdges() {
			resp.CrossEdges = append(resp.CrossEdges, EdgeSnap{
				ConsumerProject: e.ConsumerProject,
				ConsumerModule:  e.ConsumerModule,
				DepPath:         e.DepPath,
				DepVersion:      e.DepVersion,
				OwnerProject:    e.OwnerProject,
			})
		}
		for _, e := range inv.IntraEdges() {
			resp.IntraEdges = append(resp.IntraEdges, EdgeSnap{
				ConsumerProject: e.ConsumerProject,
				ConsumerModule:  e.ConsumerModule,
				DepPath:         e.DepPath,
				DepVersion:      e.DepVersion,
				OwnerProject:    e.OwnerProject,
			})
		}
		sort.Strings(resp.ProjectPaths)
		sort.Strings(resp.SkippedPaths)
		return resp, nil

	case OpSourceReleases:
		// Classic TDD: ResolveSourceReleases under design.
		if req.SourceMain == "" {
			return nil, fmt.Errorf("SourceMain required for OpSourceReleases")
		}
		result, err := wrkcli.ResolveSourceReleases(req.SourceMain)
		if err != nil {
			return nil, err
		}
		resp := &Response{
			Missing: append([]string(nil), result.Missing...),
		}
		for _, r := range result.Releases {
			resp.Releases = append(resp.Releases, ReleaseSnap{
				ModulePath: r.ModulePath,
				Tag:        r.Tag,
				Version:    r.Version,
			})
		}
		sort.Strings(resp.Missing)
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown Op %q", req.Op)
	}
}
```
