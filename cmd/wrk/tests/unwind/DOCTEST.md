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
- **Pin path (prior cycle):** **pin consumers at in-scope StackMember.Path**
  (primary checkout + nested external repos under it — status scan), **not**
  remapped to `MainRepo` when Path is a linked/nested checkout whose MainRepo
  lies outside the current scope.
- **Prior cycle (still open RED on apply):** multi-module pin selectivity;
  tidy go-stderr surfacing; pin Path on linked consumer.
- **This cycle (Classic TDD):** **follow local filesystem replaces** into the
  stack inventory via full **BFS fixpoint** over all Go modules under known
  checkouts. Resolve `NewPath` relative to the **module directory**, map to
  git checkout via **ShowToplevel**, skip intra-repo, emit `warning:` for
  missing/non-git targets, add synthetic DAG edge **C→D** when follow adds D
  from C. Dirty-only peel (v1) unchanged. Dry-run leaves under
  `dry-run/follow-local-replace/` are the primary RED drivers.

**Layer:** **L2** — in-process CLI via `wrkcli.Capture` (`req.InProcess=true`).
Fixture setup may use session-built `wrk` for `--new` / worktree materialization.
Prefer pure cores: path display helper (statusDirLine policy), leave-N porcelain
count, `PlanUnwind` / `FormatUnwindDryRun`, apply peel driver, inventory expand.

**Out of scope this cycle:** `--status` parity; opt-out flag; walking parent
`external/` without a replace; module-path-only replaces as discovery seeds;
implying land flags; global `--propagate-tags` as unwind default.

# DSN (Domain Specific Notion)

- **wrk --unwind** — top-level primary mode. Compose with ship/land flags and
  `--dry-run`, and the optional **reinstall-local tail**. Mutually exclusive with unrelated modes (`--list`, `--status`,
  bare create, …). Event `command: "unwind"` when recorded (optional for leaves).
- **Stack** — inventory of checkouts for unwind:
  1. **Seed:** **primary** git toplevel (main or linked) plus **nested
     external** independent repos under that root (status-like
     `discoverStatusRepos` — typically in-tree `external/*`).
  2. **Expand (this cycle):** full **BFS fixpoint** over **local filesystem
     replaces** on **all Go modules** (including nested go.mod) under every
     known stack checkout:
     - Local replace = `NewPath` is `./`, `../`, or absolute **and** no
       version (same semantics as `isLocalFilesystemReplace` /
       `Module.LocalFilesystemReplaces()`).
     - Resolve `NewPath` relative to the **module directory** (not always
       repo root).
     - Map target → git checkout via **`ShowToplevel` of resolved path**
       (replace into `…/go-pkgs` or `…/nested` becomes the dep **repo root**).
     - **Intra-repo** (toplevel already the owner / same git root already in
       stack): **never** add a separate stack member.
     - **Extra-repo:** add checkout to stack; scan it next (transitive).
     - **Missing / non-git targets:** emit **`warning:`** on **stderr**, skip
       that target; do not fail unwind solely for that (exit 0 if plan
       otherwise OK).
- **Module → repo DAG** — scan Go modules per stack repo; collect
  `require` / `replace` edges among **stack-owned** modules; contract to a
  **repo DAG**. Identity keys are **normalized absolute MainRepo paths**
  (basename collisions must not merge distinct mains). Edge **`Rc → Rd`**
  means repo **Rc depends on** repo **Rd**.
- **Synthetic DAG edges (option B)** — when expansion adds dep checkout **D**
  because consumer checkout **C** has a local filesystem replace resolving into
  **D**, always add edge **C → D** (C depends on D), **deduped** with edges
  from existing require/OldPath contraction in `BuildRepoDAG`. Synthetic edges
  count for pin-flag validation and free-first residual graph.
- **Human short names** — pin / consumer short text still uses **basename** of
  MainRepo (e.g. `dot-pkgs`), not the peel display path.
