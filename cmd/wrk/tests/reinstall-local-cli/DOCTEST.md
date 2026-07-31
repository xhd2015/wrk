# wrk --reinstall-local — CLI dry-run + execute + multi + --main compose (P6)

## Version
0.0.8

Decision tree for **Phase 2** CLI dry-run surface, **Phase 3** execute path,
**Phase 4** hardening (events.jsonl), **Phase 3 multi-module dry-run**,
**Phase 4 product compose** (`--main --reinstall-local`), **Phase 5
multi-module execute**, and **Phase 6** events with `--main` (args include
`--main`). Builds and runs the real `wrk` binary (session fixture cache), with
**GOBIN** isolation for bin-dir filtering.

Depends on P1 pure API `wrkcli.PlanLocalReinstalls` / multi
`PlanLocalReinstallsFromWorkDir` (under `cmd/wrk/tests/reinstall-local/`).
This tree is a **sibling nested root** so P1 sealed pure-API asserts stay
untouched.

**P2 (sealed):** `--reinstall-local --dry-run` prints a full reinstall plan and
mutates nothing (**single-module** walk-up format preserved when K=1).

**P3 execute (sealed except soft-exit update):** bare `wrk --reinstall-local`
(without `--dry-run`) runs planned `go install` / `go run` commands sequentially,
continues on failure, and prints an execute summary. When `failed > 0`, prints
stderr `warning:` (mentions reinstall/fail) and exits **0** (soft; not
ExitCodeError 1).

**P3 multi dry-run (sealed when GREEN):** wire CLI dry-run to
`PlanLocalReinstallsFromWorkDir(workDir, binDir, useMain=false)` and print
**grouped multi-module** dry-run output when more than one module is planned.
Single-module (K=1) keeps the sealed single-mod line format (no `# module`
headers; summary without `across K modules`) so existing `dry-run/*` leaves
stay GREEN.

**P4 (hardening, sealed):** successful dry-run / execute appends `events.jsonl` with
`command: "reinstall-local"` (and dry-run records `--dry-run` in `args`).

**P4 product compose (sealed):** `wrk --main --reinstall-local [--dry-run]` is
**allowed** (no nested shell). Planning uses
`PlanLocalReinstallsFromWorkDir(workDir, binDir, useMain=true)` so scan root is
the **main repository** of the checkout. Flag order free
(`--reinstall-local --main` same as `--main --reinstall-local`). Without
`--main`, linked worktree cwd still plans the worktree checkout
(`useMain=false`). Compose remains exclusive with other modes (e.g. `--list`).
Leaves under `main-compose/`; **do not** weaken sealed P2/P3/P4 ASSERT bodies.

**P5 multi execute (coverage backfill — GREEN if P3 already runs multi plan):**
bare `wrk --reinstall-local` uses the **same multi plan** as dry-run
(`PlanLocalReinstallsFromWorkDir`) then `executeMultiLocalReinstalls`: for each
module in plan order, run install/skip items with `Dir=ModuleRoot`. Continue
after failures; one summary totals N/M/F across modules (no `across K modules`
suffix — that remains dry-run-only). Cross-module install×install collision
fails at plan time **before** any `go install`. New leaves under
`execute-multi/`; sealed `execute/*` and `dry-run-multi/*` ASSERTs stay intact.
E3 dry-run multi regression remains `dry-run-multi/*` (zero mutation).

**P6 events + --main (coverage backfill — GREEN expected):** successful
`wrk --main --reinstall-local --dry-run` appends `events.jsonl` with
`command: "reinstall-local"` and `args` including `--main`,
`--reinstall-local`, and `--dry-run` (same `extractEventArgs` path as other
modes). H1 remains `events/dry-run` (no `--main`). New leaf
`events/main-dry-run` (H2). Sealed H1/execute event ASSERTs stay intact.

