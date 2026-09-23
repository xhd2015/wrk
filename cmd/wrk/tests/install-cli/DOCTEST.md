# wrk --install — named local binary install (no binDir gate)

## Version
0.0.1

Decision tree for the `wrk --install name...` CLI surface: named install
resolution shared with `--reinstall-local`, but **forced** (no
GOBIN/GOPATH/bin presence gate) and **requiring at least one name**.

Scope split (MECE, significance-first — the product distinction nearest the
root first):

- **Forced install of a name that is not installed yet** — the reason the flag
  exists (`dry-run/install-absent-bin`, `dry-run/present-bin-forced`,
  `execute/present-bin-reinstalls`).
- **Resolution unchanged** — same discovery as `--reinstall-local`: name
  selection/order, `cmd/<name>` → `go install`, `script/<name>/install` →
  `go run`, nearest-`go.mod` re-root (`dry-run/select-names`,
  `dry-run/script-method`, `execute/installs-bin`).
- **Real install outcome** — progress lines + `installed N, failed F` summary;
  compile failure stays soft (exit 0 + `warning:`) (`execute/*`).
- **Scan root choice** — `--main` scans the main repo modules; without it the
  linked worktree checkout (`main/from-linked-wt/*`).
- **Flag surface** — required arg, unknown/ambiguous/colliding name,
  mutual exclusion (`error/*`), help (`help/mentions-flag`).
- **Observability** — `events.jsonl` command identity (`events/dry-run`).

## DSN (Domain Specific Notion)

- **wrk --install name...** — exclusive top-level mode: install the named local
  module binaries resolved from the current checkout. `Varargs` with
  `WithMinimum(1)`: names are arbitrary non-flag tokens after `--install`
  (repeatable, stops at the next `-` token), exactly like `--bring p1 p2`.
  Bare `--install` (or `--install` followed only by flags) is a parse error
  with the **library wording** `--install requires a value`.
- **Partners** — `--main` (scan main repo modules), `--dry-run`,
  `--color` / `--no-color`, `-v`. Everything else is rejected: primary modes
  (`wrk: --install is mutually exclusive with other modes`) and
  `--reinstall-local` (`wrk: --install is mutually exclusive with
  --reinstall-local`). Never a pipeline/compose partner (unlike
  `--reinstall-local`, which composes with `--done` / `--merge-back`).
- **Resolution (shared with --reinstall-local)** — binDir is `GOBIN` if set,
  else `$(go env GOPATH)/bin`; the scan root is `PlanLocalReinstallsFromWorkDir`
  (worktree toplevel, or the main repo with `--main`, or go.mod walk-up outside
  git); discovery is `cmd/<name>` (`go-install`) with `script/<name>/install`
  (`go-run-install`) preferred when present, plus nested-script fallback and
  the standard `notice:` / `warning:` diagnostics on stderr. Install paths are
  re-rooted to the package dir's nearest `go.mod` (nested-module shape).
- **No preinstall gate** — the requested names are looked up among discovered
  candidates and every selected item is forced to `Action=install`, so a name
  installs even when `$GOBIN/<name>` does not exist. `skip:` lines never appear
  in `--install` output.
- **Name errors (hard, non-zero, prefixed `wrk: --install:`)** — a name with no
  discovered candidate → `no install candidate for "<name>"`; a name whose
  candidates are ambiguous (`./cmd/foo` + `./cmd/nested/foo`) → `bin "<name>"
  is ambiguous (<paths>)`; a name claimed by two modules → `bin "<name>"
  claimed by multiple modules: <root> (<mod>) and <root> (<mod>)`.
- **Dry-run stdout** — for each selected item, one line re-rooted to the
  package's own module: `would: go install <ownRel>` or
  `would: go run <ownRel>`. Diagnostics (if any) go to **stderr** first.
  Summary is the last stdout line:
  - one module: `would: install N binaries`
  - K>1 modules: `# module <ModulePath> (<RelDir>)` before each module's items,
    then `would: install N binaries across K modules`
