# wrk — PlanLocalReinstallsMulti (P1 multi-module plan API)

## Version
0.0.2

Decision tree for **Phase 1 multi-module pure API**: given an ordered list of
module roots and a shared `binDir`, build a multi-module reinstall plan by
running per-module discovery (existing single-module rules) and applying a
**cross-module hard error** when the same bin has `Action=install` from two
modules.

**Nested root** under `reinstall-local/`: self-contained so sealed single-module
leaves under `plan/` and `error/` keep their own `Run` contract and stay GREEN.

**Classic TDD**: `wrkcli.PlanLocalReinstallsMulti` is not implemented yet —
expect **RED** (compile failure or assert failure) until the implementer lands
types + function. Do **not** implement production code in this design pass.

Out of scope here: git scan-root / mod discovery (P2), CLI dry-run UX (P3),
`--main` (P4), execute (P5).

# DSN (Domain Specific Notion)

- **PlanLocalReinstallsMulti** — pure function in package `wrkcli`:
  `PlanLocalReinstallsMulti(moduleRoots []string, binDir string) (*MultiLocalReinstallPlan, error)`.
  No subprocesses, no network, no writes outside reading each module tree and
  stating entries under `binDir`. Callers supply absolute (or resolvable)
  module roots; this API does not walk git or go.work to find modules.
- **moduleRoots** — list of directories that each contain a parseable `go.mod`.
  Empty list → empty multi plan, nil error (no modules, no items).
  Invalid module (missing go.mod) on any root → non-nil error (same class as
  single-module planner failure for that root).
- **Per-module discovery** — for each root, reuse the same rules as
  `PlanLocalReinstalls`: `./cmd/...` package main → `go-install`;
  `./script/.../install` package main → `go-run-install` (script wins same
  BinName within a module); filter against shared `binDir` → `install` or
  `skip`; items sorted by BinName within the module.
- **MultiLocalReinstallPlan** — `{ BinDir, Modules []ModuleReinstallPlan }`.
  **Modules** ordered lexicographically by absolute `ModuleRoot` path (not
  caller list order). Each **ModuleReinstallPlan** exposes at least
  `ModuleRoot` (abs), `ModuleName` (basename of module path), and `Items`
  (`PlanItem` rows with BinName, Method, RelPath, Action — same shape as
  single-module). Optional fields (`ModulePath`, `RelDir`) may appear in the
  production type; asserts lock ModuleRoot / ModuleName / Items / BinDir.
- **Cross-module collision** — after planning every module, if the same
  `BinName` appears with `Action=install` in **two or more** modules →
  **hard error** (no multi plan). Error text must name the bin and identify
  both claiming modules (paths and/or module names). **Skip-only** duplicates
  do **not** collide: install×skip for the same bin across modules is allowed
  (prefer no error; both rows may remain on their modules).
- **Single-module equivalence** — a one-element `moduleRoots` list produces
  the same item set (BinName, Method, RelPath, Action) as
  `PlanLocalReinstalls` for that root and binDir.
- **Non-goals** — CLI, dry-run text, execute/`go install`, module discovery
  beyond the explicit root list, changing sealed single-module trees.

## Tree Overview

```
multi/                                    # nested pure-API root
├── plan/                                 # multi plan succeeds
│   ├── empty/
│   │   └── no-roots/                     # M5: moduleRoots=[] → empty Modules
│   ├── single/
│   │   └── matches-single-api/           # M1: one root ≡ PlanLocalReinstalls
│   ├── distinct/
│   │   └── root-and-tools/               # M2: two modules, distinct bins
│   └── skip-dup/
│       └── install-and-skip/             # M4: same bin skip×skip → no error
└── error/
    └── install-install/                  # M3: same bin install×install → error
```

Split factor (MECE, significance-first):

1. **Plan validity** — success vs hard collision error.
2. Within success: **module-list shape** — empty / single / multi-distinct /
   multi same-bin skip-only duplicate (no hard error).