**Conflict diagnostics (Classic TDD):** plan diagnostics (prefer-script notice,
ambiguous-cmd/script warnings) print on **stderr** during dry-run/execute.
Default pipe harness: plain prefixes (no ANSI). With `--color`: colorize only
the prefix token (`notice:` grey `#90`, `warning:` orange `#33`); rest of line
plain. Exit 0 for successful dry-run plans that include diagnostics (non-fatal).

**Out of scope:** full root DOCTEST.md mega-doc, `--force`, parallel installs,
JSON CLI, bare `--main` nested shell e2e, FORCE_COLOR / CLICOLOR, coloring
`would:` / `go install` stdout lines.

# DSN (Domain Specific Notion)

- **wrk --reinstall-local** — top-level mode flag. Mutually exclusive with
  other modes (e.g. `--list`, `--status`) **except** composition with
  **`--main`** (P4). Help (`wrk -h` / `wrk --help`) documents
  `--reinstall-local`.
- **wrk --main --reinstall-local [--dry-run]** — compose: no nested shell.
  Same bin-dir resolution and dry-run / execute printers as bare
  `--reinstall-local`, but plan via
  **`PlanLocalReinstallsFromWorkDir(cwd, binDir, useMain=true)`** so scan root
  is **ResolveMainRepo(ShowToplevel(cwd))** (main repository path). Flag order
  free. Still exclusive with unrelated modes (`--list`, etc.). Event command
  remains `"reinstall-local"`; event `args` include `--main` when that flag was
  passed (P6).
- **wrk --reinstall-local --dry-run** — resolve bin dir as **GOBIN if set, else
  `$(go env GOPATH)/bin`**; plan via **`PlanLocalReinstallsFromWorkDir(cwd,
  binDir, useMain=false)`** unless `--main` is also set (then `useMain=true`);
  scan all modules under the resolved scan root; print plan lines; exit 0 on
  successful plan (including N=0 install actions). **No** `go install` /
  `go run` of candidates in dry-run. Cross-module install×install collision →
  non-zero, stderr names the bin (and identifying modules). Legacy
  single-module walk-up-only planning is superseded by the multi scan
  entrypoint; **K=1 output format is preserved**.
- **would: lines (install actions)** — for each plan item with `Action=install`,
  one stdout line:
  - method `go-install` → `would: go install <RelPath>`
  - method `go-run-install` → `would: go run <RelPath>`
  Optional trailing ` # <bin>` comment is allowed by product but **not** required
  by these asserts (locked form has no comment).
- **skip: lines** — for each plan item with `Action=skip`:
  `skip: <bin> (not in <bindir>)` where `<bindir>` is the resolved bin directory
  path (absolute under test isolation). Same wording in dry-run and execute.
- **Diagnostics on stderr** — when the plan has diagnostics, print one line per
  diagnostic on **stderr** (not stdout), exit still 0 for successful dry-run:
  - prefer-script:
    `notice: bin <bin>: preferring <scriptPath> over <cmdPath>`
  - ambiguous-cmd:
    `warning: bin <bin>: ambiguous under cmd (<path1>, <path2>, ...); skipping`
  - ambiguous-script:
    `warning: bin <bin>: ambiguous under script (<path1>, <path2>, ...); skipping`
  Paths inside messages: slash-form `./…`, **lexicographically sorted** in
  parenthesized lists. Prefer-script uses the script path first then cmd path
  (message order), not the diagnostic Paths sort order for the phrase.
  Prefix tokens exactly `notice:` and `warning:`.
- **Diagnostic color (CLI only)** — when color is enabled for diagnostics:
  - colorize **only** the prefix token `notice:` / `warning:`
  - `notice:` → grey (`ansiGrey` `\x1b[90m` … `\x1b[0m`)
  - `warning:` → orange/yellow (`ansiOrange` `\x1b[33m` … `\x1b[0m`)
  - rest of the line plain; never put ANSI on stdout plan lines
  Color enablement (aligned with go-best-practice / existing wrkcli):
  - `--color` → always on for these prefixes
  - auto (no `--color`): on when stderr is a TTY and `NO_COLOR` is empty;
    pipe harness → plain prefixes
  Harness strips ambient `NO_COLOR` from the process env so leaves own color
  policy; force `--color` for color-on asserts.
