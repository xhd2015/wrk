# wrk --unwind — stack DAG, dry-run plan (P3 GREEN) + apply peel/pin (P4 TDD)

## Version
0.0.2

Decision tree for **stack unwind**:

- **P3 (GREEN):** discover checkout **stack**, build module→**repo DAG**,
  **reject cycles before any mutation**, print free-first peel order under
  `wrk --unwind --dry-run`. Flag validation (pin/land) before plan apply.
- **P4 (Classic TDD RED for new apply leaves):** non-dry-run **apply** peels
  free-first with **explicit** ship/land flags only; after shipping a dep that
  has stack consumers, **Pin** consumers to live tags (`gotool/update.Pin`) and
  `go mod tidy` in consumer module dirs. Soft reinstall remains P1.

**Layer:** **L2** — in-process CLI via `wrkcli.Capture` (`req.InProcess=true`).
Fixture setup may use session-built `wrk` for `--new` / worktree materialization.
Prefer pure cores: `BuildStackInventory`, `BuildRepoDAG`, `PeelOrder` / cycle
detect, `PlanUnwind`, apply peel driver calling existing land/tag/push + Pin.

**P3 contracts stay GREEN/unchanged.** New apply leaves may stay **RED** until
implementer lands non-dry-run apply (today stubs with
`wrk: --unwind apply is not implemented`).

**Out of scope:** implying land flags; global `--propagate-tags` as unwind default.

# DSN (Domain Specific Notion)

- **wrk --unwind** — top-level primary mode. Compose with ship/land flags and
  `--dry-run`. Mutually exclusive with unrelated modes (`--list`, `--status`,
  bare create, …). Event `command: "unwind"` when recorded (optional for leaves).
- **Stack** — status-partition style inventory of the checkout: **primary** git
  toplevel (main or linked) plus **nested external** independent repos under
  the tree (typically `external/*` dep worktrees). Discovered from cwd via
  git toplevel + nested scan (same shape as status external section).
- **Module → repo DAG** — scan Go modules per stack repo; collect
  `require` / `replace` edges among **stack-owned** modules; contract to a
  **repo DAG**. Edge **`Rc → Rd`** means repo **Rc depends on** repo **Rd**.
- **Cycle preflight** — if the stack-repo DAG has a cycle among repos with
  edges → **exit non-zero**, message mentions **cycle**; **no mutations** and
  **no successful peel plan**. Cycle check runs **before any mutation** on
  both dry-run and apply (including non-dry-run apply leaves).
- **Dirty pending (v1)** — only **dirty** stack repos enter the pending peel
  set (fixtures control dirtiness via untracked/modified files). Clean stack
  members are **skipped** from peel; consumers may still **receive Pin** when
  a dep they require is peeled. Apply fixtures may also leave **commits ahead**
  on linked WTs so land/tag have content after dirt is handled.
- **Peel order (free-first / Kahn)** — among **pending** dirty repos, peel
  nodes with **no dependency edge to other pending** repos first (leaves of
  the residual DAG). Example chain `root → agent-pro → dot-pkgs` (depends-on
  arrows): peel **`dot-pkgs` then `agent-pro` then `root`**.
- **Flag validation (before any mutation)**  
  - Cross-repo **pin** needed (pending graph has edges) → require
    **`--tag-next` + `--push`**.  
  - **Land** (`--merge-back` | `--done`) required when any pending node is a
    **linked worktree** (not already on main).  
  - **Already main:** no land required for that node.  
  - **`--unwind` alone** does **not** default any land flags.
- **Apply peel (P4, per free dirty repo in order)** — explicit flags only:
  1. Optional pre: `--add-all` / `--gen-commit-msg` / `--commit` when set
     (tests prefer committed ahead + small porcelain dirt, or pin-only dirt).
  2. Linked WT + `--merge-back`/`--done` → land as today; already main → skip land.
  3. `--sync` / `--tag-next` / `--push` / `--reinstall-local` as flags;
     reinstall **soft** (P1).
  4. After peeling **U** that has stack consumers: for each consumer **C**
     depending on U, for each module of U that C requires/replaces:
     `Pin(ConsumerDir=C module, DepDir=U main module dir, Version from new tags
     if known)`; then `go mod tidy` in consumer module dirs.
  5. Prefer tags created by this peel's tag-next; else latest on main.
  6. `--done` removes worktree as usual.
  7. Fail-fast on hard errors (land/tag/push/pin/tidy); reinstall soft.
  8. Non-dry-run replaces P3 stub `"not implemented"`.
