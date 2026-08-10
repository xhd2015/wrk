# wrk --unwind — stack DAG, dry-run plan + apply peel/pin + show-graph + verify + display fidelity

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
- **Follow local-replace (prior cycle):** BFS fixpoint over local filesystem
  replaces into stack inventory; synthetic edges; missing-target `warning:`.
- **Show-graph (prior cycle):** read-only `wrk --unwind --show-graph [--json]
  [--color|--no-color]`. Repo DAG + dirty free-first peel + full stack module
  DAG + summary. Human polish + reject/cycle/json under `show-graph/`. **Do not
  rewrite sealed show-graph assert meaning** in this cycle.
- **P1 (sealed):** **global free-module cascade plan** under
  `wrk --unwind --dry-run --tag-next` — after existing **repo peel** plan lines,
  free-first **module** cascade `would: tag-next` / `would: pin` from
  `PlanUnwindCascade`. Dry-run only (zero mutations). Leaves under
  `cascade/dry-run/` — do not rewrite sealed asserts unjustifiably.
- **P2 (sealed GREEN):** **apply cascade core** (clean tree or `--add-all`) under
  `wrk --unwind --tag-next` (not dry-run). After free-first **repo land prelude**,
  global free-module cascade: tag **one scope** → pin consumers (**keep local
  replace**; bump require) → selective commit when pin dirties → drop free; push
  when a main has no remaining pending modules. **No TagNextAll on this path.**
  Leaves under `cascade/apply/clean/` and `dirty-gomod/with-add-all/` sealed.
- **P3 (partial edit):** when cascade must pin, **no `--add-all`**, and WT
  go.mod/go.sum differ from Base (index else HEAD): save WIP → write Base → pin
  + `go mod tidy` on Base → selective cascade commit → restore WIP + **surgical
  require version bumps only** (no tidy on WT). On failure: restore WIP from
  save, non-zero. **C-AP5** flipped from P2 hard Error to partial-edit
  **success** (product intent D11). Leaves under `cascade/apply/partial-edit/`.
- **P4 reinstall-local (prior):** nested skip pin + reinstall under
  `cascade/apply/reinstall-local/`. Dry-run reinstall vocabulary sealed **C-DR6**.
- **This cycle (Classic TDD — verify):** read-only post-job audit
  `wrk --unwind --verify [--json] [--color|--no-color]`. Six error-severity
  checks (dirty-peel, needs-land, owned-changed, require-drift,
  droppable-replace, cascade-pending); exit **1** on any FAIL with report on
  stdout (no `Error:` for logical FAIL); exit **0** when all pass; soft
  `warning:` allowed. Leaves under `verify/`. **Do not** rewrite sealed
  show-graph / cascade asserts. No product implementation yet → verify leaves
  **RED**.

**Layer:** **L2** — in-process CLI via `wrkcli.Capture` (`req.InProcess=true`).
Fixture setup may use session-built `wrk` for `--new` / worktree materialization.
Prefer pure cores: path display helper (statusDirLine policy), leave-N porcelain
count, `PlanUnwind` / `FormatUnwindDryRun`, **`PlanUnwindCascade`** + dry-run
printer (P1), **apply cascade driver** (P2), **partial-edit pin path** (P3),
**reinstall-local tail after cascade** (P4), **BuildUnwindGraphReport** /
**FormatUnwindGraphHuman** / **FormatUnwindGraphJSON**, inventory expand,
**BuildUnwindVerifyReport** / human+JSON verify formatters (this cycle).

**Out of scope this cycle (verify):** remote push checks; projects.json /
propagate-tags; worktree-removed / tag-points-at-main-HEAD; rewriting sealed
show-graph/cascade asserts; L3 e2e.

# DSN (Domain Specific Notion)

- **wrk --unwind** — top-level primary mode. Compose with ship/land flags and
  `--dry-run`, optional **reinstall-local tail**, or read-only **`--show-graph
  [--json]`** / **`--verify [--json]`**. Mutually exclusive with unrelated modes
  (`--list`, `--status`, bare create, …). Event `command: "unwind"` when recorded
  (optional for leaves).