- **Cycle preflight** — if the stack-repo DAG has a cycle among repos with
  edges → **exit non-zero**, message mentions **cycle**; **no mutations** and
  **no successful peel plan**. Cycle check runs **before any mutation** on
  both dry-run and apply.
- **Dirty pending (v1)** — only **dirty** stack repos enter the pending peel
  set (fixtures control dirtiness via untracked/modified files). Clean stack
  members (including **clean followed** checkouts) stay in inventory for
  DAG/cycle but produce **no peel line**. Consumers may still **receive Pin**
  when a dep they require is peeled. Apply fixtures may also leave **commits
  ahead** on linked WTs so land/tag have content after dirt is handled.
- **Peel order (free-first / Kahn)** — among **pending** dirty repos, peel
  nodes with **no dependency edge to other pending** repos first (leaves of
  the residual DAG, including synthetic edges). Example nested chain
  `root → agent-pro → dot-pkgs`: peel **leaf external then mid external then
  primary**. Example follow chain A→B→C via local replaces: peel **C then B
  then A**.
- **Peel display path** — human peel line / apply banner uses the **relative
  path of the peel checkout** vs invocation cwd, same policy as status
  `Dir:` (`statusDirLine`: slash form; abs fallback if Rel fails or too many
  `..`). Nested under primary → `external/<name>-main-<date>`; sibling
  outside primary → often `../external/<name>-…`; primary at cwd → `.`.
  When replace targets a nested module subdir, display still uses **dep git
  toplevel** Path. **Not** bare MainRepo basename alone as the full peel path.
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
     depending on U, for each module of U that **C requires or replaces**
     (module-path match — **not** a Cartesian product of every dep module dir
     into every consumer module dir):
     `Pin(…)` at **C's in-scope StackMember.Path** (the checkout discovered
     in the current stack inventory — primary ShowToplevel and nested
     externals). **Do not** rewrite the pin target to `C.MainRepo` when
     Path is a linked/nested checkout and MainRepo is outside this scope.
     When the scope primary *is* main, Path is main (in scope) and pin
     **does** edit that path. **Do not** force-add a nested dep module path
     the consumer never required/replaced. Then `go mod tidy` in those
     consumer module dirs.
  5. Prefer **per-module** versions from this peel's tag-next created tags
     when available; do not invent a nested version from the root tag when
     that nested scope was not tagged this peel (prefer skip / leave that
     require alone). Else latest on main for modules that are actually pinned.
     (Dep land/tag/push still use peeled dep **main** after merge — not changed
     by this pin-path / pin-selectivity rule.)
  6. `--done` removes worktree as usual.
  7. Fail-fast on hard errors (land/tag/push/pin/tidy); reinstall soft.
     On tidy failure: error must include **trimmed go child stderr** (concrete
     diagnostic such as `unknown revision` / missing proxy file / module path),
     not only `exit status 1`. Quiet success without `-v` (no spam on OK tidy).
  8. Apply banner: `==== unwind: peel <display-path> ====` (same display as dry-run).
  9. Pin log: basename short form; one line per **real** pin (no cartesian spam
     for non-matching modules).
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
  leaf main / bare origin, consumer **Path** `go.mod` require version (and
  separate consumer **MainRepo** `go.mod` baseline when Path ≠ MainRepo),
  worktree presence/absence, main HEAD advanced, zero mutation on cycle reject,
  banner display path, staging honor for `--add-all`.
- **Error surfaces**  
  - Cycle: combined stderr/stdout contains `cycle`; exit ≠ 0; no mutations.  
  - Missing pin flags when edges exist: names `--tag-next` and `--push`; exit ≠ 0.
  - Tidy fail: mentions `go mod tidy` (and consumer path); **includes go child
    diagnostic body** (not only `exit status 1`).
  - Missing / non-git local-replace target: **`warning:`** on stderr; plan
    continues; exit 0 when otherwise OK (not a hard fail).