- **Dry-run stdout vocabulary (P3 locked)**  
  - Banner containing `unwind` and `dry-run` (e.g. `==== unwind (dry-run) ====`).  
  - One peel line per pending repo in free-first order:
    `would: peel <label>`
    where `<label>` is the **stable stack label** = basename of the **main
    repo directory** (`dot-pkgs`, `agent-pro`, `root`, …).  
  - Successful dry-run exit **0**; **zero mutations**.
- **Apply asserts (P4)** — side effects over exact multi-stage stdout (stdout
  shape implementer-owned unless a leaf locks a substring). Prefer:
  tags on leaf main / bare origin, consumer `go.mod` require version, worktree
  presence/absence, main HEAD advanced, zero mutation on cycle reject.
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
- **Implementer note** — wrk `go.mod` pins `go-pkgs v0.0.94`; `update.Pin` lives
  in brought `external/dot-pkgs-master-2026-07-31/go-pkgs` — may need
  `replace` to the local path. Designer does **not** add that replace.
  Linked `--done` refuses dirty porcelain today: implementer may commit pending
  dirt before land, or expand pending/ship pre-stage; fixtures leave both
  **commits ahead** and a small **DIRTY** marker so either path is testable.

## Tree Overview

```
unwind/
├── dry-run/                              # P3 acyclic plan (GREEN — do not break)
│   ├── free-first-order/                 # 3-repo chain all dirty + pin/land flags
│   ├── single-repo-no-edges/             # main-only dirty; no tag/push required
│   ├── clean-leaf-skipped/               # leaf clean; mid+root dirty → peel mid, root
│   └── missing-flags-with-edges/         # edges present; no tag-next/push → error
├── apply/                                # P4 non-dry-run peel / pin (Classic TDD)
│   ├── leaf-then-pin/                    # peel leaf WT → pin root require to new tag
│   ├── already-main-no-land/             # single main dirty; tag-next+push; no land
│   └── done-removes-leaf-wt/             # --done peels leaf; external path gone
└── cycle/
    ├── two-cycle/                        # P3 dry-run cycle reject (GREEN)
    └── apply-two-cycle/                  # P4 apply-mode cycle still pre-mutation
```

Split factor (MECE, significance-first):

1. **Mode** — dry-run plan (P3) | apply mutation (P4) | cycle preflight (both).
2. Within apply: **pin-after-peel** | **already-main no land** | **done removes WT**.
3. Cycle: dry-run vs apply-mode (same reject, no mutations).

## Test Case Index

| # | Leaf | Description | Expect |
|---|------|-------------|--------|
| D1 | dry-run/free-first-order | 3-repo dirty chain; dry-run peel order leaf→mid→root; zero mutations | GREEN |
| D2 | dry-run/single-repo-no-edges | sole root main dirty; dry-run single peel; no pin flags | GREEN |
| D3 | dry-run/clean-leaf-skipped | leaf clean; peel mid then root only | GREEN |
| D4 | dry-run/missing-flags-with-edges | edges + dry-run without tag/push → error | GREEN |
| C1 | cycle/two-cycle | A↔B dry-run → cycle error; zero mutations | GREEN |
| A1 | apply/leaf-then-pin | 2-repo leaf WT dirty+ahead + root requires leaf; `--unwind --done --tag-next --push`; pin root to new tag; leaf main advanced; local bare push | RED until apply |
| A2 | apply/already-main-no-land | single main dirty + root-bump + bare origin; `--unwind --tag-next --push` (no land); tag+push; no worktree land | RED until apply |
| A3 | apply/done-removes-leaf-wt | same 2-repo apply stack; `--done` peels leaf; `external/…` leaf WT gone; pin still on root main go.mod | RED until apply |
| C2 | cycle/apply-two-cycle | same A↔B stack; **no** `--dry-run`; cycle error before any mutation | GREEN (preflight already runs) |

## How to Run

```sh
doctest vet ./cmd/wrk/tests/unwind
doctest test -count=1 ./cmd/wrk/tests/unwind
doctest test -count=1 ./cmd/wrk/tests/unwind/dry-run/free-first-order
doctest test -count=1 ./cmd/wrk/tests/unwind/apply/leaf-then-pin
doctest test -count=1 ./cmd/wrk/tests/unwind/cycle/apply-two-cycle
```

P3 leaves remain GREEN. New apply leaves (A1–A3) expect **RED** until
implementer lands non-dry-run apply + pin-after-peel. C2 should stay GREEN
(cycle aborts before the apply stub).

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

	// PeelOrder is the expected free-first label sequence for dry-run success leaves.
	PeelOrder []string

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
