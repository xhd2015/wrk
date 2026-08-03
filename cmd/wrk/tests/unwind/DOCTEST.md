# wrk --unwind — stack DAG, dry-run plan + apply peel/pin + display/add-all fidelity

## Version
0.0.2

Decision tree for **stack unwind**:

- **Plan / cycle (baseline):** discover checkout **stack**, build module→**repo
  DAG** keyed by **normalized absolute MainRepo**, **reject cycles before any
  mutation**, print free-first peel order under `wrk --unwind --dry-run`. Flag
  validation (pin/land) before plan apply. `--reinstall-local` is an accepted
  tail request (not mutual-exclusion with `--unwind`).
- **Apply:** non-dry-run peels free-first with **explicit** ship/land flags;
  after shipping a dep that has stack consumers, **Pin** consumers to live tags
  and `go mod tidy`. Soft reinstall remains P1.
- **This cycle (Classic TDD):** peel **display path** = relative path of peel
  **checkout** vs invocation cwd (same policy as status `Dir:`); dry-run
  reflects `--add-all` / leave-N under gen-commit; apply stages only when
  `--add-all`; apply banner uses the same display path.

**Layer:** **L2** — in-process CLI via `wrkcli.Capture` (`req.InProcess=true`).
Fixture setup may use session-built `wrk` for `--new` / worktree materialization.
Prefer pure cores: path display helper (statusDirLine policy), leave-N porcelain
count, `PlanUnwind` / `FormatUnwindDryRun`, apply peel driver.

**Out of scope:** implying land flags; global `--propagate-tags` as unwind default;
non-unwind compose dry-run redesign.

# DSN (Domain Specific Notion)

- **wrk --unwind** — top-level primary mode. Compose with ship/land flags and
  `--dry-run`, and the optional **reinstall-local tail**. Mutually exclusive with unrelated modes (`--list`, `--status`,
  bare create, …). Event `command: "unwind"` when recorded (optional for leaves).
- **Stack** — status-partition style inventory of the checkout: **primary** git
  toplevel (main or linked) plus **nested external** independent repos under
  the tree (typically `external/*` dep worktrees). Discovered from cwd via
  git toplevel + nested scan (same shape as status external section).
- **Module → repo DAG** — scan Go modules per stack repo; collect
  `require` / `replace` edges among **stack-owned** modules; contract to a
  **repo DAG**. Identity keys are **normalized absolute MainRepo paths**
  (basename collisions must not merge distinct mains). Edge **`Rc → Rd`**
  means repo **Rc depends on** repo **Rd**.
- **Human short names** — pin / consumer short text still uses **basename** of
  MainRepo (e.g. `dot-pkgs`), not the peel display path.
- **Cycle preflight** — if the stack-repo DAG has a cycle among repos with
  edges → **exit non-zero**, message mentions **cycle**; **no mutations** and
  **no successful peel plan**. Cycle check runs **before any mutation** on
  both dry-run and apply.
- **Dirty pending (v1)** — only **dirty** stack repos enter the pending peel
  set (fixtures control dirtiness via untracked/modified files). Clean stack
  members are **skipped** from peel; consumers may still **receive Pin** when
  a dep they require is peeled. Apply fixtures may also leave **commits ahead**
  on linked WTs so land/tag have content after dirt is handled.
- **Peel order (free-first / Kahn)** — among **pending** dirty repos, peel
  nodes with **no dependency edge to other pending** repos first (leaves of
  the residual DAG). Example chain `root → agent-pro → dot-pkgs` (depends-on
  arrows): peel **leaf external then mid external then primary** (display paths
  below).
- **Peel display path** — human peel line / apply banner uses the **relative
  path of the peel checkout** vs invocation cwd, same policy as status
  `Dir:` (`statusDirLine`: slash form; abs fallback if Rel fails or too many
  `..`). Nested externals → `external/<name>-main-<date>`; primary at cwd →
  `.`. **Not** bare MainRepo basename alone as the full peel path.
- **Flag validation (before any mutation)**  
  - Cross-repo **pin** needed (pending graph has edges) → require
    **`--tag-next` + `--push`**.  
  - **Land** (`--merge-back` | `--done`) required when any pending node is a
    **linked worktree** (not already on main).  
  - **Already main:** no land required for that node.  
  - **`--unwind` alone** does **not** default any land flags.