- **Execute stdout** — one progress line per attempt (no `would:` prefix, no
  `# module` headers): `go install <ownRel>` / `go run <ownRel>`; child `go`
  output streams to stderr. Summary is the last stdout line:
  `installed N, failed F` (no skipped segment — a forced install never skips).
- **Execute exit code (soft failures)** — exit **0** even when `failed > 0`;
  then stderr also carries `warning: install finished with F failed`
  (prefix-only coloring under `--color`). Hard plan errors (unknown/ambiguous/
  colliding name, no modules, no go.mod) remain non-zero.
- **Module order** — modules in multi-plan order (lexicographic absolute
  ModuleRoot); items within a module in **request order** after dedupe.
- **events.jsonl** — every successful/failing run appends one entry with
  `command: "install"` and `args` including `--install` and each name (plus
  `--dry-run` / `--main` when passed). `wrk -h` skips the event.
- **Layer policy** — leaves are **L2 in-process** (`wrkcli.Capture`,
  `req.InProcess = true`) except `execute/installs-bin`, which is **L3**
  (`label: e2e`, real product binary + real `go install` into an isolated
  GOBIN).

## Tree Overview

```
install-cli/
├── dry-run/                        # forced install plan (L2, zero mutation)
│   ├── install-absent-bin/         # name absent from GOBIN → would: go install (no skip:)
│   ├── present-bin-forced/         # stub present → still a forced install (no skip:)
│   ├── select-names/               # two bins, one requested → only that item
│   └── script-method/              # script/<name>/install → would: go run
├── execute/                        # real installs (isolated GOBIN)
│   ├── installs-bin/               # L3 e2e: product binary installs; bin runs
│   ├── present-bin-reinstalls/     # L2: stub replaced by real install; no skip:
│   └── soft-failure/               # L2: compile failure → exit 0 + warning:, failed 1
├── main/
│   └── from-linked-wt/             # diverged main vs linked worktree fixture
│       ├── with-main/              # --main --install … → main modules (K=2)
│       └── without-main/           # contrast: worktree checkout (K=1)
├── error/
│   ├── requires-arg/               # bare --install → requires a value
│   ├── unknown-name/               # no install candidate
│   ├── ambiguous-cmd/              # cmd/foo + cmd/nested/foo
│   ├── cross-module-collision/     # two modules claim the same bin
│   └── exclusive/
│       ├── with-list/              # --install … --list
│       └── with-reinstall-local/   # --install … --reinstall-local
├── events/
│   └── dry-run/                    # events.jsonl command "install" + args
└── help/
    └── mentions-flag/              # wrk -h documents --install
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| I1 | dry-run/install-absent-bin | `--install tool --dry-run`, no GOBIN/tool → `would: go install ./cmd/tool` + `would: install 1 binaries`; no `skip:`; no mutation |
| I2 | dry-run/present-bin-forced | stub `$GOBIN/tool` + `--install tool --dry-run` → same forced plan (gate is not consulted); stub unchanged |
| I3 | dry-run/select-names | `cmd/tool` + `cmd/other`, `--install tool --dry-run` → only `./cmd/tool`; `other` absent from stdout |
| I4 | dry-run/script-method | only `script/tool/install` main → `would: go run ./script/tool/install` |
| I5 | execute/installs-bin | **L3**: real binary + real `go install`; `$GOBIN/tool` runs `tool-ok`; `installed 1, failed 0`; exit 0 |
| I6 | execute/present-bin-reinstalls | stub present → real install replaces it; `installed 1, failed 0`; no `skip:` |
| I7 | execute/soft-failure | `cmd/broken` does not compile → exit 0; summary `installed 0, failed 1`; stderr `warning: install finished with 1 failed` |
| I8 | main/from-linked-wt/with-main | linked WT cwd + `--main --install mainbin toolbin --dry-run` → multi plan for **main** modules (K=2) |
| I9 | main/from-linked-wt/without-main | same cwd without `--main` → worktree-only K=1 plan (`wtbin`) |
| I10 | error/requires-arg | bare `--install` → non-zero; stderr `requires a value`; stdout empty |
| I11 | error/unknown-name | `--install nope` → non-zero; stderr `wrk: --install:` + `no install candidate` + `nope` |
| I12 | error/ambiguous-cmd | `cmd/dup` + `cmd/nested/dup` → non-zero; stderr `ambiguous` + both paths |
| I13 | error/cross-module-collision | mod-a + mod-b both `cmd/same` → non-zero; stderr `multiple modules` + both module names |
| I14 | error/exclusive/with-list | `--install tool --list` → non-zero; `mutually exclusive`; stdout empty |
| I15 | error/exclusive/with-reinstall-local | `--install tool --reinstall-local` → non-zero; stderr names both flags |
| I16 | events/dry-run | success dry-run → last event `command: "install"`, `exit_code: 0`, args include `--install`, `tool`, `--dry-run` |
| I17 | help/mentions-flag | `wrk -h` → exit 0; help contains `--install`, `at least one name`, and keeps `--reinstall-local` |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/install-cli
doctest test ./cmd/wrk/tests/install-cli
doctest test ./cmd/wrk/tests/install-cli/dry-run
doctest test ./cmd/wrk/tests/install-cli/execute
doctest test ./cmd/wrk/tests/install-cli/execute/installs-bin
doctest test ./cmd/wrk/tests/install-cli/main/from-linked-wt
doctest test ./cmd/wrk/tests/install-cli/error
doctest test ./cmd/wrk/tests/install-cli/events
doctest test ./cmd/wrk/tests/install-cli/help
```

