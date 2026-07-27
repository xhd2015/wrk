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
- **Per-tree classification (intra-cmd / intra-script)** — for each BinName
  within **cmd** paths alone and within **script** paths alone:
  - **unique** = exactly one path for that BinName in that tree
  - **ambiguous** = two or more paths → emit a **warning** diagnostic and
    **drop** that tree’s candidates for that BinName (no contribution)
- **Merge (per BinName, after per-tree classification)**:

  | Cmd | Script | Plan Item | Diagnostics |
  |-----|--------|-----------|-------------|
  | unique | unique | **script** install | `notice` `prefer-script` |
  | unique | none | **cmd** | — |
  | none | unique | **script** | — |
  | ambiguous | unique | **script** | `warning` `ambiguous-cmd` |
  | unique | ambiguous | **cmd** | `warning` `ambiguous-script` |
  | ambiguous | none | **none** (omit) | `warning` `ambiguous-cmd` |
  | none | ambiguous | **none** (omit) | `warning` `ambiguous-script` |
  | ambiguous | ambiguous | **none** (omit) | both warnings |

  Ambiguous-only bins (no survivor) are **omitted** from `Items` — not a
  binDir skip row. When cmd is ambiguous and script is unique, keep script
  and emit **only** the ambiguous-cmd warning (**no** prefer-script notice).
  Symmetric for unique cmd + ambiguous script.
- **Filter** — for each surviving candidate, if `$binDir/<binName>` does **not**
  exist as a regular file (or a symlink that resolves to a file), set `Action`
  to `skip`; otherwise `install`. Skip entries remain in the plan (assertable).
- **Sort** — plan `Items` ordered lexicographically by `BinName`.
- **LocalReinstallPlan** — `{ ModuleRoot, ModuleName, BinDir, Items []PlanItem,
  Diagnostics []ReinstallDiagnostic }`. Each `PlanItem` has `BinName`, `Method`
  (`go-install` | `go-run-install`), `RelPath` (`./cmd/...` or
  `./script/.../install`), `Action` (`install` | `skip`).
- **ReinstallDiagnostic** — `{ Level, Kind, BinName, Paths }`:
  - **Level**: `"notice"` | `"warning"`
  - **Kind**: `"prefer-script"` | `"ambiguous-cmd"` | `"ambiguous-script"`
  - **BinName**: contested binary name
  - **Paths**: slash-form `./…` relative paths involved, **lexicographically sorted**
  - Prefer-script `Paths` include both the losing cmd path and the winning
    script path. Ambiguous kinds include only paths from the ambiguous tree.
- **Diagnostics ordering** — stable sort: by `BinName` ascending, then by
  `Kind` ascending (lexicographic: `ambiguous-cmd` < `ambiguous-script` <
  `prefer-script`). Empty when no notices/warnings.
- **Non-goals (P1)** — CLI flags, dry-run text / ANSI color, real install/run,
  `--force`, multi-module / `go.work` walk, mutual exclusion with other wrk modes.

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
│   ├── conflict/                       # cmd×script merge + ambiguity diagnostics
│   │   ├── script-wins/                # unique cmd+script → script + prefer-script notice
│   │   ├── bare-prefer-script/         # ./cmd/demo + bare ./script/install → notice
│   │   ├── ambiguous-cmd-skip/         # two cmd same bin, no script → omit + warning
│   │   ├── ambiguous-script-skip/      # two script same bin, no cmd → omit + warning
│   │   ├── ambiguous-cmd-script-unique/# two cmd + unique script → script + cmd warning only
│   │   ├── ambiguous-script-cmd-unique/# two script + unique cmd → cmd + script warning only
│   │   └── both-ambiguous/             # two cmd + two script → omit + both warnings
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
2. **Discovery composition** — cmd-only, script-only, conflict/merge, nested cmd,
   mixed multi-bin, empty.
3. Within conflict: **prefer-script** (nested + bare) vs **ambiguous skip** vs
   **fallback** (ambiguous side drops; unique other side wins) vs **both drop**.