- **Single-module dry-run (K=1)** — when the multi plan has exactly one module:
  print items **without** `# module` headers (same order as plan: lex by
  BinName), then summary:
  `would: reinstall N binaries (M skipped)\n`
  (no `across 1 modules` suffix). This is the sealed format of `dry-run/*`.
- **Multi-module dry-run (K>1)** — for each module in multi-plan order
  (lexicographic absolute ModuleRoot):
  1. Header line: `# module <ModulePath> (<RelDir>)`
     - **ModulePath** = full `module` path from that module's `go.mod`
       (e.g. `example.com/cli-multi-tools`), not only the basename.
     - **RelDir** = path of the module root relative to the scan root, slash-form
       when nested; **`.`** when the module root is the scan root itself.
  2. That module's `would:` / `skip:` item lines (lex by BinName within module).
  Then summary last line:
  `would: reinstall N binaries (M skipped) across K modules\n`
  where N/M are totals across all modules and K = `len(Modules)`.
  Implementer may extend `ModuleReinstallPlan` with `ModulePath` / compute
  RelDir from scan root + ModuleRoot — field shape is implementer-owned; CLI
  asserts lock the stdout vocabulary only.
- **Dry-run summary line** — always last content line of successful dry-run stdout.
- **wrk --reinstall-local (execute)** — same bin-dir resolution and the **same
  multi plan** as dry-run via `PlanLocalReinstallsFromWorkDir(cwd, binDir,
  useMain)` (`useMain` true only with `--main`). Then
  `executeMultiLocalReinstalls`: for each module in multi-plan order (lex
  absolute ModuleRoot), for each plan item (lex BinName within module):
  - `Action=install` + method `go-install` → run `go install <RelPath>` with
    `Dir=that module's ModuleRoot`, stream child stdout/stderr to parent
    streams; count success as reinstalled or failure as failed.
  - `Action=install` + method `go-run-install` → run `go run <RelPath>` with
    `Dir=ModuleRoot`, stream child stdout/stderr; count reinstalled/failed.
  - `Action=skip` → do **not** invoke go; count skipped.
  Continue after a failed install (do not abort remaining items or remaining
  modules). Cross-module install×install collision fails at plan time (non-zero
  stderr; **no** `go install` / no execute summary).
- **Execute progress lines** — for each install attempt, one wrk-owned stdout
  line before/as the command runs (mirrors dry-run without the `would:` prefix;
  **no** `# module` headers on execute):
  - `go install <RelPath>`
  - `go run <RelPath>`
  Child compiler/tool output is streamed (typically on stderr) and is **not**
  locked to exact match in execute leaves.
- **Execute summary line** — always last content line of execute stdout:
  `reinstalled N, skipped M, failed F\n`
  where N/M/F are totals across **all** modules in the multi plan (no
  `across K modules` suffix — that is dry-run-only).
- **Execute exit code (soft failures)** — exit **0** always after a finished
  execute plan, including when `failed > 0`. When `failed > 0`, also print a
  **warning** on **stderr** with prefix `warning:` that mentions reinstall
  and/or fail (exact wording implementer-owned; may share stderr with streamed
  child `go` noise). Hard errors (unknown method, plan failure, no go.mod,
  cross-module install×install collision) remain non-zero as today. All-skip
  success (N=0, F=0) stays exit 0 with **no** soft-failure warning required.
- **Item print order** — single-mod (K=1): lexicographic by bin name
  (interleaved install/skip by name), then summary. Multi-mod dry-run: modules
  by ModuleRoot lex, items by BinName within each module, then summary with
  `across K modules`. Multi-mod execute: same module/item order as multi
  dry-run items, but progress lines only (no `# module` headers), then execute
  summary.
