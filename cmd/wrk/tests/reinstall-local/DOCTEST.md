# wrk — PlanLocalReinstalls discovery & filter (P1)

## Version
0.0.2

Decision tree for **Phase 1** pure API that plans local binary reinstalls for a
single Go module: discover `package main` under `./cmd/...` and
`./script/.../install`, filter against an explicit bin directory, and return a
deterministic sorted plan. **No CLI flags, no dry-run output, no `go install` /
`go run` execution** in this tree.

Later phases (out of scope for this single-module tree): `wrk --reinstall-local`
wiring, GOBIN / `GOPATH/bin` resolution, real install, `--force`,
continue-on-error. **Multi-module plan API** lives in the nested root
`multi/` (`PlanLocalReinstallsMulti`). **Scan-root resolution + module
discovery** lives in nested root `scan/`
(`ResolveReinstallScanRoot` / `PlanLocalReinstallsFromWorkDir`) so sealed
single-module and multi plan leaves stay untouched.

# DSN (Domain Specific Notion)

- **PlanLocalReinstalls** — pure function in package `wrkcli`:
  `PlanLocalReinstalls(moduleRoot, binDir string) (*LocalReinstallPlan, error)`.
  No subprocesses, no network, no writes outside reading the module tree and
  stating entries under `binDir`. Callers later resolve `binDir` from GOBIN or
  `$(go env GOPATH)/bin`; P1 takes the path as an explicit argument.
- **moduleRoot** — directory that must contain a parseable `go.mod`. Missing or
  unparseable module path → non-nil error (no plan).
- **ModuleName** — last path segment of the `module` line in `go.mod`
  (e.g. `example.com/demo` → `demo`; `disk-usage-analyser` →
  `disk-usage-analyser`). Exposed on the plan for bare-script naming and asserts.
- **cmd discovery** — walk `moduleRoot/cmd` recursively; keep directories that
  are Go `package main` (any non-test `*.go` with `package main`). Skip
  `testdata`, `vendor`, and hidden directory names (leading `.`). Bin name =
  last path segment. Method = `go-install`. Relative package path =
  `./cmd/...` from module root (slash-separated).
- **script discovery** — walk under `moduleRoot/script` for directories **named**
  `install` that are `package main`. Bin name = parent directory name.
  **Exception**: bare `./script/install` (parent is `script`) → bin =
  **ModuleName**. Method = `go-run-install`. Relative path =
  `./script/.../install`.
- **Dedup (script wins)** — when the same bin name is produced by both cmd and
  script discovery, keep **one** plan item: the script (`go-run-install`) wins.
- **Filter** — for each candidate, if `$binDir/<binName>` does **not** exist as a
  regular file (or a symlink that resolves to a file), set `Action` to `skip`;
  otherwise `install`. Skip entries remain in the plan (assertable shape).
- **Sort** — plan `Items` ordered lexicographically by `BinName`.
- **LocalReinstallPlan** — `{ ModuleRoot, ModuleName, BinDir, Items []PlanItem }`.
  Each `PlanItem` has `BinName`, `Method` (`go-install` | `go-run-install`),
  `RelPath` (`./cmd/...` or `./script/.../install`), `Action` (`install` |
  `skip`).
- **Non-goals (P1)** — CLI flags, dry-run text, real install/run, `--force`,
  multi-module / `go.work` walk, mutual exclusion with other wrk modes.

## Tree Overview

```
reinstall-local/
├── error/                              # plan fails (sealed single-module)
│   └── no-go-mod/                      # moduleRoot has no go.mod
├── plan/                               # plan succeeds (sealed single-module)
│   ├── cmd/                            # ./cmd/... only
│   │   ├── present-install/            # bin present → install
│   │   ├── absent-skip/                # bin missing → skip
│   │   └── non-main-ignored/           # non-main under cmd → not a candidate
│   ├── script/                         # ./script/.../install only
│   │   ├── nested-foo-present/         # ./script/foo/install → bin foo
│   │   └── bare-module-basename/       # ./script/install → ModuleName
│   ├── conflict/
│   │   └── script-wins/                # cmd+script same bin → script only
│   ├── nested-cmd/
│   │   └── present-install/            # ./cmd/nested/tool → bin tool
│   ├── mixed/
│   │   └── sorted-filter/              # multi bins; sort + install/skip mix
│   └── empty/
│       └── no-candidates/              # no package main → empty Items
├── multi/                              # nested DOCTEST root: PlanLocalReinstallsMulti
│   ├── plan/                           # multi plan ok (empty / single / distinct / skip-dup)
│   └── error/                          # install×install cross-module collision
└── scan/                               # nested DOCTEST root: scan-root + mod discovery (P2)
    ├── git/                            # ShowToplevel (useMain=false)
    ├── use-main/                       # main repo path (useMain=true)
    ├── non-git/                        # walk-up go.mod
    └── error/                          # no go.mod / unresolvable
```

Split factor (MECE, significance-first):