4. Within cmd/script only: **bin presence** (install vs skip) and naming rules.

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| S1 | plan/cmd/present-install | `./cmd/present` main + bin `present` → one `go-install` install |
| S2 | plan/cmd/absent-skip | `./cmd/missing` main, no bin file → one skip item |
| S3 | plan/script/nested-foo-present | `./script/foo/install` + bin `foo` → `go-run-install` install |
| S4 | plan/conflict/script-wins | unique `./cmd/foo` + `./script/foo/install`, bin present → single script item + `prefer-script` notice |
| S4b | plan/conflict/bare-prefer-script | `./cmd/demo` + bare `./script/install`, module `…/demo` → script + prefer-script notice |
| S4c | plan/conflict/ambiguous-cmd-skip | `./cmd/foo` + `./cmd/nested/foo`, no script → empty Items; warning ambiguous-cmd |
| S4d | plan/conflict/ambiguous-script-skip | two script installs same bin, no cmd → empty Items; warning ambiguous-script |
| S4e | plan/conflict/ambiguous-cmd-script-unique | two cmd + unique script → script item; warning cmd only; **no** prefer-script |
| S4f | plan/conflict/ambiguous-script-cmd-unique | two script + unique cmd → cmd item; warning script only |
| S4g | plan/conflict/both-ambiguous | two cmd + two script → empty Items; both warnings |
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
doctest test ./cmd/wrk/tests/reinstall-local/plan/conflict/bare-prefer-script
doctest test ./cmd/wrk/tests/reinstall-local/plan/conflict/ambiguous-cmd-skip
doctest test ./cmd/wrk/tests/reinstall-local/plan/conflict/ambiguous-script-skip
doctest test ./cmd/wrk/tests/reinstall-local/plan/conflict/ambiguous-cmd-script-unique
doctest test ./cmd/wrk/tests/reinstall-local/plan/conflict/ambiguous-script-cmd-unique
doctest test ./cmd/wrk/tests/reinstall-local/plan/conflict/both-ambiguous
doctest test ./cmd/wrk/tests/reinstall-local/error/no-go-mod
# multi-module nested root:
doctest vet ./cmd/wrk/tests/reinstall-local/multi
doctest test ./cmd/wrk/tests/reinstall-local/multi
# scan-root nested root (Classic TDD RED until ResolveReinstallScanRoot /
# PlanLocalReinstallsFromWorkDir):
doctest vet ./cmd/wrk/tests/reinstall-local/scan
doctest test ./cmd/wrk/tests/reinstall-local/scan
```

Single-module tree: existing Items-only leaves stay **GREEN** where behavior is
unchanged; **conflict/** leaves that assert Diagnostics / ambiguity skip / prefer
notice expect **RED** (compile or assert) until implementer lands
`ReinstallDiagnostic`, `LocalReinstallPlan.Diagnostics`, and the merge table
above. Mixed RED/GREEN within the tree is OK before seal. Multi nested root
covers explicit `moduleRoots` lists. Scan nested root is independent. Do **not**
rewrite multi/scan sealed ASSERTs when adding diagnostics coverage.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"

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

	// WantDiagnostics is the full expected Diagnostics list (sorted by BinName,
	// then Kind). Nil/empty means zero diagnostics.
	WantDiagnostics []WantDiagnostic
}

// WantPlanItem is the assertable shape of one plan row.
type WantPlanItem struct {
	BinName string
	Method  string // "go-install" | "go-run-install"
	RelPath string // "./cmd/..." or "./script/.../install"
	Action  string // "install" | "skip"
}

// WantDiagnostic is the assertable shape of one plan diagnostic.
type WantDiagnostic struct {
	Level   string   // "notice" | "warning"
	Kind    string   // "prefer-script" | "ambiguous-cmd" | "ambiguous-script"
	BinName string
	Paths   []string // sorted slash-form ./ paths involved
}

// Response mirrors the pure plan for asserts (string fields for Method/Action/Level/Kind).
type Response struct {
	ModuleRoot  string
	ModuleName  string
	BinDir      string
	Items       []WantPlanItem
	Diagnostics []WantDiagnostic
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	// Classic TDD: PlanLocalReinstalls / LocalReinstallPlan / Diagnostics are the
	// production API under design. RED (compile or assert) until implementer lands them.
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
	diags := make([]WantDiagnostic, len(plan.Diagnostics))
	for i, d := range plan.Diagnostics {
		paths := append([]string(nil), d.Paths...)
		diags[i] = WantDiagnostic{
			Level:   string(d.Level),
			Kind:    string(d.Kind),
			BinName: d.BinName,
			Paths:   paths,
		}
	}
	return &Response{
		ModuleRoot:  plan.ModuleRoot,
		ModuleName:  plan.ModuleName,
		BinDir:      plan.BinDir,
		Items:       items,
		Diagnostics: diags,
	}, nil
}
```