Sealed regression neighbours (must stay GREEN; `--install` shares the named
lookup and execute engine with them):

```sh
doctest test ./cmd/wrk/tests/reinstall-local-cli
doctest test ./cmd/wrk/tests/reinstall-local
go test ./wrkcli/...
```

## Harness

`Request.InProcess = true` runs through `wrkcli.Capture` (L2) with isolated
`WRK_HOME` / `GOBIN`; leave it `false` only for the L3 execute leaf, which
builds the session product binary once and runs it as a child process. The root
`Setup` allocates `WorkRoot`, `WrkHome`, `ModuleRoot`, and `BinDir`; leaves
write fixtures and set `Args`.

```go
import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

// Request drives wrk --install under GOBIN isolation.
// Root Setup allocates WorkRoot / WrkHome / ModuleRoot / BinDir (gobin).
// Leaves write go.mod + package mains + optional stub bins, then set Args.
type Request struct {
	WorkRoot   string
	WrkHome    string
	ModuleRoot string // process cwd for install leaves (module root, git worktree, subdir)
	BinDir     string // GOBIN for this leaf ({WorkRoot}/gobin)
	Args       []string
	ExtraEnv   []string // additional KEY=VAL (GOBIN is always set by Run)

	// InProcess runs via wrkcli.Capture (L2 short path) instead of the product
	// binary. Default true; only the L3 leaf (execute/installs-bin) leaves it false.
	InProcess bool
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	adoptDoctestContext(d)
	args := append([]string(nil), req.Args...)

	if req.InProcess {
		res := wrkcli.Capture(wrkcli.CaptureOpts{
			Args: args,
			Dir:  req.ModuleRoot,
			Env:  installCLIEnv(req),
		})
		return &Response{
			Stdout:   res.Stdout,
			Stderr:   res.Stderr,
			ExitCode: res.ExitCode,
		}, nil
	}

	bin := getWrkBin(t)

	cmd := exec.Command(bin, args...)
	cmd.Dir = req.ModuleRoot
	cmd.Env = installCLIEnv(req)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}

	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
```