1. **Plan validity** — error vs successful plan (moduleRoot / go.mod).
2. **Discovery composition** — cmd-only, script-only, conflict, nested cmd,
   mixed multi-bin, empty.
3. Within cmd/script: **bin presence** (install vs skip) and naming rules.

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| S1 | plan/cmd/present-install | `./cmd/present` main + bin `present` → one `go-install` install |
| S2 | plan/cmd/absent-skip | `./cmd/missing` main, no bin file → one skip item |
| S3 | plan/script/nested-foo-present | `./script/foo/install` + bin `foo` → `go-run-install` install |
| S4 | plan/conflict/script-wins | both `./cmd/foo` and `./script/foo/install`, bin present → single script item |
| S5 | plan/script/bare-module-basename | bare `./script/install`, module `example.com/demo`, bin `demo` → install |
| S6 | plan/nested-cmd/present-install | `./cmd/nested/tool` + bin `tool` → `go-install` `./cmd/nested/tool` |
| S7 | plan/mixed/sorted-filter | several candidates; present=install, absent=skip; sorted by BinName |
| S8 | plan/empty/no-candidates | go.mod only, no cmd/script mains → empty Items, ok |
| S9 | error/no-go-mod | directory without go.mod → error |
| S10 | plan/cmd/non-main-ignored | `./cmd/lib` is `package lib` only → empty Items |

Multi-module leaves (M1–M5) are a **nested root** — see
`multi/DOCTEST.md` (not inherited by this Run contract). Scan-root resolution
+ module discovery (R1–R4) is a **nested root** — see `scan/DOCTEST.md`.

## How to Run

```sh
doctest vet ./cmd/wrk/tests/reinstall-local
doctest test ./cmd/wrk/tests/reinstall-local
doctest test ./cmd/wrk/tests/reinstall-local/plan/cmd/present-install
doctest test ./cmd/wrk/tests/reinstall-local/plan/conflict/script-wins
doctest test ./cmd/wrk/tests/reinstall-local/error/no-go-mod
# multi-module nested root:
doctest vet ./cmd/wrk/tests/reinstall-local/multi
doctest test ./cmd/wrk/tests/reinstall-local/multi
# scan-root nested root (Classic TDD RED until ResolveReinstallScanRoot /
# PlanLocalReinstallsFromWorkDir):
doctest vet ./cmd/wrk/tests/reinstall-local/scan
doctest test ./cmd/wrk/tests/reinstall-local/scan
```

Single-module tree is **GREEN** (`PlanLocalReinstalls` already landed). Multi
nested root covers explicit `moduleRoots` lists. Scan nested root expects
**RED** until implementer lands `ResolveReinstallScanRoot` /
`PlanLocalReinstallsFromWorkDir`. Do **not** rewrite sealed `plan/` / `error/`
or `multi/` ASSERT leaves when adding scan coverage.

```go
import (
	"testing"

	"github.com/xhd2015/wrk/wrkcli"
)

// Request drives PlanLocalReinstalls against a per-leaf fixture module + bin dir.
// Root Setup allocates WorkRoot / ModuleRoot / BinDir; leaves write go.mod,
// package mains, and stub bin files, then set Want* expectations.
type Request struct {
	WorkRoot   string
	ModuleRoot string
	BinDir     string

	// WantError: when true, Assert expects Run's error to be non-nil.
	WantError bool

	// WantModuleName is the expected plan ModuleName (module path basename).
	// Empty when WantError.
	WantModuleName string

	// WantItems is the full expected plan Items list (already sorted by BinName).
	// Nil/empty means zero items. Skip entries are included when expected.
	WantItems []WantPlanItem
}

// WantPlanItem is the assertable shape of one plan row.
type WantPlanItem struct {
	BinName string
	Method  string // "go-install" | "go-run-install"
	RelPath string // "./cmd/..." or "./script/.../install"
	Action  string // "install" | "skip"
}

// Response mirrors the pure plan for asserts (string fields for Method/Action).
type Response struct {
	ModuleRoot string
	ModuleName string
	BinDir     string
	Items      []WantPlanItem
}

func Run(t *testing.T, req *Request) (*Response, error) {
	// Classic TDD: PlanLocalReinstalls / LocalReinstallPlan are the production
	// API under design. RED (compile or assert) until implementer lands them.
	plan, err := wrkcli.PlanLocalReinstalls(req.ModuleRoot, req.BinDir)
	if err != nil {
		return &Response{}, err
	}
	if plan == nil {
		return &Response{}, nil
	}
	items := make([]WantPlanItem, len(plan.Items))
	for i, it := range plan.Items {
		items[i] = WantPlanItem{
			BinName: it.BinName,
			Method:  string(it.Method),
			RelPath: it.RelPath,
			Action:  string(it.Action),
		}
	}
	return &Response{
		ModuleRoot: plan.ModuleRoot,
		ModuleName: plan.ModuleName,
		BinDir:     plan.BinDir,
		Items:      items,
	}, nil
}
```
