# wrk — reinstall scan-root resolution + module discovery (P2)

## Version
0.0.2

Decision tree for **Phase 2 scan-root resolution**: from a `workDir` and
`useMain` flag, resolve the directory that should be scanned for Go modules,
discover every module under that root via `gotool/mod/scan.Scan`, then feed the
absolute module roots into `PlanLocalReinstallsMulti`.

**Nested root** under `reinstall-local/`: self-contained so sealed
single-module leaves (`plan/`, `error/`) and sealed multi-module plan leaves
(`multi/`) keep their own `Run` contracts and stay GREEN/RED independently.

**Classic TDD**: `ResolveReinstallScanRoot` and
`PlanLocalReinstallsFromWorkDir` are not wired yet — expect **RED** (compile
failure or assert failure) until the implementer lands them. Do **not**
implement production code in this design pass. Do **not** rewrite sealed
`multi/` plan/error ASSERT leaves.

Out of scope here: CLI dry-run formatting (P3), `--main` flag compose in
`run.go` (P4), execute (P5). Explicit `moduleRoots` list API remains
`PlanLocalReinstallsMulti` under `multi/`.

# DSN (Domain Specific Notion)

- **ResolveReinstallScanRoot** — pure-ish function in package `wrkcli`:
  `ResolveReinstallScanRoot(workDir string, useMain bool) (string, error)`.
  Returns the absolute **scan root** directory from which module discovery
  begins. Rules (in priority order for callers):
  1. **`useMain == true`**: workDir is inside a git checkout → scan root =
     main repository path (`worktree.ResolveMainRepo` of
     `worktree.ShowToplevel(workDir)`). Hard error if not in git / main cannot
     be resolved.
  2. **`useMain == false` and in git**: scan root =
     `worktree.ShowToplevel(workDir)` (checkout toplevel for this work tree,
     including linked worktrees).
  3. **Not in git**: walk up from `workDir` looking for a `go.mod` (same class
     of walk as today's single-module reinstall path). First `go.mod` directory
     is the scan root.
  4. **No scan root**: not in git and no `go.mod` on the walk-up → non-nil
     error (message should mention `go.mod` or the failed walk).
- **PlanLocalReinstallsFromWorkDir** — orchestration entrypoint:
  `PlanLocalReinstallsFromWorkDir(workDir, binDir string, useMain bool) (*MultiLocalReinstallPlan, error)`.
  Resolves scan root, runs `mod/scan.Scan(scanRoot, …)` to collect every Go
  module under that root (root `go.mod` plus nested modules such as `tools/`),
  maps each scanned module to an absolute module root directory, then calls
  `PlanLocalReinstallsMulti(moduleRoots, binDir)`. Cross-module install×install
  collision rules are unchanged (owned by multi API).
- **Scan root vs workDir** — when workDir is a **subdirectory** of a git
  checkout (no own go.mod required at workDir), git modes still scan the full
  checkout toplevel (or main when `useMain`), so **all** nested modules under
  that root are included — not only a walk-up single module.
- **mod/scan** — `github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan`: finds
  modules under a root with vendor/testdata/gitignore/nested-repo skips.
  `Module.Dir` is relative to the scan root (`"."` or `"tools"`). Absolute
  module root = `scanRoot` when Dir is `"."`, else `filepath.Join(scanRoot, Dir)`.
- **MultiLocalReinstallPlan** — same shape as multi nested root: `{ BinDir,
  Modules []ModuleReinstallPlan }` sorted by absolute ModuleRoot; items by
  BinName. Per-module discovery reuses single-module cmd/script rules.
- **useMain / linked worktree** — with `useMain=true` from a linked worktree
  path, scan root identity is the **main repo path**, not the linked worktree
  path (even if content is similar). With `useMain=false`, linked worktree
  cwd uses that worktree's ShowToplevel.
- **Non-goals** — CLI flags, dry-run text, execute/`go install`, changing
  sealed `PlanLocalReinstalls` / `PlanLocalReinstallsMulti` leaf contracts.

## Tree Overview

```
scan/                                       # nested root: workDir+useMain → scan → multi plan
├── git/                                    # useMain=false, workDir inside git
│   └── multi-from-subdir/                  # R1: root+tools; workDir=subdir → both modules
├── use-main/                               # useMain=true
│   └── linked-worktree/                    # R3: scan root = main repo path
├── non-git/                                # not a git work tree
│   └── walk-up-go-mod/                     # R2: walk-up finds go.mod → single module
└── error/
    └── no-go-mod/                          # R4: no git, no go.mod on walk-up → error
```

Split factor (MECE, significance-first):

1. **Scan-root strategy / environment** — git toplevel (`useMain=false`) vs
   main-repo (`useMain=true`) vs non-git walk-up vs unresolvable error.
2. Within each success strategy: one concrete fixture that locks the P2 exit
   criterion for that branch (multi-module from subdir; main path from linked
   wt; single module via walk-up).

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| R1 | git/multi-from-subdir | git repo with root go.mod + nested tools/; workDir=pkg/sub → ScanRoot=toplevel; multi plan installs rootbin + toolbin |
| R2 | non-git/walk-up-go-mod | non-git tree; workDir under module → ScanRoot=module; single module plan |
| R3 | use-main/linked-worktree | linked worktree cwd + useMain → ScanRoot=main abs path (not linked path); plan from main modules |
| R4 | error/no-go-mod | non-git empty dir, no go.mod → Resolve/Plan error |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/reinstall-local/scan
doctest test ./cmd/wrk/tests/reinstall-local/scan
doctest test ./cmd/wrk/tests/reinstall-local/scan/git/multi-from-subdir
doctest test ./cmd/wrk/tests/reinstall-local/scan/use-main/linked-worktree
doctest test ./cmd/wrk/tests/reinstall-local/scan/non-git/walk-up-go-mod
doctest test ./cmd/wrk/tests/reinstall-local/scan/error/no-go-mod
```

Sealed trees (do not rewrite ASSERT leaves):

```sh
doctest test ./cmd/wrk/tests/reinstall-local/plan
doctest test ./cmd/wrk/tests/reinstall-local/error
doctest test ./cmd/wrk/tests/reinstall-local/multi
```

Classic TDD: scan leaves **RED** until implementer lands
`ResolveReinstallScanRoot` / `PlanLocalReinstallsFromWorkDir` (and reuses
`PlanLocalReinstallsMulti` + `mod/scan`). Compile failure or assert failure
both count as RED.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/wrk/wrkcli"
)

// Request drives scan-root resolution + PlanLocalReinstallsFromWorkDir.
// Root Setup allocates WorkRoot and BinDir; leaves build fixtures, set
// WorkDir / UseMain, and Want* expectations.
type Request struct {
	WorkRoot string
	// WorkDir is the directory passed to Resolve / FromWorkDir (absolute).
	WorkDir string
	BinDir  string
	UseMain bool

	// WantError: when true, Assert expects Run's error to be non-nil.
	WantError bool

	// WantErrSubstrs: optional substrings that must all appear in err.Error()
	// when WantError. Empty when ok.
	WantErrSubstrs []string

	// WantScanRoot is the expected absolute scan root from
	// ResolveReinstallScanRoot. Empty when WantError.
	WantScanRoot string

	// WantModules is the full expected Modules list from
	// PlanLocalReinstallsFromWorkDir, sorted by ModuleRoot. Used when
	// WantError is false.
	WantModules []WantModulePlan
}

// WantModulePlan is the assertable shape of one module block in the multi plan.
type WantModulePlan struct {
	ModuleRoot string // absolute
	ModuleName string // go.mod module path basename
	Items      []WantPlanItem
}

// WantPlanItem is the assertable shape of one plan row.
type WantPlanItem struct {
	BinName string
	Method  string // "go-install" | "go-run-install"
	RelPath string // "./cmd/..." or "./script/.../install"
	Action  string // "install" | "skip"
}

// Response mirrors resolver + multi plan for asserts.
type Response struct {
	ScanRoot string
	BinDir   string
	Modules  []WantModulePlan
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	_ = d
	// Classic TDD: ResolveReinstallScanRoot + PlanLocalReinstallsFromWorkDir
	// are the production APIs under design. RED until implementer lands them.
	scanRoot, err := wrkcli.ResolveReinstallScanRoot(req.WorkDir, req.UseMain)
	if err != nil {
		return &Response{}, err
	}
	plan, err := wrkcli.PlanLocalReinstallsFromWorkDir(req.WorkDir, req.BinDir, req.UseMain)
	if err != nil {
		return &Response{ScanRoot: scanRoot}, err
	}
	if plan == nil {
		return &Response{ScanRoot: scanRoot}, nil
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
		ScanRoot: scanRoot,
		BinDir:   plan.BinDir,
		Modules:  mods,
	}, nil
}
```