- **GOBIN isolation** — harness sets `GOBIN={WorkRoot}/gobin` via `ExtraEnv` so
  filter never touches the developer machine's real bin dir. Stub binaries are
  plain files under that gobin; execute path replaces stubs via real `go install`.
- **Bare --dry-run** — `wrk --dry-run` without a host mode is rejected; after P2
  the host-list error must mention `--reinstall-local` among valid hosts
  (substring checks; exact host-list order is implementer-owned).
- **No go.mod** — cwd (and ancestors / scan root) without parseable modules →
  non-zero exit and a clear error (stderr mentions `go.mod` or module resolution).
- **WRK_HOME** — isolated per leaf at `{WorkRoot}/.wrk`; reinstall-local does not
  require git for single-mod fixtures; multi from-subdir leaves use real git.
- **events.jsonl** — successful `--reinstall-local` (dry-run, execute, or
  compose with `--main`) appends one JSON line with `command: "reinstall-local"`,
  `exit_code` matching the process exit, and `args` including the CLI flags
  (`--reinstall-local`; also `--dry-run` when that modifier was passed; also
  `--main` when compose was used). `work_dir` is the resolved process cwd
  (module/repo root under fixtures). Help (`-h`) still skips event append.

## Tree Overview

```
reinstall-local-cli/
├── dry-run/                         # happy-path CLI dry-run single-mod (P2) + conflict diags
│   ├── present-install/             # C1 / E4 / C2: present bin → would: go install + summary; no mutation
│   ├── skip-only/                   # C2-orig: absent bin → skip + reinstall 0
│   ├── script-wins/                 # C3-orig: cmd+script → would: go run + stderr prefer-script notice (plain)
│   ├── ambiguous-cmd/               # two cmd same bin → warning stderr; reinstall 0; no go install
│   ├── ambiguous-script-fallback-cmd/ # two script + unique cmd → warning script; would: go install cmd
│   ├── prefer-script-color/         # --color + conflict → grey notice: on stderr; stdout plain
│   └── warning-color/               # --color + ambiguous cmd → orange warning: on stderr
├── dry-run-multi/                   # multi-module dry-run (P3 multi, sealed) — P5 E3 regression
│   ├── nested-modules/              # C1: root + tools; headers + both would; across 2 modules
│   ├── install-collision/           # C3: install×install same bin → non-zero; stderr names bin
│   └── from-subdir/                 # C4: git cwd=pkg/sub still plans both modules
├── execute/                         # real --reinstall-local without --dry-run (P3, sealed single-mod)
│   ├── present-install/             # E1-s: go install runs; bin executable; reinstalled 1
│   ├── skip-only/                   # E2-s: no go install; skipped only; exit 0
│   └── continue-on-failure/         # E3-s: one compile fail among installs; soft exit 0 + warning:; continue
├── execute-multi/                   # multi-module execute (P5)
│   ├── nested-modules/              # E1: root + tools both install into GOBIN; reinstalled 2
│   ├── continue-on-failure/         # E2: fail in root; tools still installs; soft exit 0 + warning:
│   └── install-collision/           # collision: plan error before any go install; stub unchanged
├── events/                          # P4/P6 hardening: events.jsonl command identity
│   ├── dry-run/                     # H1: dry-run success → command=reinstall-local; args include --dry-run
│   ├── execute/                     # execute success → command=reinstall-local; args without --dry-run
│   └── main-dry-run/                # H2 (P6): --main --reinstall-local --dry-run → args include --main
├── main-compose/                    # P4 product: --main + --reinstall-local (Classic TDD)
│   ├── from-linked-wt/              # diverged main vs linked-wt modules
│   │   ├── main-then-reinstall/     # MC1: --main --reinstall-local --dry-run → main modules
│   │   ├── reinstall-then-main/     # MC3: --reinstall-local --main --dry-run → same
│   │   └── without-main/            # contrast: no --main → worktree modules (K=1)
│   └── exclusive/
│       └── with-list/               # MC2: --main --reinstall-local --list → exclusive
├── mutual-exclusion/
│   └── with-list/                   # C4-orig: --reinstall-local --list → exclusive
├── dry-run-host/
│   └── bare-dry-run/                # C5: bare --dry-run mentions --reinstall-local
├── error/
│   └── no-go-mod/                   # C6: no go.mod → non-zero clear error
└── help/
    └── mentions-flag/               # C7: wrk -h contains --reinstall-local
```