- **Verify (read-only post-job audit)** — `wrk --unwind --verify [--json]
  [--color|--no-color]`. Same stack inventory / edges as show-graph + cascade;
  emits six error-severity checks and status summary. Exit **1** if any check
  FAILs (report on stdout; no `Error:` for logical FAIL); exit **0** when all
  pass. Soft inventory `warning:` on stderr. Rejects dry-run/apply partners and
  `--show-graph`; bare `--verify` requires `--unwind`. Never mutates.
- **Show-graph (read-only)** — `wrk --unwind --show-graph [--json]
  [--color|--no-color]`. Collect stack inventory + repo DAG + PlanUnwind peel
  order; scan full stack modules; print human graph (dir identity, collapsed
  `→` edges, `replaced`, optional drift) or stable JSON. Rejects `--dry-run`
  and apply partners; does **not** run ValidateUnwindFlags / ApplyUnwind; never
  mutates.
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
  4. **Legacy peel pin (pre-cascade / non-tag-next paths):** after peeling **U**
     that has stack consumers: for each consumer **C** depending on U, for each
     module of U that **C requires or replaces** (module-path match — **not**
     Cartesian): `Pin(…)` at **C's in-scope Path**; tidy. Prefer per-module
     versions from this peel's created tags.
  5. **Global free-module cascade apply (P2, when `--tag-next`):** after free-first
     **land prelude** for dirty linked repos, run `PlanUnwindCascade` order:
     - Tag **one scope** only (not TagNextAll whole main).
     - Pin each stack consumer that requires/replaces it: bump require; **keep
       existing local replace** (do not drop replace).
     - `go mod tidy` on consumer module dir (on Base for clean / `--add-all`).
     - If pin dirtied go.mod/go.sum: selective commit
       `wrk: cascade pin <mod> @ <ver>` (stage those files; with `--add-all` may
       add more).
     - Drop module from pending; push each main when it has no remaining pending
       cascade modules (if `--push`).
     - **Commit-before-tag:** a module free only after receiving pins is tagged
       only after its cascade pin commits are on history.
     - **Dirty go.mod/go.sum without `--add-all` (P3 partial edit):** save WT
       go.mod/go.sum → write Base → pin+tidy on Base → selective commit
       (Base+pin+tidy, no WIP) → restore WT + surgical require bumps only. On
       failure restore save and exit non-zero. Ordinary WIP is **not** hard Error.
  6. `--done` removes worktree as usual (land prelude).
  7. Fail-fast on hard errors (land/tag/push/pin/tidy); reinstall soft.
     On tidy failure: error must include **trimmed go child stderr** (concrete
     diagnostic such as `unknown revision` / missing proxy file / module path),
     not only `exit status 1`. Quiet success without `-v` (no spam on OK tidy).
  8. Apply banner: `==== unwind: peel <display-path> ====` (same display as dry-run)
     for land/peel prelude when printed.
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
      (singular `file` / plural `files`); then generate/commit plan language.  
    - if no `--add-all` and fully staged only: **no** leave-N line; still
      generate/commit plan language when gen-commit is set.
  - Optional ship lines under peel: merge-back / sync / tag / push / pin as flags
    (legacy under-peel vocabulary may still print `  would: create release tag`
    / `  would: pin stack consumers` — **distinct** from cascade module lines).
  - **Global free-module cascade (P1, only when `--tag-next`)** — after **all**
    peel lines (and their under-peel ship lines), emit free-first cascade steps
    from the **stack module DAG** (same inventory / edges as show-graph; Kahn
    free-first on module edges where **From depends on To**):
    - Pending modules = owned-changed ∪ modules that require a pending stack
      module ∪ modules with require-drift vs stack latest.
    - Exclude tagscope forever-skip / **testdata** scopes from **`would: tag-next`**
      lines (scan already prunes `testdata/` go.mod; tagscope path scopes under
      `testdata/` must also never appear as cascade tags).
    - Tag line (top-level, no indent):  
      `would: tag-next <module-path> @ <next-tag>`  
      (`module-path` = Go module path; `next-tag` = tagscope planned full tag,
      e.g. `v0.0.2` or `pkgs/shared/v0.0.2`).
    - Pin line (top-level):  
      `would: pin <consumer-module-path> <- <dep-module-path> @ <ver>`  
      when a consumer must pin a dep that was / would be tagged earlier in the
      cascade (keep-local-replace is apply-time; dry-run still shows pin).
    - Free-first order: **leaf/dep module before consumer module** across the
      **global** graph (not per-repo cascade sections).
    - Without `--tag-next`: **no** cascade `would: tag-next` / cascade pin lines
      (`would: pin … <- … @ …`). Status-quo peel-only (+ reinstall tail).
    - Cycle reject: still `cycle` error; **no** successful cascade body
      (no ordered multi-step `would: tag-next` plan).
  - Tail: `would: reinstall local binaries` when `--reinstall-local` (after peels
    and cascade when both present).
  - Trailing newline on plan; exit **0**; **zero mutations** on dry-run.