- **Apply peel (per free dirty repo in order)** — explicit flags only:
  1. Optional pre: `--gen-commit-msg` / `--commit` when set; **`--add-all`
     only** stages (`git add -A`) before gen-commit — **no** unconditional
     always-`git add -A` when gen-commit is set without `--add-all`.
  2. Linked WT + `--merge-back`/`--done` → land as today; already main → skip land.
  3. `--sync` / `--tag-next` / `--push` / `--reinstall-local` as flags;
     reinstall **soft** (P1).
  4. After peeling **U** that has stack consumers: for each consumer **C**
     depending on U, for each module of U that C requires/replaces:
     `Pin(…)`; then `go mod tidy` in consumer module dirs.
  5. Prefer tags created by this peel's tag-next; else latest on main.
  6. `--done` removes worktree as usual.
  7. Fail-fast on hard errors (land/tag/push/pin/tidy); reinstall soft.
  8. Apply banner: `==== unwind: peel <display-path> ====` (same display as dry-run).
- **Dry-run stdout vocabulary (locked)**  
  - Banner: `==== unwind (dry-run) ====`.  
  - One peel line per pending repo in free-first order:
    `would: peel <display-path>`  
    (`external/…` or `.`, not bare label alone).  
  - Under a peel when `--gen-commit-msg` (+ `--commit` validation):
    - if **`--add-all`**: `  would: git add -A` then
      `  would: generate commit message and commit staged changes`
    - if **no `--add-all`** and peel has **N>0** not-fully-staged porcelain
      paths (unstaged + untracked):  
      `  would: leave N file(s) uncommitted (use --add-all if necessary)`  
      (singular `file` / plural `files`); then generate/commit line.  
    - if no `--add-all` and fully staged only: **no** leave-N line; still
      generate/commit plan language when gen-commit is set.
  - Optional ship lines under peel: merge-back / sync / tag / push / pin as flags.
  - Tail: `would: reinstall local binaries` when `--reinstall-local`.
  - Trailing newline on plan; exit **0**; **zero mutations**.
- **Apply asserts** — side effects over multi-stage stdout. Prefer: tags on
  leaf main / bare origin, consumer `go.mod` require version, worktree
  presence/absence, main HEAD advanced, zero mutation on cycle reject, banner
  display path, staging honor for `--add-all`.
- **Error surfaces**  
  - Cycle: combined stderr/stdout contains `cycle`; exit ≠ 0; no mutations.  
  - Missing pin flags when edges exist: names `--tag-next` and `--push`; exit ≠ 0.
- **Local remotes / offline tidy** — apply leaves attach **local bare** origins
  for `--push` (no network). Pin+tidy leaves may set `GOPROXY=file://…` via
  `req.ExtraEnv` and seed `{WorkRoot}/modproxy`.
- **WRK_HOME** — isolated per leaf at `{WorkRoot}/.wrk`.
- **WRK_DATE** — tests set `2026-06-30` for deterministic naming when creating
  worktrees.
- **Colors** — pipe harness → plain text (no ANSI required).
- **Implementer note** — Linked `--done` refuses dirty porcelain today:
  implementer may commit pending dirt before land, or expand pending/ship
  pre-stage; fixtures leave both **commits ahead** and a small **DIRTY**
  marker so either path is testable.

## Tree Overview

```
unwind/
├── dry-run/                              # acyclic plan (+ display / gen-commit staging)
│   ├── free-first-order/                 # 3-repo dirty; free-first with rel display paths
│   ├── single-repo-no-edges/             # main-only dirty → would: peel .
│   ├── clean-leaf-skipped/               # leaf clean; mid+primary display paths only
│   ├── missing-flags-with-edges/         # edges; no tag-next/push → error
│   ├── reinstall-local-tail/             # accepted tail; peel . + reinstall plan
│   └── gen-commit/                       # gen-commit dry-run add-all / leave-N
│       ├── add-all-reflected/            # --add-all → would: git add -A
│       ├── leave-uncommitted/            # no --add-all + unstaged/untracked → leave-N
│       └── leave-skipped-when-fully-staged/  # only staged → no leave-N line
├── apply/                                # non-dry-run peel / pin
│   ├── leaf-then-pin/                    # peel + pin; banner uses rel display path
│   ├── already-main-no-land/             # single main; tag-next+push; no land
│   └── done-removes-leaf-wt/             # --done peels leaf; external path gone
└── cycle/
    ├── two-cycle/                        # dry-run cycle reject
    └── apply-two-cycle/                  # apply-mode cycle still pre-mutation
```

Split factor (MECE, significance-first):

1. **Mode** — dry-run plan | apply mutation | cycle preflight.
2. Within dry-run: **order/skip/flags** | **gen-commit staging vocabulary**.
3. Within apply: **pin-after-peel (+ banner path)** | **already-main no land** | **done removes WT**.
4. Cycle: dry-run vs apply-mode (same reject, no mutations).