Split factor (MECE, significance-first):

1. **Mode** — dry-run plan (single vs multi) vs real execute (single vs multi) vs events vs main-compose vs flag surface / errors.
2. **Within single dry-run** — present install / skip-only / script-wins notice / ambiguous / color.
3. **Within multi dry-run** — happy nested modules / install×install collision / from git subdir.
4. **Within multi execute** — both modules install / continue across module failure / collision before install.
5. **Events** — dry-run / execute / main-compose dry-run success recording of
   `command: "reinstall-local"` (P6 locks `--main` in args for compose).
6. **Main compose** — linked-WT scan-root choice (main vs worktree) / flag order / still exclusive with --list.
7. **Flag surface** — mutual exclusion, bare dry-run host list, help text.
8. **Module resolution error** — missing go.mod.

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| C1-s / E4 / **C2** | dry-run/present-install | `./cmd/present` + GOBIN/present → `would: go install ./cmd/present`; summary N=1 M=0 **without** `across`; exit 0; **stub unchanged** (single-mod regression lock for multi wire-up) |
| C2-s | dry-run/skip-only | `./cmd/missing` no bin → `skip: missing (not in <gobin>)`; `would: reinstall 0 binaries (1 skipped)`; exit 0 |
| C3-s | dry-run/script-wins | cmd+script foo + bin → `would: go run ./script/foo/install` only; stderr plain `notice: bin foo: preferring ./script/foo/install over ./cmd/foo`; exit 0 |
| C3-amb | dry-run/ambiguous-cmd | two cmd foo, no script → warning ambiguous cmd on stderr; reinstall 0; no `go install`; exit 0 |
| C3-fb | dry-run/ambiguous-script-fallback-cmd | two script + unique cmd → warning script; `would: go install ./cmd/foo`; exit 0 |
| C3-nc | dry-run/prefer-script-color | `--color` + unique conflict → stderr grey `notice:`; stdout plan uncolored |
| C3-wc | dry-run/warning-color | `--color` + ambiguous cmd → stderr orange `warning:`; stdout uncolored |
| **C1** | dry-run-multi/nested-modules | root + nested tools, both bins present → `# module … (.)` + `# module … (tools)` + both `would: go install`; `across 2 modules`; stubs unchanged |
| **C3** | dry-run-multi/install-collision | nested mod-a + mod-b both `./cmd/samebin` + GOBIN/samebin → non-zero; stderr contains `samebin`, `mod-a`, `mod-b` |
| **C4** | dry-run-multi/from-subdir | git repo root+tools; cwd=`pkg/sub` → same multi dry-run as full tree; both modules; `across 2 modules` |
| E1-s | execute/present-install | `./cmd/tool` prints version + GOBIN stub → real `go install`; exit 0; GOBIN/tool executable and runs; summary `reinstalled 1, skipped 0, failed 0` |
| E2-s | execute/skip-only | `./cmd/missing` no bin → no go; exit 0; `skip:` + `reinstalled 0, skipped 1, failed 0` |
| E3-s | execute/continue-on-failure | present `broken` (does not compile) + `good` (ok) → continue; exit 0; stderr `warning:` (reinstall/fail); good installed; `failed 1` in summary |
| **E1** | execute-multi/nested-modules | root + nested tools both bins present → `go install` each with per-module Dir; both GOBIN bins run; `reinstalled 2, skipped 0, failed 0` |
| **E2** | execute-multi/continue-on-failure | root `broken` fails compile; tools `toolgood` still installs; exit 0; stderr `warning:`; `reinstalled 1, skipped 0, failed 1` |
| E-coll | execute-multi/install-collision | mod-a + mod-b both `./cmd/samebin` + GOBIN → non-zero **before** install; stub unchanged; stderr names `samebin`/`mod-a`/`mod-b` |
| **E3** | dry-run-multi/* (regression) | multi dry-run still zero mutation; `across K modules` / collision plan errors unchanged (sealed C1/C3/C4) |
| H1 | events/dry-run | success dry-run → `events.jsonl` last event `command: "reinstall-local"`, `exit_code: 0`, args include `--reinstall-local` and `--dry-run` |
| H-exec | events/execute | success execute (skip-only) → last event `command: "reinstall-local"`, args include `--reinstall-local`, **not** `--dry-run` |
| **H2** | events/main-dry-run | `--main --reinstall-local --dry-run` (git module) → last event `command: "reinstall-local"`, args include `--main`, `--reinstall-local`, and `--dry-run` |
| C4-s | mutual-exclusion/with-list | `wrk --reinstall-local --list` → non-zero; mutually exclusive |
| C5 | dry-run-host/bare-dry-run | `wrk --dry-run` → non-zero; stderr mentions `--reinstall-local` host |
| C6 | error/no-go-mod | cwd without go.mod → non-zero; clear error |
| C7 | help/mentions-flag | `wrk -h` → exit 0; help contains `--reinstall-local` |
| **MC1** | main-compose/from-linked-wt/main-then-reinstall | linked WT + `--main --reinstall-local --dry-run` → main multi plan (`mainbin`+`toolbin`); not `wtbin` |
| **MC3** | main-compose/from-linked-wt/reinstall-then-main | `--reinstall-local --main --dry-run` same main multi plan (flag order free) |
| MC-contrast | main-compose/from-linked-wt/without-main | linked WT + bare `--reinstall-local --dry-run` → wt-only K=1 plan (`wtbin`) |
| **MC2** | main-compose/exclusive/with-list | `--main --reinstall-local --list` → non-zero; mutually exclusive |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/reinstall-local-cli
doctest test ./cmd/wrk/tests/reinstall-local-cli
doctest test ./cmd/wrk/tests/reinstall-local-cli/dry-run/present-install
doctest test ./cmd/wrk/tests/reinstall-local-cli/dry-run/script-wins
doctest test ./cmd/wrk/tests/reinstall-local-cli/dry-run/ambiguous-cmd
doctest test ./cmd/wrk/tests/reinstall-local-cli/dry-run/ambiguous-script-fallback-cmd
doctest test ./cmd/wrk/tests/reinstall-local-cli/dry-run/prefer-script-color
doctest test ./cmd/wrk/tests/reinstall-local-cli/dry-run/warning-color
doctest test ./cmd/wrk/tests/reinstall-local-cli/dry-run-multi
doctest test ./cmd/wrk/tests/reinstall-local-cli/dry-run-multi/nested-modules
doctest test ./cmd/wrk/tests/reinstall-local-cli/dry-run-multi/install-collision
doctest test ./cmd/wrk/tests/reinstall-local-cli/dry-run-multi/from-subdir
doctest test ./cmd/wrk/tests/reinstall-local-cli/execute/present-install
doctest test ./cmd/wrk/tests/reinstall-local-cli/execute/skip-only
doctest test ./cmd/wrk/tests/reinstall-local-cli/execute/continue-on-failure
doctest test ./cmd/wrk/tests/reinstall-local-cli/execute-multi
doctest test ./cmd/wrk/tests/reinstall-local-cli/execute-multi/nested-modules
doctest test ./cmd/wrk/tests/reinstall-local-cli/execute-multi/continue-on-failure
doctest test ./cmd/wrk/tests/reinstall-local-cli/execute-multi/install-collision
doctest test ./cmd/wrk/tests/reinstall-local-cli/events
doctest test ./cmd/wrk/tests/reinstall-local-cli/events/dry-run
doctest test ./cmd/wrk/tests/reinstall-local-cli/events/execute
doctest test ./cmd/wrk/tests/reinstall-local-cli/events/main-dry-run
doctest test ./cmd/wrk/tests/reinstall-local-cli/main-compose
doctest test ./cmd/wrk/tests/reinstall-local-cli/main-compose/from-linked-wt/main-then-reinstall
doctest test ./cmd/wrk/tests/reinstall-local-cli/main-compose/from-linked-wt/reinstall-then-main
doctest test ./cmd/wrk/tests/reinstall-local-cli/main-compose/from-linked-wt/without-main
doctest test ./cmd/wrk/tests/reinstall-local-cli/main-compose/exclusive/with-list
doctest test ./cmd/wrk/tests/reinstall-local-cli/mutual-exclusion/with-list
doctest test ./cmd/wrk/tests/reinstall-local-cli/dry-run-host/bare-dry-run
doctest test ./cmd/wrk/tests/reinstall-local-cli/error/no-go-mod
doctest test ./cmd/wrk/tests/reinstall-local-cli/help/mentions-flag
```

**Coverage (conflict diagnostics):** expect **RED** (or mixed) on
`dry-run/script-wins` (stderr notice), `dry-run/ambiguous-*`, and color leaves
until implementer prints plan Diagnostics on stderr with optional ANSI under
`--color`. Items-only stdout leaves without diagnostic asserts stay GREEN.

**Coverage (P5 multi execute):** expect **GREEN** on happy-path
`execute-multi/nested-modules` and `install-collision` when multi execute
already works. **Classic TDD soft-exit:** expect **RED** on
`execute/continue-on-failure` and `execute-multi/continue-on-failure` until
implementer changes `executeLocalReinstalls` /
`executeMultiLocalReinstalls` so `failed > 0` prints stderr `warning:` and
returns nil (exit 0) instead of `ExitCodeError{Code: 1}`. Do **not** weaken
sealed success-path `execute/*` / `dry-run-multi/*` ASSERT bodies.

**Coverage (P4 main-compose):** expect **GREEN** when flag mutex allows
`--main` + `--reinstall-local` and `runReinstallLocal` passes `useMain=true`.
`without-main` contrast and `exclusive/with-list` stay sealed. Do **not**
weaken sealed P2/P3/P4 ASSERT bodies for single-mod / multi / execute / events.

**Coverage (P6 events + --main):** expect **GREEN** on `events/main-dry-run`
(H2) when `extractEventArgs` records `--main` and `resolveCommand` keeps
`command: "reinstall-local"`. H1 `events/dry-run` and `events/execute` remain
sealed. Do **not** invent false RED.

P1 pure API tree stays independent:

```sh
doctest test ./cmd/wrk/tests/reinstall-local
```

```go
import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

// Request drives the wrk binary under GOBIN isolation.
// Root Setup allocates WorkRoot / WrkHome / ModuleRoot / BinDir (gobin).
// Leaves write go.mod + package mains + optional stub bins, then set Args.
// Multi dry-run leaves may reassign ModuleRoot to a git repo or subdir.
type Request struct {
	WorkRoot   string
	WrkHome    string
	ModuleRoot string // process cwd for reinstall-local leaves (module or subdir fixture)
	BinDir     string // GOBIN for this leaf ({WorkRoot}/gobin)
	Args       []string
	ExtraEnv   []string // additional KEY=VAL (GOBIN is always set by Run)

	// InProcess runs via wrkcli.Capture (L2 short path) instead of the product binary.
	// Prefer for help / mutual-exclusion leaves that do not need a process boundary.
	// Leave false (default) for true L3 e2e integration.
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
			Env:  reinstallLocalCLIEnv(req),
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
	cmd.Env = reinstallLocalCLIEnv(req)

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