- **Apply asserts** — side effects over multi-stage stdout. Prefer: tags on
  leaf main / bare origin, consumer **Path** `go.mod` require version (and
  separate consumer **MainRepo** `go.mod` baseline when Path ≠ MainRepo),
  worktree presence/absence, main HEAD advanced, zero mutation on cycle reject,
  banner display path, staging honor for `--add-all`.
- **Show-graph output (locked structure)**  
  - Human banners: `==== unwind graph (repo) ====`, `==== unwind graph (module) ====`,
    `==== status summary ====`.  
  - Repo: all stack members (dirty + clean); peel order = dirty free-first only
    (`(none)` when empty); display paths via peelDisplayPath policy.  
  - Module: all go.mod under stack; require/replace edges among stack modules.  
  - Summary counts + optional `hint: apply would need …` (not an error).  
  - `--json`: snake_case keys `repos` / `modules` / `summary` / `warnings`;
    `repos.peel_order` array; no ANSI.
- **Error surfaces**  
  - Cycle: combined stderr/stdout contains `cycle`; exit ≠ 0; no mutations;
    show-graph must not print a successful graph body.  
  - Missing pin flags when edges exist (dry-run/apply only): names `--tag-next`
    and `--push`; exit ≠ 0. Show-graph does **not** require those flags.  
  - Show-graph + `--dry-run` / apply partner: Error mentions `--show-graph` and
    the partner flag; exit ≠ 0.  
  - Tidy fail: mentions `go mod tidy` (and consumer path); **includes go child
    diagnostic body** (not only `exit status 1`).
  - Missing / non-git local-replace target: **`warning:`** on stderr; plan or
    show-graph continues; exit 0 when otherwise OK (not a hard fail).
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
├── cycle/
│   ├── two-cycle/                        # dry-run cycle reject
│   └── apply-two-cycle/                  # apply-mode cycle still pre-mutation
├── cascade/                              # free-module cascade (P1–P4 + replace-only pin)
│   ├── dry-run/                          # P1 plan vocabulary (+ C-DR7/8 replace-only)
│   │   ├── with-tag-next/                # cascade section emitted
│   │   │   ├── single-repo-two-modules/  # C-DR1: free-first tag leaf → pin root
│   │   │   ├── multi-repo-free-first/    # C-DR2: peel + module free-first across stack
│   │   │   ├── skip-testdata-scope/      # C-DR4: no tag-next for testdata scopes
│   │   │   ├── reinstall-local-tail/     # C-DR6: cascade + reinstall tail; zero mut
│   │   │   ├── replace-only-external-clean-dep/ # C-DR7: external replace ⇒ pin @ current
│   │   │   └── replace-only-intra-no-pin/       # C-DR8: intra replace alone ⇒ no pin
│   │   ├── without-tag-next/             # C-DR3: peel-only; no cascade tag/pin
│   │   │   └── peel-only-no-cascade/
│   │   └── cycle/                        # C-DR5: cycle reject; no cascade body
│   │       └── two-cycle-no-cascade/
│   └── apply/                            # P2 clean; P3 partial-edit; P4 reinstall; B1 pin-before-feature
│       ├── clean/
│       │   ├── single-repo-two-modules/  # C-AP1+3+4: tag/pin/commit-before-tag/keep-replace
│       │   ├── multi-repo-free-first/    # C-AP2: free-first across repos + pin commit
│       │   └── root-only-nested-tool-pin/ # C-AP: root-only tagscope + browser-debug pin; no tag collide
│       ├── dirty-gomod/
│       │   ├── without-add-all/          # C-AP5 / P3-1: partial-edit success + WIP preserve
│       │   └── with-add-all/             # C-AP6 / P3-5: --add-all → succeed (no partial)
│       ├── partial-edit/                 # P3 variants (no --add-all)
│       │   ├── unrelated-wip-file/       # P3-2: only go.mod/go.sum in pin commit
│       │   ├── sequential-two-pins/      # P3-3: two pins same consumer; both bumps
│       │   └── tidy-fail-restores/       # P3-4: tidy fail → exact WIP restore
│       ├── reinstall-local/              # P4: cascade + reinstall-local ship gate
│       │   ├── nested-skip-consumer/     # C-RI1: nested skip pin + reinstall success
│       │   └── multi-repo/               # C-RI2: multi-repo cascade + reinstall tail
│       └── pin-before-feature/           # B1 interleaved: pin auto-commit before feature gen-commit
│           ├── external-clean-dep-gen-commit/       # T1: clean free + hook; pin @ v0.0.1 then feature
│           └── free-dirty-then-consumer-gen-commit/ # T2: free peel/tag → pin → consumer gen-commit
├── show-graph/                           # read-only repo+module graph
│   ├── reject/                           # mutual exclusion with dry-run / apply partners
│   │   ├── with-dry-run/
│   │   ├── with-tag-next/
│   │   ├── with-done/
│   │   ├── with-merge-back/
│   │   ├── with-reinstall-local/
│   │   └── with-gen-commit-msg/
│   ├── cycle/
│   │   └── two-cycle/                    # show-graph cycle reject (no success body)
│   └── success/                          # acyclic → graph printed; exit 0; zero mutations
│       ├── human/                        # polished human format (dir / → / replaced)
│       │   ├── single-repo/
│       │   │   ├── clean-peel-none/      # clean main; peel (none); dir .
│       │   │   └── dirty-peel-dot/       # dirty main; peel .; dir identity
│       │   ├── multi-repo/
│       │   │   ├── free-first-order/     # 3 dirty free-first + modules @
│       │   │   ├── clean-leaf-listed/    # clean leaf in nodes; not in peel
│       │   │   ├── require-edges/        # collapsed require edges + modules @
│       │   │   └── apply-hint/           # hint: apply would need land/pin flags
│       │   ├── module/
│       │   │   ├── replace-edge/         # replaced (not replace =>); multi-repo keys
│       │   │   ├── nested-module-listed/ # dirs . + pkgs/shared; collapsed replaced
│       │   │   └── require-drift/        # (latest …) when require ≠ dep latest
│       │   ├── tagscope/
│       │   │   └── latest-and-next/      # latest + next / owned-changed on dir .
│       │   ├── inventory-warn/
│       │   │   └── missing-replace-target/  # warning: + graph + exit 0
│       │   └── color/
│       │       ├── force-color/          # --color → ANSI on stdout
│       │       └── no-color/             # --no-color → plain stdout
│       └── json/
│           ├── shape-keys/               # repos/modules/summary/warnings + peel_order
│           └── free-first-peel-order/    # peel_order free-first array
└── verify/                               # read-only post-job audit (Classic TDD RED)
    ├── reject/                           # mutual exclusion + bare --verify
    │   ├── bare-verify-without-unwind/
    │   ├── with-show-graph/
    │   ├── with-dry-run/
    │   ├── with-tag-next/
    │   ├── with-done/
    │   ├── with-merge-back/
    │   ├── with-reinstall-local/
    │   ├── with-gen-commit-msg/
    │   └── with-push/
    ├── cycle/
    │   └── two-cycle/                    # cycle preflight; no verify body
    ├── pass/                             # all error checks pass → exit 0
    │   ├── clean-stack/
    │   ├── multi-repo-pinned/            # require == latest; no droppable replace
    │   ├── inventory-warn/
    │   │   └── missing-replace-target/   # warning: + exit 0
    │   └── color/
    │       ├── force-color/
    │       └── no-color/
    ├── fail/                             # any error check FAIL → exit 1
    │   ├── dirty-peel/
    │   ├── needs-land/
    │   ├── owned-changed/
    │   ├── require-drift/
    │   ├── droppable-replace/
    │   ├── cascade-pending/
    │   └── force-color/                  # ANSI on FAIL
    └── json/
        ├── shape-keys-pass/
        └── fail-status/