- **Local remotes / offline tidy** — apply leaves attach **local bare** origins
  for `--push` (no network). Pin+tidy leaves may set `GOPROXY=file://…` via
  `req.ExtraEnv` and seed `{WorkRoot}/modproxy`. Multi-module pin leaves seed
  only modules/versions that should resolve; nested-not-required fixtures omit
  nested next-tag proxy entries so a spurious nested pin fails tidy today.
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
├── dry-run/                              # acyclic plan (+ display / gen-commit / follow)
│   ├── free-first-order/                 # F4 regression: nested external free-first
│   ├── single-repo-no-edges/             # main-only dirty → would: peel .
│   ├── clean-leaf-skipped/               # leaf clean; mid+primary display paths only
│   ├── missing-flags-with-edges/         # edges; no tag-next/push → error
│   ├── reinstall-local-tail/             # accepted tail; peel . + reinstall plan
│   ├── gen-commit/                       # gen-commit dry-run add-all / leave-N
│   │   ├── add-all-reflected/            # --add-all → would: git add -A
│   │   ├── leave-uncommitted/            # no --add-all + unstaged/untracked → leave-N
│   │   └── leave-skipped-when-fully-staged/  # only staged → no leave-N line
│   └── follow-local-replace/             # BFS local-filesystem replace expansion
│       ├── sibling-both-dirty/           # F1: sibling ../external; both dirty
│       ├── nested-module-owns-replace/   # F2: nested go.mod owns replace
│       ├── intra-repo-only/              # F3: ./pkgs/shared → only peel .
│       ├── clean-dep-skipped/            # F5: clean followed dep omitted from peel
│       ├── transitive-chain/             # F6: A→B→C free-first C,B,A
│       ├── missing-target-warns/         # F7: warning: + continue peel .
│       └── nested-mod-target-toplevel/   # F8: replace nested subdir → toplevel peel
├── apply/                                # non-dry-run peel / pin
│   ├── leaf-then-pin/                    # pin when primary Path == main (in scope)
│   ├── pin-on-linked-consumer-not-main/  # pin Path=linked WT; MainRepo go.mod untouched
│   ├── multi-module-pin-require-root-only/ # multi-mod dep; consumer root-only → no nested pin
│   ├── tidy-error-surfaces-go-stderr/    # tidy fail must include go child diagnostic
│   ├── already-main-no-land/             # single main; tag-next+push; no land
│   └── done-removes-leaf-wt/             # --done peels leaf; external path gone
└── cycle/
    ├── two-cycle/                        # dry-run cycle reject
    └── apply-two-cycle/                  # apply-mode cycle still pre-mutation