## Test Case Index

| # | Leaf | Description | Expect |
|---|------|-------------|--------|
| D1 | dry-run/free-first-order | 3-repo dirty; peel order external leaf → external mid → `.`; zero mutations | **RED** until rel display |
| D2 | dry-run/single-repo-no-edges | sole main dirty; `would: peel .`; no pin flags | **RED** until rel display (was bare `root`) |
| D3 | dry-run/clean-leaf-skipped | leaf clean; peel mid external then `.` only | **RED** until rel display |
| D4 | dry-run/missing-flags-with-edges | edges + dry-run without tag/push → error | **GREEN** (error path; no peel display lock) |
| D5 | dry-run/reinstall-local-tail | dirty main + `--reinstall-local`; peel `.`; no mutual-exclusion; zero mutation | **RED** until rel display |
| D6 | dry-run/gen-commit/add-all-reflected | gen-commit + `--add-all` + dry-run → `would: git add -A` then generate | **RED** until add-all plan |
| D7 | dry-run/gen-commit/leave-uncommitted | gen-commit, no add-all, untracked → leave-N line; exit 0; zero mutations | **RED** until leave-N |
| D8 | dry-run/gen-commit/leave-skipped-when-fully-staged | gen-commit, no add-all, only staged → no leave-N | **RED** until leave-N gate |
| C1 | cycle/two-cycle | A↔B dry-run → cycle error; zero mutations | **GREEN** |
| A1 | apply/leaf-then-pin | peel leaf + pin; stdout banner `==== unwind: peel external/… ====`; pin require | **RED** until banner display path (side effects already land) |
| A2 | apply/already-main-no-land | single main dirty + tag-next+push; no land | **GREEN** (no display path assert) |
| A3 | apply/done-removes-leaf-wt | `--done` peels leaf; external path gone; pin on root | **GREEN** (path absence, not banner string) |
| C2 | cycle/apply-two-cycle | A↔B apply-mode; cycle error before mutation | **GREEN** |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/unwind
doctest test -count=1 ./cmd/wrk/tests/unwind
doctest test -count=1 ./cmd/wrk/tests/unwind/dry-run/free-first-order
doctest test -count=1 ./cmd/wrk/tests/unwind/dry-run/gen-commit/add-all-reflected
doctest test -count=1 ./cmd/wrk/tests/unwind/apply/leaf-then-pin
doctest test -count=1 ./cmd/wrk/tests/unwind/cycle/apply-two-cycle
```

Expected **RED** until implementer lands rel peel display, gen-commit dry-run
add-all/leave-N, and apply banner display path. Error/cycle leaves and side-effect-only
apply leaves stay **GREEN**.

```go
import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot string
	WrkHome  string
	RepoDir  string // process cwd for wrk --unwind
	Args     []string

	// InProcess runs via wrkcli.Capture (L2). Prefer true for all leaves.
	InProcess bool

	// ExtraEnv is appended to Capture Env (e.g. GOPROXY=file://… for tidy).
	ExtraEnv []string

	// Stack fixture paths (filled by helpers).
	MainRepo        string // root consumer main
	WtDir           string // root consumer linked worktree (when used)
	WtBranch        string
	DepPath         string // agent-pro main (or cycle A main)
	SecondRepo      string // dot-pkgs / leaf main (or cycle B main)
	ExternalWtDir   string // agent-pro external under stack (or cycle A ext)
	DepsLinkedWtDir string // leaf (dot-pkgs) external under stack (or cycle B ext)

	// PeelOrder is the expected free-first peel *display path* sequence for dry-run
	// success leaves (statusDirLine policy: "external/…", ".", not bare MainRepo basenames).
	PeelOrder []string
	// LeaveN is expected not-fully-staged path count for leave-uncommitted dry-run leaves.
	LeaveN int

	// Apply fixture extras (P4).
	OriginBare         string // bare remote for leaf or single-main push
	ExpectedPinVersion string // e.g. v0.0.2 after peel tag-next
	OldRequireVersion  string // e.g. v0.0.1 pre-pin require
	LeafModulePath     string // e.g. example.com/dot-pkgs
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
			Dir:  req.RepoDir,
			Env:  unwindWrkEnv(req),
		})
		return &Response{
			Stdout:   res.Stdout,
			Stderr:   res.Stderr,
			ExitCode: res.ExitCode,
		}, nil
	}

	bin := getWrkBin(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = unwindWrkEnv(req)

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
