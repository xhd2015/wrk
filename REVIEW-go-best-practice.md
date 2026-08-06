# Go best-practice review: `wrk`

**Date:** 2026-08-06  
**Scope:** codebase structure, CLI design, flag handling, package layout  
**Reference:** [go-best-practice](https://github.com/xhd2015) skill topics — `cli/*`, `flags-parsing/*`, `cmd-exec`, `go-embed-assets`, `kool-create`  
**Method:** read-only inspection of `cmd/wrk`, `wrkcli`, `workops`, `wrk-react`, `docs/skills/wrk`, `script/`, and `go.mod`. No code changes in this review.

---

## Executive summary

`wrk` is a mature, flag-composed CLI with strong test coverage (large doctest tree), deliberate use of `less-flags` (including `Cut("--exec")` and `HelpNoExit`), and a partial library split (`workops`). Several practices already match go-best-practice well: skill packaging (`wrk skill`), dry-run gates on commit/push/tag pipelines, git helpers via `xgo/support/cmd`, and a fat embedded SPA for `go install`.

The main gaps are **maintainability and consistency**, not missing features:

1. **God-router in `wrkcli/run.go` (~3.8k lines)** — one flat flag table + a huge mutual-exclusion matrix instead of mode-local parse/help.
2. **Color policy incomplete** — `--color` + `NO_COLOR`, but no `--no-color` and duplicated resolve logic.
3. **External commands mixed** — `xgocmd` for some git paths; raw `os/exec` for go/gh/agent/UX.
4. **Dry-run inconsistency** — plan-then-gate in core flows; separate dry-run functions for bash integration (and some reinstall print helpers).
5. **Package layout pressure** — `wrkcli` is a kitchen-sink package; `workops` extraction is incomplete; web registration works around an import cycle.

Findings below are ordered by severity. Recommendations are grounded in the named go-best-practice topics. **No fixes implemented in this pass.**

---

## Project snapshot

| Area | Layout |
|------|--------|
| Entry | `cmd/wrk/main.go` → `wrkcli.Run` (registers web serve to break import cycle) |
| Core CLI | `wrkcli/` (~50+ `.go` files; `run.go` ~3849 lines) |
| Library ops | `workops/` (status, sync, push, tag, merge-back, where, projects) |
| Web | `wrk-react/` + `wrkcli/web` (`//go:embed all:dist`) + `wrkcli/wrkserver` |
| Skill | `docs/skills/wrk` (`//go:embed SKILL.md`) via `skillcmd.SingleSkill` |
| Flags | `github.com/xhd2015/less-flags` (top-level only for most modes) |
| Tests | Extensive doctests under `cmd/wrk/tests/` |

**CLI surface (product choice):** almost everything is **long flags / short aliases**, not subcommands. Modes compose (`--done --sync --tag-next --push`, pipeline with `--gen-commit-msg`, `--main` as scope). The only true subcommand dispatch is `wrk skill …` (and flag-triggered exclusive paths like `--bash-integration`, `--set-config`).

That flag-compose model is intentional and tested; it is also the root cause of the mutual-exclusion complexity below.

---

## Findings (by severity)

### Critical / high

#### H1. Monolithic flag router and mutual-exclusion matrix (`wrkcli/run.go`)

**Evidence**

- Single `lessflags` chain registers ~50 flags (`--done`, `--status`, `--bring`, create UX pair flags, `--exec` Cut, etc.) then hundreds of lines of `otherMode` / mutual-exclusion checks (~63 “mutually exclusive / only valid with” sites in `run.go` alone).
- Mode resolution is centralized (`resolveCommand`, `isCreateMode`, early peels for skill / bash-integration / set-config / version / gen-commit-msg).
- Help is one giant `usage()` string; many modes lack a first-class `wrk <mode> --help` surface.

**Why it matters (go-best-practice)**

- `flags-parsing/subcommand`: every command level should parse its own flags and answer `-h`/`--help` for **that** level. A single top-level parse works for small CLIs; at this size it becomes a correctness and review risk.
- `flags-parsing/collect`: parent/child flag peeling is done ad hoc (`peelGenCommitMsgForCompose`, manual scans) instead of `CollectParsedFlags` / `Remove` / `Reconstruct`.

**Impact**

- Hard to extend modes without reopening the global matrix (easy to miss a compose partner or exclusivity).
- Error messages often say “mutually exclusive with other modes” without naming the conflicting flag.
- Onboarding cost for contributors is dominated by `run.go`, not by domain packages.

**Recommended changes**

1. **Keep flag-compose UX** (do not force git-style subcommands overnight), but **structurally split parse + dispatch**:
   - Top-level: global flags only (`-v`, `--color`/`--no-color`, maybe `-h`) + `StopOnFirstArg` **or** explicit mode detection first, then mode-local `lessflags.Parse`.
   - Or: keep flat argv, but extract **mode handlers** with their own option structs and validation (`runDoneOpts`, `runWebOpts`, …) so `run()` is a thin dispatcher.
2. Encode exclusivity in a **data table** (mode × allowed modifiers) rather than hand-written `otherMode := a \|\| b \|\| …` blocks; generate tests from the table.
3. Where flags are peeled for a child tool (`--gen-commit-msg`), adopt **`CollectParsedFlags` + `Remove` + `Reconstruct`** (`flags-parsing/collect`).
4. Per-mode help: `wrk --done --help` / nested handlers for set-config (already partially done) and bash-integration (currently weak).

**Topics:** `flags-parsing`, `flags-parsing/subcommand`, `flags-parsing/collect`, `flags-parsing/cut` (already used well for `--exec`).

---

#### H2. Color policy incomplete vs `cli/color`

**Evidence**

- `Bool("--color", &colorFlag)` only; **no `--no-color`** anywhere.
- `color.go`: `--color` forces on; auto uses TTY + `NO_COLOR`; good prefix-only Error/warning coloring.
- Duplicate resolvers:
  - `stderrColorEnabled()` / package `forceStderrColor`
  - `reinstallDiagColorEnabled(colorFlag)`
  - `task_like.go` ignores `--color` (TTY + `NO_COLOR` only)
  - `create_ux.go` / `pr.go` `stdoutColorEnabled` patterns

**Why it matters**

`cli/color` requires a three-mode policy:

| Mode | Selection | Behavior |
|------|-----------|----------|
| Auto | neither flag | TTY unless `NO_COLOR` non-empty |
| Always | `--color` | on (ignore TTY/`NO_COLOR`) |
| Never | `--no-color` | off (ignore TTY/`NO_COLOR`) |

Conflict: `--color` and `--no-color` must error with a fixed message.

**Recommended changes**

1. Add `--no-color`; mutual exclusion with `--color`.
2. Introduce a single `ColorMode` / `ResolveColor(...)` used by stderr prefixes, PR success tokens, reinstall diagnostics, and task-like prompts (respect `--color` everywhere).
3. Document in `usage()` and README: `--color`, `--no-color`, `NO_COLOR=1`.

**Topics:** `cli/color`.

---

#### H3. Package layout: `wrkcli` kitchen sink + incomplete `workops` boundary

**Evidence**

- `wrkcli` owns CLI parsing, storage side effects, git/go orchestration, dashboard TUI, PR/`gh`, reinstall planner, bash integration, skill glue, capture harness, and more.
- `workops` is a thin library for a subset (push/sync/tag/merge-back/where/projects/status) with `DryRun` on options — good direction — but many parallel CLI paths still live only in `wrkcli` (`runDone`, `runBring`, `runCreate`, `plan_local_reinstall`, `propagate_tags`, etc.).
- Import cycle workaround: `cmd/wrk` registers `web.Serve` into `wrkcli` because `wrkserver` imports `wrkcli`.

**Why it matters**

Go package best practice for large CLIs: thin `cmd/`, library packages by domain, CLI glue that only parses flags and maps to library APIs. Partial extraction without a clear dependency rule invites dual implementations (CLI vs library drift).

**Recommended changes**

1. **Define layers** and enforce import direction:
   - `workops` (or split: `wrk/gitops`, `wrk/projects`, `wrk/pipeline`) — pure operations, no `os.Args`, minimal stdout policy.
   - `wrkcli` — flags, human I/O, confirmation, follow-up cd, TUI.
   - `wrkcli/wrkserver` + `wrkcli/web` — HTTP; prefer depending on workops, not on full CLI.
2. Move remaining domain logic out of `run.go` into named packages/files; leave `run.go` as parse → validate → call.
3. Revisit the web cycle: extract shared API types / run entrypoints so `wrkserver` does not need the full CLI package.

**Topics:** general Go layout (supporting all CLI recipes); enables cleaner dry-run and cmd-exec later.

---

### Medium

#### M1. Dry-run: mixed adherence to `cli/dry-run`

**What is good**

- Manual commit (`commit_manual.go`): same path; `if dryRun { would: …; return }` after shared staging checks.
- Push (`push_main.go` / `workops.Push`): plan lines + library `DryRun`.
- Tag-next / propagate-tags / unwind: plan formatting with dry-run vocabulary; compose stages thread `dryRun`.
- Active-root pipeline passes one `dryRun` through stages — aligns with “one pipeline, gate side effects.”

**What drifts**

- **Bash integration:** `installBashIntegrationDryRun()` / `uninstallBashIntegrationDryRun()` as **sibling functions** to live install/uninstall — classic anti-pattern if discovery/status logic diverges from apply.
- **Reinstall-local:** shared plan (`PlanLocalReinstallsFromWorkDir`) then branches to `printMultiLocalReinstallDryRun` vs `executeMultiLocalReinstalls` — acceptable if execute never recomputes a different plan; keep print helpers as pure formatters of the same plan (delete legacy single-module dry-run path if unused).
- Vocabulary varies: `would:`, `dry-run: would …`, plan JSON, FormatUnwindDryRun — fine for humans, but document a house style.

**Recommended changes**

1. Bash integration: one `applyBashIntegration(action, dryRun bool)` that computes statuses once and either mutates or prints would-lines (`cli/dry-run`).
2. Prefer `if dryRun` gates at side-effect sites over `*DryRun()` function pairs.
3. Keep compose pipelines single-path (already largely true).

**Topics:** `cli/dry-run`.

---

#### M2. External commands: partial `cmd-exec` adoption

**Evidence**

- `gitexec.go` documents and uses `github.com/xhd2015/xgo/support/cmd` for non-interactive git (`gitRunDir`, `gitOutputDir`, capture variants).
- Still heavy `os/exec.Command` for: `go mod tidy/edit`, `go env`, `go build` (propagate), `go install` (reinstall), `gh` (PR), agent/iTerm UX, Vite in web serve, `--exec` user command.
- Verbose logging is custom (`logGitCommand` / `logGoCommand`) rather than `cmd.Debug()` consistently.

**Why it matters**

`cmd-exec` recipe: fluent builder, Dir/Env, capture, Debug pre-line, inherit or redirect streams consistently. Mixed styles make env, cwd, and error wrapping inconsistent (and harder to mock).

**Recommended changes**

1. Prefer `xgocmd` for **all non-interactive** go/git/gh probes; keep raw `*exec.Cmd` only where streaming to TTY or full duplex is required (worktree add verbose, interactive shells, user `--exec`).
2. Align verbose pre-lines with one helper (or `cmd.Debug()` when process stderr is free).
3. Centralize “run in module dir” for go tools next to `goModTidy`.

**Topics:** `cmd-exec`.

---

#### M3. Nested help gaps (`flags-parsing` / `cli/skill-cli`)

**What is good**

- Top-level `Help("-h,--help", usage()).HelpNoExit()`.
- `wrk skill` empty args → skill help; uses `skillcmd.SingleSkill` with list/show/install — matches **Shape 1** skill CLI.
- `wrk --set-config` peels nested help for create/show (good multi-level pattern).

**Gaps**

- `--bash-integration`: `-h`/`--help` are accepted but effectively no dedicated help text (“Hidden from main help”).
- Mode modifiers (`--scan-git-repos`, `--web`, pipeline stages) only appear in the giant root usage; users cannot get a focused “only these flags apply here” page without reading the whole blob.
- Skill is single-skill only (fine); `wrkModeFlags` / `skillLocalFlags` / `flagValueArgs` / `setConfigDisallowedFlags` / `bashIntegrationDisallowedFlags` are **parallel flag inventories** that can drift from the lessflags registration list.

**Recommended changes**

1. Generate or share one **canonical flag registry** used by: lessflags registration, completion (`bashintegration`), skill conflict detection, set-config disallow list.
2. Add bash-integration-level help; optionally mode help via `HelpFunc` after mode is known.
3. Keep skill path early-return before top-level parse (already correct).

**Topics:** `flags-parsing/subcommand`, `cli/skill-cli`.

---

#### M4. `go:embed` assets: fat embed works; hydrate/placeholder layers unused

**Evidence**

- `wrkcli/web/serve.go`: `//go:embed all:dist` with committed `dist/index.html` + hashed JS/CSS (~250KB+).
- `script/build-frontend.sh`: bun build → copy into `wrkcli/web/dist`.
- Skill: `//go:embed SKILL.md`; version: `//go:embed VERSION.txt`.

**Assessment vs `go-embed-assets`**

| Layer | Status |
|-------|--------|
| Layer 1 placeholders for empty embed | N/A today — fat tree is committed, so bare `go install` compiles **and** ships UI |
| Layer 2 completeness check | Not formalized (`EmbedComplete`); binary always assumed complete |
| Layer 3 fat local / release | Effectively always-on via committed dist |
| Layer 4 runtime hydrate | Not used |

**Risks**

- Binary size and noisy diffs when frontend changes.
- If someone gitignores `dist/` without placeholders, `go install` breaks (classic anti-pattern).
- README documents rebuild; no automated “embed complete?” guard in CI described in-tree.

**Recommended changes** (pick a strategy, document it)

- **Stay fat-embed (current):** add CI check that `dist` has index + assets; document that release tags must include built assets.
- **Or** move to placeholder + release hydrate for leaner source module, per `go-embed-assets` Layer 1+4 — only if install size/source hygiene becomes a goal.

**Topics:** `go-embed-assets`.

---

#### M5. README vs `go.mod` local replace

**Evidence**

- README: local dev expects sibling `dot-pkgs` and `go.mod` `replace … => ../dot-pkgs/go-pkgs`.
- Current `go.mod`: **no `replace` directives**; module pulls published `dot-pkgs/go-pkgs v0.0.97`.

**Recommended changes**

- Align docs with reality (optional replace snippet for contributors), or document a `go.work` / script for local replaces so README does not imply a missing line.

---

### Low / positive notes

#### L1. Strengths already aligned with go-best-practice

| Practice | Where |
|----------|--------|
| `less-flags` fluent parse, aliases, `HelpNoExit` | `run.go` |
| `Cut("--exec")` for opaque command tails | `run.go` — textbook `flags-parsing/cut` |
| `*string` / `*int` for optional presence | task, port, PR title/comment |
| Skill Shape 1 via `skillcmd` | `skill.go` + embedded `SKILL.md` |
| Dry-run gates on commit/push/tag compose | multiple pipeline paths |
| Git capture for TUI (no stderr leak) | `gitOutputDirCapture` |
| Thin `cmd/wrk` + teapre side-effect init | `cmd/wrk/main.go` |
| workops options with `DryRun` | library-friendly API |
| Extensive doctests | `cmd/wrk/tests/**` |
| Reject equals-form footguns (`--where=`, `--pr=`) | intentional less-flags Bool contract |

#### L2. Create UX flag pairs

`--new-window` / `--no-new-window`, etc., are validated in `createUXFlags.validate()` — good local model. Consider the same “opts struct + validate()” for **every** mode as the extraction path from H1.

#### L3. `kool-create`

`wrk-react` is a Vite + React + Bun app with `file:../dot-pkgs/react` dependency — not necessarily scaffolded by `kool create go-react`, but it fills the same role. No requirement to re-scaffold; for **new** sibling tools, prefer `kool create go-react|frontend` so embed/scripts match house templates (`kool-create`).

#### L4. Streaming / TUI

Dashboard and verbose git/go streaming show awareness of live output (`cli/streaming`). Inline TUI mouse has dedicated tests under `wrkcli/tui/tests` — out of main CLI-flag scope but aligned with `cli/inline-tui-mouse` investment.

#### L5. Global package state

`forceStderrColor`, `invocationVerbose`, `worktree.GitVerboseLogger`, capture dir for skill install — workable for a single-process CLI, but complicate library reuse and parallel tests. Prefer `RunOpts` / context values as more modes go in-process (partially started with `RunOpts.ScanTestPauseAfterFirstPrint`).

---

## Recommended roadmap (no implementation in this review)

Phased so each step is independently shippable:

### Phase 1 — Quick consistency (low risk)

1. Add `--no-color` + single `ResolveColor` (`cli/color`).
2. Unify color checks in task-like / reinstall / PR / stderr.
3. Fix README replace/`go.mod` story.
4. Canonical flag name list shared by skill/bash/set-config disallow maps.

### Phase 2 — `run.go` decomposition (high value)

1. Extract option structs + `validate()` per mode; thin `run()` dispatcher.
2. Replace sprawling `otherMode` booleans with a mode×modifier matrix + generated tests.
3. Use `CollectParsedFlags` for gen-commit peel/forward.
4. Mode- or feature-level `--help` where doctests allow.

### Phase 3 — Library boundary

1. Move more pure ops into `workops` (or split packages); CLI only formats and confirms.
2. Break `wrkserver` → `wrkcli` cycle via shared non-CLI API package.
3. Standardize external commands on `xgocmd` except interactive/stream cases (`cmd-exec`).

### Phase 4 — Dry-run / embed polish

1. Collapse bash-integration dry-run siblings into gated single path (`cli/dry-run`).
2. Document embed strategy (fat commit vs placeholder+hydrate) and add completeness CI (`go-embed-assets`).

---

## Mapping: findings → go-best-practice topics

| Finding | Primary topics |
|---------|----------------|
| H1 Monolithic router / flags matrix | `flags-parsing`, `flags-parsing/subcommand`, `flags-parsing/collect` |
| H2 Color policy | `cli/color` |
| H3 Package layout | (layout supports all topics); enables dry-run & cmd-exec |
| M1 Dry-run mix | `cli/dry-run` |
| M2 exec vs xgocmd | `cmd-exec` |
| M3 Nested help / flag inventories | `flags-parsing/subcommand`, `cli/skill-cli` |
| M4 Embed strategy | `go-embed-assets` |
| M5 Docs vs go.mod | docs hygiene |
| L3 Frontend scaffold | `kool-create` (future only) |
| Cut `--exec` (positive) | `flags-parsing/cut` |
| Skill packaging (positive) | `cli/skill-cli` |

---

## Out of scope / not raised as defects

- Product choice of **flag composition** over git-style subcommands (valid; complexity must be managed, not abandoned).
- Bubbletea dashboard design details beyond noting TUI tests exist.
- Frontend React architecture (`wrk-react`) beyond embed and build script.
- Doctest volume/cost (quality signal; not a go-best-practice CLI topic).

---

## Conclusion

`wrk` already adopts several house best practices (less-flags, Cut, skillcmd, dry-run gates on pipelines, fat embed for installable UI, partial workops library). The highest-severity issue is **structural**: one enormous flag-and-exclusion router in `wrkcli/run.go`, combined with incomplete color policy and uneven cmd/dry-run consistency. Addressing those without changing user-facing compose semantics will make the project match go-best-practice recipes more closely while preserving the CLI’s distinctive pipeline UX.

**Next step (when ready):** implement Phase 1 items only, or open a design doc for Phase 2 mode-matrix extraction before large refactors.
)