3. Error branch: install×install claim only (skip dups are under plan/).

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| M1 | plan/single/matches-single-api | one module root; multi Items match `PlanLocalReinstalls` for same root+binDir |
| M2 | plan/distinct/root-and-tools | root module + nested `tools/` module; both contribute distinct install bins; modules lex by ModuleRoot |
| M3 | error/install-install | two modules both `./cmd/samebin` with bin present → error naming bin + both modules |
| M4 | plan/skip-dup/install-and-skip | same bin skip×skip (bin absent): nil error; both modules list skip (skip-only dups ok) |
| M5 | plan/empty/no-roots | `moduleRoots` empty → BinDir set, Modules empty, nil error |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/reinstall-local/multi
doctest test ./cmd/wrk/tests/reinstall-local/multi
doctest test ./cmd/wrk/tests/reinstall-local/multi/plan/single/matches-single-api
doctest test ./cmd/wrk/tests/reinstall-local/multi/plan/distinct/root-and-tools
doctest test ./cmd/wrk/tests/reinstall-local/multi/error/install-install
doctest test ./cmd/wrk/tests/reinstall-local/multi/plan/skip-dup/install-and-skip
doctest test ./cmd/wrk/tests/reinstall-local/multi/plan/empty/no-roots
```

Sealed single-module tree (must stay GREEN; do not rewrite its ASSERT leaves):

```sh
doctest test ./cmd/wrk/tests/reinstall-local/plan
doctest test ./cmd/wrk/tests/reinstall-local/error
```

Classic TDD: multi leaves **RED** until implementer lands
`PlanLocalReinstallsMulti` / `MultiLocalReinstallPlan` / `ModuleReinstallPlan`.
Compile failure or assert failure both count as RED.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/wrk/wrkcli"
)

// Request drives PlanLocalReinstallsMulti against per-leaf fixture modules + bin dir.
// Root Setup allocates WorkRoot and BinDir; leaves create module dirs under WorkRoot,
// fill ModuleRoots, and set Want* expectations.
type Request struct {
	WorkRoot string
	BinDir   string

	// ModuleRoots are absolute paths to go.mod dirs passed to the multi API.
	// May be empty (M5). Order here is caller order; product re-sorts lex by path.
	ModuleRoots []string

	// WantError: when true, Assert expects Run's error to be non-nil.
	WantError bool

	// WantErrSubstrs: optional substrings that must all appear in err.Error()
	// when WantError (e.g. bin name + both module identifiers). Empty when ok.
	WantErrSubstrs []string

	// WantModules is the full expected Modules list, already sorted by ModuleRoot.
	// Nil/empty means zero modules. Skip entries are included when expected.
	// Used when WantError is false.
	WantModules []WantModulePlan
}

// WantModulePlan is the assertable shape of one module block in the multi plan.
type WantModulePlan struct {
	ModuleRoot string // absolute
	ModuleName string // go.mod module path basename
	Items      []WantPlanItem
}

// WantPlanItem is the assertable shape of one plan row (same as single-module).
type WantPlanItem struct {
	BinName string
	Method  string // "go-install" | "go-run-install"
	RelPath string // "./cmd/..." or "./script/.../install"
	Action  string // "install" | "skip"
}

// Response mirrors MultiLocalReinstallPlan for asserts.
type Response struct {
	BinDir  string
	Modules []WantModulePlan
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	// Classic TDD: PlanLocalReinstallsMulti / MultiLocalReinstallPlan are the
	// production API under design. RED (compile or assert) until implementer lands them.
	plan, err := wrkcli.PlanLocalReinstallsMulti(req.ModuleRoots, req.BinDir)
	if err != nil {
		return &Response{}, err
	}
	if plan == nil {
		return &Response{}, nil
	}
	mods := make([]WantModulePlan, len(plan.Modules))
	for i, m := range plan.Modules {
		items := make([]WantPlanItem, len(m.Items))
		for j, it := range m.Items {
			items[j] = WantPlanItem{
				BinName: it.BinName,
				Method:  string(it.Method),
				RelPath: it.RelPath,
				Action:  string(it.Action),
			}
		}
		mods[i] = WantModulePlan{
			ModuleRoot: m.ModuleRoot,
			ModuleName: m.ModuleName,
			Items:      items,
		}
	}
	return &Response{
		BinDir:  plan.BinDir,
		Modules: mods,
	}, nil
}
```