```

Split factor (MECE, significance-first):

1. **Mode** — dry-run plan | apply mutation | cycle preflight.
2. Within dry-run: **order/skip/flags** | **gen-commit staging vocabulary** |
   **follow-local-replace inventory expansion**.
3. Within apply: **pin target Path** | **pin module selectivity (multi-mod)** |
   **tidy error surfacing** | **already-main no land** | **done removes WT**.
4. Cycle: dry-run vs apply-mode (same reject, no mutations).

## Test Case Index

| # | Leaf | Description | Expect |
|---|------|-------------|--------|
| D1 | dry-run/free-first-order | **F4** nested 3-repo dirty; free-first external leaf → mid → `.` | **GREEN** (regression) |
| D2 | dry-run/single-repo-no-edges | sole main dirty; `would: peel .`; no pin flags | **GREEN** |
| D3 | dry-run/clean-leaf-skipped | leaf clean; peel mid external then `.` only | **GREEN** |
| D4 | dry-run/missing-flags-with-edges | edges + dry-run without tag/push → error | **GREEN** |
| D5 | dry-run/reinstall-local-tail | dirty main + `--reinstall-local`; peel `.`; no mutual-exclusion; zero mutation | **GREEN** |
| D6 | dry-run/gen-commit/add-all-reflected | gen-commit + `--add-all` + dry-run → `would: git add -A` then generate | **GREEN** |
| D7 | dry-run/gen-commit/leave-uncommitted | gen-commit, no add-all, untracked → leave-N line; exit 0; zero mutations | **GREEN** |
| D8 | dry-run/gen-commit/leave-skipped-when-fully-staged | gen-commit, no add-all, only staged → no leave-N | **GREEN** |
| F1 | dry-run/follow-local-replace/sibling-both-dirty | sibling `../external/…` replace; both dirty; peel dep then `.` | **RED** until follow lands |
| F2 | dry-run/follow-local-replace/nested-module-owns-replace | nested module owns replace; dep still in plan | **RED** until nested-module scan follow |
| F3 | dry-run/follow-local-replace/intra-repo-only | `replace => ./pkgs/shared`; only `would: peel .` | **GREEN** regression (no extra member) |
| F5 | dry-run/follow-local-replace/clean-dep-skipped | out-of-tree dep clean; only peel `.` | **GREEN** today; stays GREEN after follow |
| F6 | dry-run/follow-local-replace/transitive-chain | A→B→C local replaces; peel C then B then A | **RED** until BFS follow + synthetic edges |
| F7 | dry-run/follow-local-replace/missing-target-warns | missing replace target → `warning:`; peel `.`; exit 0 | **RED** until warning emitted |
| F8 | dry-run/follow-local-replace/nested-mod-target-toplevel | replace nested subdir → peel dep **git toplevel** display | **RED** until ShowToplevel map |
| C1 | cycle/two-cycle | A↔B dry-run → cycle error; zero mutations | **GREEN** |
| A1 | apply/leaf-then-pin | primary is main (Path == MainRepo); peel leaf + pin **that** Path go.mod; banner rel display | **GREEN** (pin-when-primary-is-main) |
| A4 | apply/pin-on-linked-consumer-not-main | primary is **linked** consumer WT; pin WT go.mod; consumer **MainRepo** go.mod baseline unchanged | **RED** until pin uses Path not MainRepo |
| A5 | apply/multi-module-pin-require-root-only | multi-mod dep (root+nested); consumer requires **only root**; peel tags root next; **must not** force-add nested require; tidy OK | **RED** until pin matches require/replace only |
| A6 | apply/tidy-error-surfaces-go-stderr | pin then tidy fails (next version missing from modproxy); stderr includes **go child diagnostic**, not only exit status 1 | **RED** until goModTidy captures child output |
| A2 | apply/already-main-no-land | single main dirty + tag-next+push; no land | **GREEN** |
| A3 | apply/done-removes-leaf-wt | `--done` peels leaf; external path gone; pin on root Path (main) | **GREEN** |
| C2 | cycle/apply-two-cycle | A↔B apply-mode; cycle error before mutation | **GREEN** |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/unwind
doctest test -count=1 ./cmd/wrk/tests/unwind
doctest test -count=1 ./cmd/wrk/tests/unwind/dry-run/follow-local-replace
doctest test -count=1 ./cmd/wrk/tests/unwind/dry-run/free-first-order
doctest test -count=1 ./cmd/wrk/tests/unwind/apply/multi-module-pin-require-root-only
doctest test -count=1 ./cmd/wrk/tests/unwind/apply/tidy-error-surfaces-go-stderr
doctest test -count=1 ./cmd/wrk/tests/unwind/apply/pin-on-linked-consumer-not-main
doctest test -count=1 ./cmd/wrk/tests/unwind/apply/leaf-then-pin
```

Expected **RED** this cycle on **F1, F2, F6, F7, F8** (follow expansion / warning /
toplevel map) until implementer lands. **F3, F5** and **D1/F4** free-first
regression stay **GREEN** (or report unexpected RED). Prior apply RED (**A4–A6**)
unchanged. Other prior leaves stay **GREEN**.

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
	LeafModulePath     string // e.g. example.com/dot-pkgs or example.com/dep
	// NestedModulePath is the nested multi-module dep path (e.g. example.com/dep/nested).
	// Empty for single-module fixtures. A5 asserts this must NOT appear in consumer go.mod.
	NestedModulePath string
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