```

Split factor (MECE, significance-first):

1. **Mode** — dry-run plan | apply mutation | cycle preflight | **cascade** |
   **show-graph** | **verify**.
2. Within dry-run: **order/skip/flags** | **gen-commit staging vocabulary** |
   **follow-local-replace inventory expansion**.
3. Within **cascade**: **dry-run vs apply** → (dry-run) `--tag-next`/cycle/shape;
   (apply) **go.mod dirt policy** (clean | dirty+`--add-all` | partial-edit without
   add-all) **or reinstall-local** integration → shape / multi-pin / fail-restore /
   nested-skip reinstall.
4. Within apply (non-cascade): **pin target Path** | **pin module selectivity** |
   **tidy error surfacing** | **already-main no land** | **done removes WT**.
5. Cycle (repo): dry-run vs apply-mode (same reject, no mutations).
6. Within **show-graph**: **outcome** (reject | cycle | success) → stack shape /
   graph content → format (human | json).
7. Within **verify**: **outcome** (reject | cycle | pass | fail | json) →
   primary check id / soft-warn / color / JSON shape.

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
| G1 | show-graph/reject/with-dry-run | `--show-graph --dry-run` mutual exclusion | **RED** until show-graph |
| G2 | show-graph/reject/with-tag-next | exclusive with `--tag-next` | **RED** |
| G3 | show-graph/reject/with-done | exclusive with `--done` | **RED** |
| G4 | show-graph/reject/with-merge-back | exclusive with `--merge-back` | **RED** |
| G5 | show-graph/reject/with-reinstall-local | exclusive with `--reinstall-local` | **RED** |
| G6 | show-graph/reject/with-gen-commit-msg | exclusive with `--gen-commit-msg` | **RED** |
| G7 | show-graph/cycle/two-cycle | cycle → non-zero; no success graph body | **RED** |
| G8 | show-graph/success/human/single-repo/clean-peel-none | clean main; peel (none); **dir** `.` identity | **RED** (human polish) |
| G9 | show-graph/success/human/single-repo/dirty-peel-dot | dirty main; peel `.`; dir identity | **RED** |
| G10 | show-graph/success/human/multi-repo/free-first-order | free-first + `modules @` grouping | **RED** |
| G11 | show-graph/success/human/multi-repo/clean-leaf-listed | clean leaf listed; not in peel | **RED** |
| G12 | show-graph/success/human/multi-repo/require-edges | collapsed `→` require + modules @ | **RED** |
| G13 | show-graph/success/human/multi-repo/apply-hint | hint names land/pin flags (not error) | soft/GREEN-ok |
| G14 | show-graph/success/human/module/replace-edge | **`replaced`** not `replace =>`; multi-repo keys | **RED** |
| G15 | show-graph/success/human/module/nested-module-listed | dirs + collapsed `→` + `replaced` | **RED** |
| G15b | show-graph/success/human/module/require-drift | `(latest …)` require-drift | **RED** |
| G16 | show-graph/success/human/tagscope/latest-and-next | latest + next on dir `.` | **RED** |
| G17 | show-graph/success/human/inventory-warn/missing-replace-target | warning: + graph + exit 0 | soft/GREEN-ok |
| G18 | show-graph/success/json/shape-keys | JSON keys repos/modules/summary/warnings | **GREEN** |
| G19 | show-graph/success/json/free-first-peel-order | peel_order free-first array | **GREEN** |
| G20 | show-graph/success/human/color/force-color | `--color` → ANSI on stdout | **RED** |
| G21 | show-graph/success/human/color/no-color | `--no-color` → plain / flag accepted | **RED** until flag |
| V1 | verify/reject/bare-verify-without-unwind | bare `--verify` requires `--unwind` | **RED** until verify flag |
| V2 | verify/reject/with-show-graph | exclusive with `--show-graph` | **RED** |
| V3 | verify/reject/with-dry-run | exclusive with `--dry-run` | **RED** |
| V4 | verify/reject/with-tag-next | exclusive with `--tag-next` | **RED** |
| V5 | verify/reject/with-done | exclusive with `--done` | **RED** |
| V6 | verify/reject/with-merge-back | exclusive with `--merge-back` | **RED** |
| V7 | verify/reject/with-reinstall-local | exclusive with `--reinstall-local` | **RED** |
| V8 | verify/reject/with-gen-commit-msg | exclusive with `--gen-commit-msg` | **RED** |
| V9 | verify/reject/with-push | exclusive with `--push` | **RED** |
| V10 | verify/cycle/two-cycle | cycle → non-zero; no verify body | **RED** |
| V11 | verify/pass/clean-stack | clean tagged main; all 6 checks pass; exit 0 | **RED** |
| V12 | verify/pass/multi-repo-pinned | require==latest; no droppable replace; pass | **RED** |
| V13 | verify/pass/inventory-warn/missing-replace-target | `warning:` + checks pass; exit 0 | **RED** |
| V14 | verify/pass/color/force-color | `--color` → ANSI on pass stdout | **RED** |
| V15 | verify/pass/color/no-color | `--no-color` → plain | **RED** |
| V16 | verify/fail/dirty-peel | dirty stack → dirty-peel FAIL; exit 1 | **RED** |
| V17 | verify/fail/needs-land | linked dirty → needs-land FAIL; exit 1 | **RED** |
| V18 | verify/fail/owned-changed | next tag planned → owned-changed FAIL | **RED** |
| V19 | verify/fail/require-drift | require ≠ latest → require-drift FAIL | **RED** |
| V20 | verify/fail/droppable-replace | external replace remains → FAIL | **RED** |
| V21 | verify/fail/cascade-pending | cascade would still run → FAIL | **RED** |
| V22 | verify/fail/force-color | `--color` ANSI on FAIL report | **RED** |
| V23 | verify/json/shape-keys-pass | JSON keys + 6 pass checks; exit 0 | **RED** |
| V24 | verify/json/fail-status | JSON require-drift fail; result fail; exit 1 | **RED** |
| C-DR1 | cascade/dry-run/with-tag-next/single-repo-two-modules | single-repo root→shared; free-first tag shared then pin root; peel `.`; zero mut | **GREEN** (P1 sealed) |
| C-DR2 | cascade/dry-run/with-tag-next/multi-repo-free-first | multi-repo both dirty; peel free-first **and** module cascade leaf before consumer | **GREEN** (P1 sealed) |
| C-DR3 | cascade/dry-run/without-tag-next/peel-only-no-cascade | no `--tag-next` → peel-only; no cascade tag/pin lines | **GREEN** |
| C-DR4 | cascade/dry-run/with-tag-next/skip-testdata-scope | real free module tag-next present; **no** tag-next for testdata scopes | **GREEN** (P1 sealed) |
| C-DR5 | cascade/dry-run/cycle/two-cycle-no-cascade | cycle → non-zero `cycle`; no successful cascade body | **GREEN** |
| C-DR6 | cascade/dry-run/with-tag-next/reinstall-local-tail | tag-next cascade + `would: reinstall local binaries` tail; zero mut | **GREEN** (P1 sealed) |
| C-DR7 | cascade/dry-run/with-tag-next/replace-only-external-clean-dep | clean free dep + matching require + external replace ⇒ `would: pin … @ v0.0.1`; no leaf peel/tag-next | **RED** until external replace ⇒ needs-pin |
| C-DR8 | cascade/dry-run/with-tag-next/replace-only-intra-no-pin | intra `replace => ./pkgs/shared` alone; matching require; no owned-changed → **no** cascade pin | **GREEN** control (D4) |
| C-AP1 | cascade/apply/clean/single-repo-two-modules | clean Base; tag shared; pin root keep-replace; pin commit; root tag after pin; push | **GREEN** (P2 sealed) |
| C-AP2 | cascade/apply/clean/multi-repo-free-first | multi-repo free-first land + cascade; leaf tagged/pushed; root pin commit | **GREEN** (P2 sealed) |
| C-AP5 | cascade/apply/dirty-gomod/without-add-all | **P3-1:** dirty go.mod WIP, no `--add-all` → partial-edit success; WIP preserve + pin commit Base-only | **RED** until partial edit (justified flip from P2 Error) |
| C-AP6 | cascade/apply/dirty-gomod/with-add-all | **P3-5:** dirty go.mod + `--add-all` → cascade success (no partial restore) | **GREEN** (P2 sealed / regression) |
| P3-2 | cascade/apply/partial-edit/unrelated-wip-file | dirty go.mod + unrelated WIP file; pin commit only go.mod/go.sum; file stays untracked | **RED** / may GREEN after P3 |
| P3-3 | cascade/apply/partial-edit/sequential-two-pins | two free modules pin same dirty consumer; both bumps on WT + commits | **RED** / may GREEN after P3 |
| P3-4 | cascade/apply/partial-edit/tidy-fail-restores | tidy fail mid partial-edit → exact WIP restore; non-zero | **RED** / may GREEN after P3 |
| C-RI1 | cascade/apply/reinstall-local/nested-skip-consumer | nested skip consumer old require+replace; cascade pin + `--reinstall-local`; no tidy/unknown revision | **RED** if nested reinstall still fails; **GREEN** backfill OK after P2/P3 |
| C-RI2 | cascade/apply/reinstall-local/multi-repo | multi-repo free-first cascade + reinstall-local tail; exit 0; no tidy/unknown revision | mixed / often **GREEN** after P2 |
| C-RI3 | *(covered)* | dry-run reinstall vocabulary = **C-DR6** sealed; C-RI1 soft-checks apply pin/tag-next log | no new leaf |

### C-AP5 sealed-test modification (P3 justification)

P2 C-AP5 expected **hard Error** for dirty go.mod without `--add-all`. Product
intent **D11** (P3): ordinary WIP uses **partial edit success**, not hard fail.
Orchestrator pre-approved rewriting C-AP5 ASSERT to success + WIP preserve.
Do not silently flip other sealed P2 leaves.

## How to Run

```sh
doctest vet ./cmd/wrk/tests/unwind
doctest vet ./cmd/wrk/tests/unwind/show-graph
doctest test -count=1 ./cmd/wrk/tests/unwind
doctest test -count=1 ./cmd/wrk/tests/unwind/cascade
doctest test -count=1 ./cmd/wrk/tests/unwind/cascade/apply
doctest test -count=1 ./cmd/wrk/tests/unwind/cascade/apply/reinstall-local
doctest test -count=1 ./cmd/wrk/tests/unwind/cascade/apply/partial-edit
doctest test -count=1 ./cmd/wrk/tests/unwind/cascade/apply/dirty-gomod/without-add-all
doctest test -count=1 ./cmd/wrk/tests/unwind/cascade/dry-run/with-tag-next/replace-only-external-clean-dep
doctest test -count=1 ./cmd/wrk/tests/unwind/cascade/dry-run/with-tag-next/replace-only-intra-no-pin
doctest test -count=1 ./cmd/wrk/tests/unwind/show-graph
doctest test -count=1 ./cmd/wrk/tests/unwind/verify
doctest test -count=1 ./cmd/wrk/tests/unwind/dry-run/follow-local-replace
doctest test -count=1 ./cmd/wrk/tests/unwind/dry-run/free-first-order
doctest test -count=1 ./cmd/wrk/tests/unwind/apply/multi-module-pin-require-root-only
doctest test -count=1 ./cmd/wrk/tests/unwind/apply/tidy-error-surfaces-go-stderr
doctest test -count=1 ./cmd/wrk/tests/unwind/apply/pin-on-linked-consumer-not-main
doctest test -count=1 ./cmd/wrk/tests/unwind/apply/leaf-then-pin
```

Expected this cycle (**verify**): all **V1–V24** leaves **RED** until `--verify`
flag, mutual exclusion, six-check report (human+JSON), and exit policy land.
Do **not** rewrite sealed show-graph / cascade asserts.

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
