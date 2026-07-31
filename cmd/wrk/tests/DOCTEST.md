# wrk Test Cases

## Version
0.0.3

Decision tree covering the `wrk` CLI: bare no-args **dashboard** (static snapshot /
interactive RUN compose), create via **`wrk --new`** (and create-selecting args), optional `wrk <dir>`
first positional, `wrk <dir> <target-dir>` second positional spawn-location override,
create UX (`$WRK_HOME/config.json` → `create.window` / `create.terminal` / `create.agent`
plus one-shot flags and `wrk --set-config --create …`), `wrk --bring` external dependency worktrees (best-effort Go replace; always worktree; soft SKIP
when not a module/dependency), `wrk --done` merge-back (including external cascade),
`wrk --done|--merge-back` post-pipeline composition with optional `--sync` /
`--tag-next` / `--push` (fixed order **sync → tag-next → push** after successful
primary; flag argv order free; subsets allowed; `--dry-run` plans all requested
stages with zero mutations; `--json` rejected with primary), `wrk --list`, `wrk --status`, project persistence
(`wrk --projects`, `wrk --add`, `wrk --rm`, `wrk --scan-git-repos`, auto-record,
events.jsonl), `wrk --cd` directory jump (in-place follow-up or fallback interactive
shell), `wrk --exec` cut-flag (run a trailing command in the mode target directory
after successful create / `--cd` / `--bring` / `--set-task` / `--done`), and
`wrk --web [--port PORT]` standalone HTTP workflow page (under `wrkcli`).

# DSN (Domain Specific Notion)

- **wrk CLI** — standalone binary; invocation form `wrk [dir] [flags...]`; first non-flag positional argument is optional `<dir>` — when present, effective cwd is the resolved absolute path of `<dir>` (process cwd unchanged); when absent, effective cwd is process cwd. **Bare no-args** (zero positionals and no mode/create-selecting flags) enters **dashboard** mode: non-TTY prints a fine-grained static stage snapshot (`[•]`/`[ ]`/`[-]`); TTY or hermetic `WRK_DASHBOARD_ACTION` drives interactive CANCEL/RUN (RUN builds real multi-stage compose argv). **Create** is **`wrk --new`** (or create-selecting args: `<dir>`, `-t`/`--task`, create UX flags, `--no-config`, `--exec`, …) — creates a git worktree and prints the target path on stdout. `wrk --done` merges the linked worktree branch back into main and removes the worktree + branch (`worktree.MergeBack` with `Remove: true`); `wrk --list` runs `git -C <effective-cwd> worktree list` and prints stdout unchanged; `wrk --status` resolves the effective cwd's git toplevel, scans it with `scan_repo.Scan`, and prints one status block per discovered git directory; missing `<dir>` → `wrk: <path> does not exist`.
- **wrk --new** — explicit create entry (former bare create). Mutually exclusive with other modes (`--done`, `--list`, `--status`, …). Event `command: "create"`; `args` may include `--new`.
- **WRK_HOME** — storage root env var (default `~/.wrk`); tests isolate per run at `{WorkRoot}/.wrk`.
- **WRK_DATE** — optional env var (`YYYY-MM-DD`) overriding the run date for deterministic naming; all tests set `WRK_DATE=2026-06-30`.
- **Work root** — temp directory holding source repos and move targets.
- **Naming** — worktree path `{WRK_HOME}/worktrees/{basename}-{token}-{YYYY-MM-DD}[-{slug}][-N]`; branch `{token}-{YYYY-MM-DD}[-{slug}][-N]` (**no** repo/dep basename on branch; **no** `/` in branch). `N` starts at 1 on collision (path exists **or** branch ref exists — joint walk). Always create a **new** branch via `git worktree add -b` (never reuse an existing branch ref; no `worktree add --no-checkout` / `checkout --ignore-other-worktrees` create path). Fixed user path (`wrk <dir> <missing-target>` parent exists): path stays exact; if preferred branch exists, suffix **branch only** (`-1`, `-2`, …). Wrk-managed invariant: `filepath.Base(path) == {basename} + "-" + branch`. No unsuffixed names without date.
- **token** — `sanitize(currentBranch)` = replace `/` with `-` for **both** path token and branch base segment; for detached HEAD, 7-char short commit hash from `git rev-parse --short=7 HEAD` (not literal `HEAD`).
- **Git source** — cwd must be inside a git checkout (main repo, linked worktree, or nested subdirectory); basename resolves from the checkout root when cwd is a linked worktree or nested subpath.
- **Removed: wrk --dep / wrk --all-deps** — hard-removed from the CLI surface (unknown flag / non-zero). External dependency worktrees use **`wrk --bring` only**. Doctest leaves under `bring/removed-flags/` assert rejection. No soft alias/deprecation. `--projects-dep-graph` is unrelated and kept.
- **wrk --bring** — sole external-dep worktree mode. Spawns a dependency worktree under `{consumerTop}/external/` as a worktree of the DEP repo (`git -C <depMain> worktree add`), registered under `<depMain>/.git/worktrees/` (NOT the consumer's). **Consumer module discovery**: scans the full consumer tree (`gotool/mod/scan.Scan`) for all Go modules — consumers without a root `go.mod` (module in a subdirectory) are supported. **Dep module discovery**: scans the dep tree for all Go modules. Matches each dep module against every consumer module's `require`/`replace`; on match runs `gotool.Replace` + `gotool.Tidy`. Appends `/external` to `.gitignore` when missing; prints external worktree abs path on stdout. Path naming: `{basename}-{token}-{WRK_DATE}[-N]`. Branch naming: `{token}-{WRK_DATE}[-N]` (no dep basename; always `worktree add -b`). Collision is joint path+branch against `depMain`. **Best-effort module analyse**: soft SKIP (exit 0 + worktree + gitignore + stdout path + stderr notice) when: dep is not a Go module; consumer has no Go modules; or modules exist but dep is not required/replaced by any consumer module. SKIP stderr substrings: `SKIP local dep replacement: <depPath> is not a go module`; `SKIP local dep replacement: consumer has no Go modules`; `SKIP local dep replacement: <depPath> is not a dependency of any consumer module`. Non-git consumer cwd: materialize under `{abs(cwd)}/external/`; soft-skip replace + gitignore. Flag name `--bring` only (no `--dep` alias). Mutually exclusive with other modes. Basename fallback via same `resolveDirArg` core as create (`allowBasenameFallback=true`). Event `command: "bring"`. `--exec` after successful `--bring` (including SKIP) runs in the external worktree. **Same-repo reuse (Policy A)**: live linked worktrees of `depMain` under `{consumerTop}/external/` → reuse lex-smallest (stderr reuse warnings; stdout path only). **`--no-dep`**: long-only bool valid **only** with `--bring`; materialize/reuse external worktree + gitignore; **skip module match scan entirely** (no SKIP messages; no replace/tidy); go.mod unchanged. Bare `--no-dep` or with other modes → non-zero, stderr `wrk: --no-dep is only valid with --bring`. With `-v`, may log/stream git worktree add but never logs `go mod tidy` under `--no-dep`.
- **wrk --done / --merge-back + --sync** — `--sync` is a **post-success modifier** of `--done` or `--merge-back` (flag order free: `--done --sync` ≡ `--sync --done`). After the primary succeeds (Action is not `"aborted"` / not error), run existing `runSync` from **main** (`MergeBack` `TargetPath`), never the removed worktree path. **Blank line** between primary stdout message and the sync stdout block when both print. Full post-pipeline order (when flags present): primary → **sync** → **tag-next** → **push** → `--exec` (**done** path only) → land/follow-up. Sync hard-fail after primary success → non-zero exit; **no undo** of primary. Partial sync skips same as standalone (`warning:` on stderr, exit 0 after successful primary). Event `command` stays `"done"` / `"merge-back"` (not `"sync"`); `args` include `--sync`. Multi-stage composition dry-run (`--done`/`--merge-back` + post flags + `--dry-run`) under `done-pipeline/dry-run/` and `merge-back-pipeline/dry-run/` plans all requested stages with zero mutations and no prompts. Still invalid: `--merge-back` + `--exec` (with or without `--sync`). Standalone `--sync` behavior unchanged; `--sync` remains exclusive with list/status/create/other non-composed modes (and with bare `--tag-next` / `--push` without a primary).
- **wrk --push (dual meaning)** — never bare (`wrk --push` alone → non-zero). **(1) Branch under primary**: with `--done` or `--merge-back` and **without** `--tag-next`, after successful primary (not aborted / not dry-run apply), call `runPushMain` from **main** with empty tag refs — pushes the main branch only (flag order free: `--done --push` ≡ `--push --done`; same for merge-back). **(2) Tags with standalone tag-next**: `wrk --tag-next --push` pushes newly created tag refs only (no branch push). **(3) Both when primary + tag-next + push**: create tags locally first, then `runPushMain(main, false, createdTags)` pushes **branch + tags**. **Remote resolution** (primary path): prefer configured upstream of main’s current branch; else fallback `origin` + current branch name. **Stdout** (primary path): blank line before push stage, then stable confirmation `pushed main → origin/main\n`. Push hard-fail after primary success → non-zero exit; **no undo** of primary. No resolvable remote → clear non-zero error.
- **wrk --done / --merge-back post-pipeline (sync → tag-next → push → propagate-tags)** — after successful primary (not aborted / not dry-run apply), optional post steps run in **fixed order** regardless of flag argv order: **sync → tag-next → push → propagate-tags** (then **exec → land** on the **done** path only; merge-back keeps the source worktree and has no exec/land). Subsets allowed (e.g. `--tag-next` only, `--tag-next --propagate-tags`, `--propagate-tags` alone on existing tags, `--sync --tag-next`, full combo). **Blank line** between major stdout stages. When `--tag-next` / `--propagate-tags` is set with primary `--done` or `--merge-back`, dispatch must be the **primary** (not bare `runTagNext` / `runPropagateTags`); tag-next apply runs on **main** after merge; propagate runs from **main** + `WRK_HOME` using newly created tags (or planned next tags on dry-run) or existing source release tags when `--tag-next` is absent. Abort of primary: no post steps. Hard fail of a post step: stop, no undo. Event `command` stays `"done"` / `"merge-back"` (args may include `--sync` / `--tag-next` / `--push` / `--propagate-tags`). Coverage: flag matrix `done-compose/`; branch push `done-push/`; real apply `done-pipeline/` + `merge-back-pipeline/`; composition dry-run under each tree’s `dry-run/`.
- **--json with primary** — `--json` is **only** valid with bare `--tag-next` (machine-readable plan/result). `--done --json` and `--merge-back --json` → non-zero; stderr names `--json` and the primary (`wrk: --json is not valid with --done` / `… with --merge-back`). Not a silent accept into merge-back or tag-next.
- **wrk --done** — resolves checkout root via `ShowToplevel(cwd)`; requires a linked worktree (not main repo); clean worktree; implicit `--rm`. **Cascade**: `scan_repo.Scan(consumerTop)` discovers every git directory under the checkout; for each row where `RepoType == worktree` and `IsLinked(path)` and `path != checkoutRoot`, run `mergeBackExternalWorktree(path)` in scan path order (path-sorted). This covers `external/*` dep worktrees **and** manually linked worktrees elsewhere (e.g. `deps/foo`). Skip `RepoTypeMain` nested repos (no merge-back/delete). Each cascaded worktree is a dep-repo worktree (registered under `<depMain>/.git/worktrees/`), so `MergeBack` resolves its main repo from the worktree's `.git` gitdir (the dep main) and merges the dep branch back into the dep repo (the branch shares the dep's history, so merge-base resolves); this ensures dep work committed on a nested linked worktree is merged back before removal. Relation to dep main: already-included → remove only; ahead/diverged → default auto-yes merge/rebase (own + cascade); `--confirm` restores Y/n (with `--confirm-from-stdin` on non-TTY). The consumer's own `checkoutRoot` is excluded (finished by the final `MergeBack` in `runDone`). **Guard**: scan **every** Go module under the checkout (`gotool/mod/scan.Scan`) — main + all sub-modules — and classify each filesystem/local `replace` (`./`, `../`, or absolute path without version) by resolving its target relative to the module's `go.mod` dir: **intra-repo** = target dir exists AND `git -C <target> rev-parse --show-toplevel` equals the consumer's toplevel (a `../../`/`./sub` nested-module reference back into the same repo); **extra-repo** = everything else (`./external/foo` dep worktree, non-existent target, absolute path to another checkout, sibling outside) (`./external/foo` dep worktree, non-existent target, absolute/sibling outside). The guard names the offending `<top>/<m.Dir>/go.mod` file and each `replace <Old> => <New>` directive in its message. Default (no flag): intra-repo → **WARN to stderr and proceed** (exit 0, merge-back runs); extra-repo → **error, block**. `--no-in-module-replace` (opt-in, valid only with `--done`) → **all** local replaces block (fully-strict). A checkout with no `go.mod` at all yields zero modules → guard is a no-op → `--done` proceeds (and the linked-worktree check inside `MergeBack` still runs for a main-repo cwd, producing `not a linked worktree`). Branch relation to main: already-included → remove only; ahead/diverged → prompt then merge/rebase (or `-y` / `--confirm-from-stdin` on own worktree).
- **wrk --list** — runs `git -C <cwd> worktree list`; prints stdout unchanged; cwd must be inside a git work tree (main repo, linked worktree, or nested subpath). Mutually exclusive with no-args create and `--done`.
- **wrk --status** — standalone reporting mode; cwd must be inside a git work tree. Resolves the effective cwd's checkout root with git toplevel discovery, calls `scan_repo.Scan(context.Background(), scan_repo.Options{Roots: []string{Root}})`, and for main-repo roots prints primary then optional external sections (see Main-repo sections below); non-main roots keep scan-order discovery. Each block includes `Dir` via **`statusDirLine`** (see below), current branch, short commit hash plus subject, and `Status` as either `clean` or `dirty (<added> added, <changed> changed, <renamed> renamed, <deleted> deleted)` (wrk taxonomy: porcelain `??` untracked counts as **added**, same as index `A` / `wrk --projects`). **`statusDirLine` (all Dir lines)**: `rel = filepath.Rel(normalize(invocationCwd), normalize(repoPath))`; on Rel failure or when cleaned rel has **more than two** leading `..` segments → absolute `storage.NormalizePath(repoPath)`; else `filepath.ToSlash(rel)`. Pure Rel (no soft force of `.` merely because cwd is inside the checkout). Applies to scan blocks, appended external blocks, and broken/prunable minimal blocks. **Main-repo status content** (`worktree.IsMainRepo(checkoutRoot)`): the main-repo scan block also includes `Remote:` — same brief upstream labels as `--projects` (`identical`, `needs push`, `needs pull`, `diverged`, `(no upstream)`); field order is `Dir`, `Branch`, `Commit`, `Status`, `Remote` (no `Worktrees:`). `Remote:` is gated on **main identity** (repoPath is main), **not** on `Dir == "."`. `Remote:` uses local upstream tracking refs by default; with `--fetch`, fetch upstream for the main repo first then compare. Nested independent `RepoTypeMain` sub-repos and **linked worktree blocks** do **not** show `Remote:`. Running from a **linked worktree cwd** omits `Remote:` on all blocks (append phase also skipped). **Linked worktrees only** (`worktree.IsLinked`) also include one-line `Master:` — brief branch-relation label comparing main repo's current branch vs the worktree's current branch via `git.CompareBranches` (`identical`, `needs merge back(+N commit(s))`, `needs fast forward(+N commit(s))`, `diverged(N commit(s))`). Main checkout blocks (other than `Remote:` above) and nested independent `RepoTypeMain` repos do **not** show `Master:`. When stdout is a TTY or `--color` is set, `Status: clean` is green; dirty status uses granular red/grey segments (same rules as `--projects`); `Master:` and `Remote:` values follow `--projects` color rules when applicable; the main-repo external section header `---- external ----` is full-line grey (`ansiGrey` / `#90`). Without color: plain text (including plain header). **Scan-discovered broken** (alive checkout, `checkout.Enrich` or `masterBriefForRepo` fails during scan phase): minimal block with `Dir` via `statusDirLine` + `Status: error: <git stderr>` only (no `Branch`/`Commit`/`Remote`/`Master:`); red `error: …` when `--color`/TTY; run continues for remaining repos; exit 0; stderr empty (unless `-v`). Same non-fatal policy as `--projects` per-repo errors and appended broken blocks. **Main-repo sections (P2+P3)**: when `worktree.IsMainRepo(checkoutRoot)`, partition via `PartitionStatusPaths(main, scanPaths, ListLinked)` — **primary** = main then all ListLinked paths in porcelain order (in-tree + out-of-tree WRK + prunable); **external** = scan paths not in primary, path-sorted. Print primary blocks first; if external non-empty, after last primary: blank line, header `---- external ----` (P3: gray ANSI `ansiGrey` / `#90` when colorEnabled via TTY or `--color`; plain ASCII when color off), blank line, then external blocks. Omit header when external empty. Main-owned WRK out-of-tree linked worktrees are **primary** (no section header for them alone). Out-of-tree/prunable primary linked keep `printAppendedLinkedBlock` presentation (healthy: full + `Master:`; broken/prunable: minimal). Nested external always `printStatusBlock`. Running from a linked worktree cwd without `--main` skips main-repo sectioning (scan-only shortcut). Mutually exclusive with `--done`, `--list`, `--bring`, create target arguments, and other modes. **Composition with `--main`**: `wrk --main --status` and `wrk --status --main` (order irrelevant) run status of the **main repository** of the resolved checkout — resolve `ShowToplevel(workDir)` → `ResolveMainRepo` then `runStatus` on main for content; **Dir labels still use original invocation cwd**. No nested shell. Event `command` is `"status"` with `args` including both `--main` and `--status`. From an in-tree linked cwd, always full main-repo status (not the linked-cwd shortcut in `runStatusLinkedInTreeCwd`). Equivalence: same blocks and Branch/Commit/Status/Master/Remote as `(cd <mainRepo> && wrk --status)`; **Dir may differ** when invocation cwd ≠ main. Pure `wrk --main` (shell) and pure `wrk --status` (current checkout) stay unchanged. `--fetch` / `--color` / `-v` remain allowed with the pair.
- **-y / --yes** — universal top-level bool flag; no-op on commands without Y/n prompts (create, `--list`, basename `Select [1-N]`, etc.). **Default auto-yes** for `--done` / `--merge-back` (own worktree **and** cascaded deps) and `--set-task` rename: bare invocations skip `Proceed? [Y/n]` without `-y`. `-y` remains valid for compatibility. **`--confirm`** restores interactive prompts for those modes; with non-TTY use `--confirm --confirm-from-stdin`. Recorded in `events.jsonl` `args` when passed.
- **--confirm-from-stdin** — when set with piped `StdinInput`, reads Y/n from stdin for **own-worktree** merge-back confirmation on non-TTY ahead/diverged cases. Superseded by `-y` when `-y` is set (no stdin read). Does **not** confirm cascaded ahead/diverged worktrees on non-TTY (option A pre-flight guard).
- **--no-in-module-replace** — bool flag (no value); valid ONLY with `--done`. Restores the fully-strict local-replace guard: every filesystem/local `replace` (intra-repo or extra-repo) blocks `--done`. Without it (default), intra-repo replaces — whose target dir exists and shares the consumer's `git rev-parse --show-toplevel` (`../../`/`./sub` nested-module reference) — only WARN and `--done` proceeds; extra-repo replaces (`./external/foo` dep worktree, non-existent/absolute/sibling) still block. Bare `wrk --no-in-module-replace`, or with any other mode (`--bring`/`--list`/no-args create) → non-zero exit, stderr `wrk: --no-in-module-replace is only valid with --done`.
- **Request.Args** — CLI arguments passed to `wrk` after optional `<dir>` (empty + no TargetDir/TaskDesc → bare dashboard; create uses `["--new"]` or positionals/task/UX; `["--bring", depPath]` for bring tests; `["--done"]` or `["--done", "--confirm-from-stdin"]` for done tests; `["--list"]` for list tests).
- **Request.TargetDir** — when set, prepended as the first positional argument to `wrk` (`wrk <dir> ...`); used by `dir-arg/` tests to run from `WorkRoot` while targeting a repo elsewhere.
- **Request.SpawnDir** — optional second positional `<target-dir>` (`wrk <dir> <target-dir>`); appended after `TargetDir` when set. Overrides the worktree spawn location: missing target with existing parent → spawn exactly at `<target-dir>` (no naming suffix on the path); existing target dir → spawn a default-named sub-dir under it; missing parent → error. Resolved relative to the process (shell) cwd, not `<dir>`. Create-only — errors with `wrk: unexpected arguments` when combined with `--list`/`--done`/`--bring`. `WRK_HOME` is ignored when set. **Named-bring same-repo reuse (Policy B)**: before create, search **any** live linked worktree of the source `mainRepo` (not only under `<target-dir>`, not only `external/`). None → create as today. Some + **stdin TTY** → prompt on stderr `<basename> already exists in <absPath>, skip? [Y/n] ` (primary path = lex-smallest; multi may warn also-present); empty/`Y`/`y` → **skip** create, stdout = existing path, exit 0; `n`/`N` → create as today. Some + **non-TTY** → hard error on stderr (`refusing non-interactive create (default is skip; re-run in a TTY)`), empty stdout, non-zero; no flag override. Bare `wrk` / `wrk <a>` (no second positional) are **unchanged** (still free to create many under `~/.wrk/worktrees` with `-N`).
- **external/** — dependency worktrees live at `{consumerTop}/external/{basename}-{token}-{WRK_DATE}[-N]`; not under `WRK_HOME`. They are linked worktrees of the DEP repo (registered under `<depMain>/.git/worktrees/`), not the consumer — the consumer only hosts the working tree on disk.
- **deps/** — manually linked worktrees may also live under `{consumerTop}/deps/...` (or any path under the checkout); created via `git -C <depMain> worktree add` into the consumer tree. `--done` cascade discovers them via `scan_repo.Scan`, same as `external/*`.
- **runGitIsolated** / **gitOutputIsolated** / **gitWorktreeListIsolated** — thin wrappers over `github.com/xhd2015/gitops/git/git_isolated` (`MustRun`, `MustOutput`, `WorktreeList`).
- **git_isolated** — hook-free git runner; ignores global/system gitconfig; repo-local `core.hookspath` still applies.
- **Session fixtures** — doctest injects `d *session.Doctest` (`DOCTEST_ROOT` / `DOCTEST_CASE` / `DOCTEST_SESSION_ID`); harness adopts via `adoptDoctestContext` (**no** process `Setenv` of `DOCTEST_*`). Seeds live at `{DOCTEST_FIXTURE_ROOT}/{session}/seeds/` (macOS `cp -c`, Linux `cp -a`); `getWrkBin` builds once per session to `{DOCTEST_FIXTURE_ROOT}/{session}/bin/wrk` with a file lock across leaf processes.
- **Request.InProcess** — when true, `Run` uses `wrkcli.Capture` (L2) with isolated `WRK_HOME`/`WRK_DATE` overrides (no product binary). Short paths only; leave false for L3 e2e. Dual-mode on mega-tree root and nested roots: `sync/`, `gen-commit-msg/`, `projects-dep-graph/`, `reinstall-local-cli/`, `propagate-tags/` (+ `compose/`), `bash-integration/`, `bash-integration-install-messaging/`, `followup-cd/`, `tag-next/`, `status/main-repo-worktrees/`.
- **Request.StdinInput** — when non-empty, piped to wrk stdin before wait (mvd merge-back pattern).
- **--dry-run** — bool flag (no value); valid with **`--tag-next`**, **`--propagate-tags`**, **`--sync`**, and **primary composition** (`--done` / `--merge-back`, alone or with post modifiers `--sync` / `--tag-next` / `--push` / `--propagate-tags`). Bare `wrk --dry-run` (no host) → non-zero exit, stderr `wrk: --dry-run is only valid with --done, --merge-back, --tag-next, --propagate-tags, or --sync` (hosts exclude removed `--all-deps`). With primary + optional post flags: plan **all requested stages** (primary merge-back plan, cascade `would:` lines when applicable, then would-be sync / tag-next plan / push / propagate) with **zero mutations** and **no confirm prompts** (no `-y` required) — see `done-pipeline/dry-run/` and `merge-back-pipeline/dry-run/`. Standalone `--tag-next --dry-run` and `--propagate-tags --dry-run` unchanged.
- **wrk --task <desc>** / **wrk -t <desc>** — `-t` and `--task` are equivalent (like `-l,--list`); `hasArg` / `taskFlagSet` detect both forms. Event `args` record whichever form was passed (e.g. `["-t", "desc"]` when `-t` is used). Flag valid only in create mode (no `--done`/`--list`/`--bring`). Derives a sanitized slug from `<desc>` (lowercase, non-letter-digit → `-`, collapse, trim, truncate 64 runes soft cap). Appends slug after the date in both dir and branch names: `{basename}-{token}-{date}-{slug}[-N]` for dir, `{token}-{date}-{slug}[-N]` for branch (token = sanitize(currentBranch); no `/`). Empty `<desc>` or slug → non-zero exit. No metadata file stored — the slug is embedded in the name. **Name budget**: after soft cap, further shorten slug so path last component and branch stay ≤ **255 bytes** (reserve **3** for `-99` suffix → fit target 252); never silently chop basename/token — if prefix alone exceeds budget → clear non-zero error. Wrk-managed invariant remains `filepath.Base(path) == basename + "-" + branch`. Agent `${task}` / create-UX prompt always receives the **full original** task text (`taskDesc`), never the fitted slug alone.
- **Forgot `-t` / task-like positionals** — when the user omits `-t`/`--task`, create-mode positionals may still be **task-like**. Heuristic (any): contains ASCII whitespace (after trim non-empty); **or** length **> 120** bytes; **or** single path component would exceed **255** bytes / ENAMETOOLONG class. **Never** task-like when path-like (`/`, `\`, leading `~`/`./`/`../`), resolves to an **existing directory**, or single-arg **source resolve** (cwd path / projects basename) succeeds. **Two-arg** `wrk <dir> <arg2>` without `-t`: task-like arg2 → TTY (or `WRK_TASK_LIKE_CONFIRM=1` + piped stdin) prompts `Treat as --task? [Y/n]` (empty/`Y`/`y` → promote to `--task arg2` under default `WRK_HOME` naming; `n`/`N` → keep target-dir semantics); **non-TTY** without `-y` → hard error (looks like task / not a target directory) + hint with `-t`; **`-y`** auto-promotes without interactive prompt. Path-like / short tokens / existing dirs stay target-dir. When **`-t` already set**, second positional remains target-dir (no treat-as-task prompt). **One-arg** `wrk <arg1>`: task-like and not a resolvable source → same confirm/`-y`/non-TTY rules; promote creates from **cwd** with task; non-TTY error says not a source directory + `wrk -t '…'` hint. Existing source path/basename → normal create, no prompt.
- **WRK_TASK_LIKE_CONFIRM** — when set to `1` with piped `StdinInput`, bypasses TTY detection for the treat-as-task Y/n prompt (same escape-hatch pattern as `WRK_BASENAME_CONFIRM` / `WRK_SET_TASK_CONFIRM`). Prefer `Request.ExtraEnv` entry `WRK_TASK_LIKE_CONFIRM=1` over `UseScriptTTY` when both work.
- **wrk --set-task <desc>** — flag valid alone (mutually exclusive with all other flags). When run from inside a linked worktree (no `<dir>`), renames that worktree. When run as `wrk <dir> --set-task <desc>`, renames the worktree at `<dir>`. Requires a **linked** worktree whose **directory basename** is wrk-shaped (contains a `YYYY-MM-DD` date segment); fixed user paths / non-wrk dir names → clear error. Parses the worktree's branch name to extract `branchBase` and `date` (branch must match wrk pattern `{branchBase}-{YYYY-MM-DD}[-slug][-N]`; non-wrk branches → error). New path/branch names use **sanitized** token (`/` → `-`) so legacy slash branches migrate on rename. Same **name budget fit** as create (slug soft-cap 64 then fit ≤255 with `-99` reserve). If path or branch would collide, **suffix-walk** `-1`, `-2`, … (joint, keeping path/branch invariant) rather than hard-fail only. If nothing changes → `task unchanged` + trailing `\n`. Before `git worktree move`, checks stdout: TTY → warns (old→new path + branch) and prompts `Proceed? [Y/n]`; confirmation executes move + `git branch -m`. Non-TTY → non-zero exit `wrk: --set-task requires a terminal (tty)`. When run with `WRK_SET_TASK_CONFIRM=1` env → auto-confirms without TTY (test escape hatch). `<dir>` resolves to the effective working directory; empty `<dir>` (or absent) defaults to cwd.
- **Request.TaskDesc** — when set, task description passed to wrk (with `Request.TaskFlag`, default `--task`).
- **Request.TaskFlag** — task tests: CLI flag form for task description (`-t` or `--task`; default `--task` when `TaskDesc` is set).
- **Request.SetTaskDesc** — when set, tests pass `--set-task <desc>` to wrk; test assertions verify rename side effects.
- **Request.SetTaskEnv** — when set, appended to wrk's environment (e.g., `WRK_SET_TASK_CONFIRM=1` to auto-confirm rename in tests).
- **WRK data storage** — under `{WRK_HOME}`: `projects.json` (recorded main repos), `events.jsonl` (append-only event log), and optional `config.json` (user config; no `hooks/` directory); tests isolate at `{WorkRoot}/.wrk`.
- **create UX (window / terminal / agent)** — first-class create-mode UX from `$WRK_HOME/config.json` `create` section + one-shot CLI flags, implemented in-process via `computer-use/macos/space` (Mission Control Desktop) and `shell/iterm2` (iTerm2 open + follow-ups). Schema: `create.window.mode` (`"new"` only in v1; absent = window off), `create.terminal.mode` (`"new"` | `"reuse"` | `"smart"`; absent = terminal off), `create.agent` (`enabled` bool, `runner` default `grok-tty`, `prompt_template` default `/brainstorm ${task}`, `args` default `["--session-id-from-prompt","--no-submit","--open"]`). Legacy `create.interceptor` (and interceptor-only keys) are **silently ignored**; interceptor code/flags/env (`--interceptor`, `--no-interceptor`, `WRK_NO_INTERCEPTOR`, template argv engine) are **removed**.
- **create UX flags (create mode)** — `--new-window` (window on + implies terminal `new` unless another terminal mode flag sets reuse/smart), `--new-terminal` / `--reuse-terminal` / `--smart-terminal` (mutually exclusive), `--open-in-agent` / `--no-open-in-agent`, `--no-new-window`, `--no-new-terminal`. Conflicts: open/no-open together, new-window/no-new-window, terminal-on/no-new-terminal, **`--new-window` + `--no-new-terminal`**, multiple terminal mode flags. Effective merge for plain create (`wrk [dir]`, no second positional): load config create.* (ignore interceptor) → apply CLI → if window on && terminal off → terminal=`new` → validate. **Create-with-target-dir** (`wrk [dir] <target-dir>`, `spawnTarget` non-empty): **skip** config create.* entirely (silent; not an error); empty base + CLI flags only → same window-implies-terminal-new after flags → validate. One-shot UX flags remain valid with `<target-dir>`. Pipeline after native create + path stdout: window (`space.CreateAndActivate`) → terminal (`iterm2.OpenConfig` with optional agent follow-up) **or** agent-in-current-process (`agent-run run <args> --agent-runner=<runner> <prompt>`) → `--exec` → follow-up cd. Never double-run agent (terminal+agent ⇒ iterm follow-up only). Non-darwin window/terminal → clear platform error.
- **--no-config** — long-only top-level bool flag (no short alias). When set, wrk must **not read** and **not apply** `$WRK_HOME/config.json` for this invocation (scope: that file only; not `WRK_HOME` layout, env vars, events/projects). One-shot CLI flags still parse and apply. Create UX gate: `applyConfig = (spawnTarget == "") && !noConfig` (same silent skip as create-with-target-dir when `noConfig`). Missing config file → no-op; corrupt config with `--no-config` → never open/read → no parse error. **Hard mutual exclusion** with `--set-config`: non-zero exit, preferred stderr `wrk: --no-config is mutually exclusive with --set-config`. `--set-config` alone still reads/writes `config.json`.
- **wrk --set-config** — management mode for `config.json` (v1: `--create` section and recommended `--show`). Mutually exclusive with create, all other modes, and **`--no-config`**. Merge-only keys implied by flags; `--new-window` also persists `terminal.mode=new`; negatives clear/disable axes; preserve unknown top-level keys. No git required. Successful write: empty stdout preferred; `--show` prints JSON + trailing `
`. Event `command: "set-config"` when implemented.
- **Request.PathPrepend / ExtraEnv / InterceptorLog** — shared harness fields: `PathPrepend` prepends a bin dir to `PATH` (fake `agent-run` for create-ux; historically fake interceptor tools); `ExtraEnv` adds `KEY=VAL` for the wrk process (create-ux mocks: `WRK_SPACE_INVOKE_LOG`, `DOT_PKGS_SPACE_GOOS`, `KOOL_ITERM2_*`, `FAKE_AGENT_RUN_LOG` / `FAKE_AGENT_RUN_CWD`); `InterceptorLog` remains for any leaf that records fake argv logs (create-ux prefers WorkRoot paths via helpers).
- **Project record** — absolute path to the **main repo** (never a linked worktree path); deduplicated by normalized absolute `path`; fields `path`, `added_at` (ISO-8601 UTC), `source` (`auto` or `manual`; historical `scan` may exist in older files but is no longer written); re-adding is idempotent (no duplicate entries; first `source` wins).
- **Auto-record** — on **every** `wrk` invocation, after resolving the effective work directory: if dir missing → no record; if not inside git → no record; otherwise resolve to main repo via `worktree.ResolveMainRepo()` and append to `projects.json` with `source: "auto"` if not already present. Auto-record runs even when the command fails later; failed commands still append an event.
- **wrk --scan-git-repos [ROOT...]** — standalone **print-only** mode; mutually exclusive with other modes (`--projects`, `--add`, `--list`, create, etc.). Discovers git directories under each root via `scan_repo.Scan` (default product CacheRoot unless overridden), filters **`RepoTypeMain` only** by default (skips linked worktrees unless `--include-worktrees`), and **never reads or writes `projects.json`** (no `RecordProject`, no `source: "scan"`). **Roots**: remaining positionals after flags; when no roots are given, default root is **`$HOME` (`~`)** if it is a directory (not `~/Projects`); empty/unresolvable/non-directory/side-effect-only home → clear error mentioning home or `~`. **Two-base cache + path filter (P5)**: each CLI root under `$HOME` (including HOME itself) maps to universe **`home`** and shares product cache files under `$HOME/.cache/git-repo-scan/home/` (`repos.json` durable index, `walk.jsonl` + `walk.cursor.json`, `meta.json`); roots **outside** home map to universe **`root`** (`…/root/…`). Scan still passes the user-provided roots so discovery walks those paths, while the home-universe index can retain siblings outside a subpath root (e.g. bare `~` then later `~/Projects` reuses the same `home/repos.json`; emit is **filtered** to paths under the CLI roots only — scanning `~/Projects` does not print `~/other-main`). Empty `CacheRoot` uses product default `$HOME/.cache/git-repo-scan` (library also maintains per-dir `mirror/` entries and adaptive walk-log consume budgets — see `scan_repo` tests). **Streaming / OnRepo**: for each valid discovery under the filter, print immediately via `scan_repo.Options.OnRepo` — **discovery order** (CLI root order + walk order), **not** sorted after the full scan. **Stdout**: each valid absolute path on its own line as soon as it is found (trailing `\n` after last); same path at most once per run. Exit 0 on success; partial root errors may warn on stderr. **Ctrl-C / SIGINT / SIGTERM**: cancel the scan context, keep scan disk cache progress when applicable, **leave `projects.json` unchanged**, write stderr `warning: scan interrupted; cache progress may be saved (projects.json unchanged)` (must include `warning:` plus interrupt intent and progress/saved/cache wording), and exit **130** via silent `ExitCodeError{Code: 130}` (no second hard-error body). **`--no-cache`**: bool flag valid **only** with `--scan-git-repos`; passes `NoCache: true` to `scan_repo.Scan` (no cache read/write); bare `wrk --no-cache` → non-zero, stderr `wrk: --no-cache is only valid with --scan-git-repos`. **Debug**: `scan_repo.Options.Debug = verbose || envTruthy(WRK_SCAN_DEBUG)` where truthy is `1`/`true`/`yes` (case-insensitive); pass `Options.Stderr = os.Stderr`. When Debug is on, stderr includes greppable phase-level `scan:` lines (`mode=warm|cold`, cache root, serve/refresh timing) plus wrk-side **two-base** lines `scan: cache_base=home|root filter=<abs-root>` per CLI root and optional summary `scan: printed=N`; when off, zero `scan:` markers from the scan path. Help documents `--scan-git-repos`, `--no-cache`, print-only / no projects.json mutation, and default root `~`. Event `command: "scan-git-repos"`.
- **WRK_PROJECTS_PERF_LOG** — when set to a file path, `wrk --projects` appends JSONL latency events (`run_start`, `project_start`, `phase`, `worktree_status`, `phase_total`, `project_end`, `run_end`) without changing stdout/stderr; zero overhead when unset.
- **Request.ProjectsPerfLog** — perf-profile tests: path written to `WRK_PROJECTS_PERF_LOG`.
- **wrk --projects** — standalone mode; mutually exclusive with all other modes; prints one **detailed status block** per recorded main repo, sorted lexicographically by absolute path, with blank lines between blocks. **Never aborts** the run due to per-project or per-worktree git failures (exit 0 unless `projects.json` is unreadable); errors surface inline in stdout blocks; stderr stays empty for these cases (unless `-v` is set). **Default (no `--fetch`)**: skip `git fetch`; `Remote:` uses `git.CompareBranches` against local upstream tracking refs. **With `--fetch`**: run scoped upstream fetch (`gitFetchUpstreamQuietNoOptionalLocks`) before `Remote:` comparison per project. **Healthy main repo** blocks include absolute `Dir`, `Branch`, `Commit`, `Status` (same fields as `--status` for the main repo), plus `Remote:` (brief upstream sync summary via `git.CompareBranches`: `identical`, `needs push(+N commit(s))`, `needs pull(N commit(s) behind)`, `diverged(N commit(s))`, `(no upstream)` when the branch has no upstream, or `error: ...` inline when fetch/compare fails), and `Worktrees:` (four spaces after colon, aligned with other fields) with composable summary segments: `N total` and `M dirty` always; `K error` only when K > 0 (alive linked worktree path exists but `git status` fails); `P prune` only when P > 0 (registered in `git worktree list` but checkout directory missing per `worktree.IsDead`). After the `Worktrees:` line, each broken (alive, git-fails) worktree emits `  <absolute-path>  error: <full git stderr message>` (two-space indent); no per-path lines for prunable/dead worktrees. **Broken main repo** blocks omit Branch, Commit, Remote, and Worktrees entirely — only `Dir:` and `Status:       error: <full git stderr message>`. When stdout is a TTY or `--color` is set, highlights attention-worthy **value** portions only: red for the word `dirty`, each dirty count segment with N > 0, `Remote: diverged(...)`, `N dirty` when N > 0, `K error` when K > 0, broken-main `Status: error: ...` value, and per-worktree `error: ...` detail values; grey (`#90`) for dirty count segments with N = 0; orange (`#33`) for `needs push(...)` and `needs pull(...)`; separators `(`, `, `, `)` in dirty status lines stay uncolored; `clean`/`identical`/no-upstream/zero-dirty stay plain (no green on `--projects`). No `<dir>` required; exit 0 when empty (no output). Note: `needs merge back(+N commit(s))` and `needs fast forward(+N commit(s))` apply only to `--status` `Master:` (not `Remote:`).
- **--fetch** — bool flag (no value); valid ONLY with `--projects` or `--status`. Default false (no network fetch). Bare `wrk --fetch`, or `--fetch` with any other mode (`--list`/`--done`/`--bring`/no-args create/etc.) → non-zero exit, stderr `wrk: --fetch is only valid with --projects or --status`. With `--projects` or `--status` from **main repo checkout cwd**: run scoped upstream fetch before `Remote:` comparison. From **linked worktree cwd** with `--status --fetch`: silently ignored (no fetch, no error, no `Remote:` added). Combinable with `--color`. Recorded in `events.jsonl` `args` when passed.
- **-v / --verbose** — global bool flag; valid with **any** wrk mode; does not change mode selection or stdout content. When set, log **major** git subprocesses (mutating/network: `worktree add`/`remove`/`move`, `fetch` when executed, `checkout`, `branch` `-D`/`-m`/`-b`, `merge`, `rebase`, `stash`) to **stderr** as one line per invocation before the command runs: `[YYYY-MM-DD HH:MM:SS] $ git <args...>` (local timezone, format `2006-01-02 15:04:05`; include `-C <dir>` when used). **Create and `--bring` external worktree create**: stream `git worktree add` subprocess stdout+stderr to process stderr as the command runs via `runGitWorktreeAdd` (after the pre-command log line; e.g. `Preparing worktree (new branch '…')`, `HEAD is now at <hash>`); success and failure paths; always uses `-b` new-branch (`--no-checkout` reuse path removed). **When `go mod tidy` runs** (bring after replace; not under `--no-dep`): log pre-command line on stderr `[YYYY-MM-DD HH:MM:SS] $ go -C <abs-module-dir> mod tidy` and stream tidy child stdout+stderr to process stderr. Without `-v`: no tidy pre-line; tidy silent. **Not logged**: read-only introspection (`rev-parse`, `log`, `status`, `diff`, `merge-base`, `rev-list --count`, `worktree list`, `show-toplevel`, `config`, etc.) and non-git commands. When `-v` is off: zero stderr logging overhead (create mode still captures `worktree add` via `CombinedOutput` silently). **With `--scan-git-repos`**: `-v`/`--verbose` also sets `scan_repo.Options.Debug=true` (same as truthy `WRK_SCAN_DEBUG`), so stderr may include greppable `scan:` phase lines in addition to any major-git logs. Recorded in `events.jsonl` `args` when passed.
- **--color** — bool flag (no value); valid with any mode; forces ANSI coloring on `--projects` and `--status` output even when stdout is a pipe (doctest-safe); no-op on other modes today (e.g. `--list --color` unchanged).
- **Stdout trailing newline** — all wrk modes that print non-empty stdout end with `\n` after the last content line (shell prompt stays on its own line). Empty stdout has no bytes.
- **Stderr hard-error trailing newline** — when `main` receives a non-nil error that is **not** `wrkcli.ExitCodeError`, it writes `err.Error()` to stderr and ensures the last byte is `\n` (prefer `Fprintln` / single append; do not double-append if the string already ends with `\n`), then exits `1`. `ExitCodeError` stays silent (exit with `Code` only). Shell prompt must not glue to the error line.
- **Stdout assertions** — doctest leaves use `assert.Output` with `version: 2` full-match templates only (no `<contains>` for stdout). Multi-block stdout (e.g. `--status` scan blocks, `--projects` project blocks) is asserted with one v2 template covering the entire stdout; blocks are joined with `\n\n`. Stderr error messages continue to use `<contains>` partial match (except sealed exact-body cases that also pin trailing `\n`).
- **wrk --projects streaming** — stdout must flush each lex-ordered project block as soon as that project's gather completes (not after all projects finish). `output-streaming/fast-before-slow-gather` probes pipe timing: first bytes are the fast `aaa` block while the slow `zzz` project (12 worktrees) is still gathering.
- **Run profile labels** — **`e2e`** marks true process-boundary leaves only (TTY/`script`, web probe, bash wrapper install, multi-invoke complete, fake-shell fallback, create-ux mock-env under Parallel, …) — sparse (~10–15% of leaves after harden). **`slow`** / **`heavy`** are cost. **L2 mass** (unlabeled discovery): library trees + `Request.InProcess` → `wrkcli.Capture` for CLI (short paths, dry-run plans, apply/status/projects when Capture-safe). Discovery runs the L2 mass; full process-boundary: `doctest test --label e2e ./tests`.
- **wrk --add `<dir>`** — standalone mode; `--add` consumes the next argument as `<dir>`; validates dir exists + is git; resolves to main repo root; records with `source: "manual"` (idempotent); mutually exclusive with other modes; prints resolved main repo path on stdout (single line) on success.
- **wrk --rm `<dir>`** — standalone mode; `--rm <dir>` (no `--remove` alias); `--rm` consumes the next argument as `<dir>`; mutually exclusive with all other modes; requires non-empty path (`wrk: --rm requires a path argument`). Help text: `--rm <dir>  remove a recorded main repository path`. Resolves target: `filepath.Abs` + `storage.NormalizePath`; if path exists and is inside a git work tree → resolve to main repo via `worktree.ShowToplevel` + `worktree.ResolveMainRepo` (same as `--add`); if path does not exist → use normalized absolute path as-is (stale/moved entries). **Success (entry removed)**: exit 0; stdout = removed main-repo absolute path (single line, trimmed). **Idempotent (not in projects.json)**: exit 0; empty stdout; no error. Does not delete worktrees, git repos, or events.jsonl history. Appends event `command: "rm"`, `args: ["--rm", "<path-arg>"]`, `exit_code: 0`. Auto-record still runs before remove.
- **wrk --where `<basename>`** — standalone read-only lookup mode; `--where` consumes the next argument as a **basename only** (no path separators, not absolute); loads `{WRK_HOME}/projects.json` via `storage.FindProjectsByBasename(wrkHome, basename)` matching `filepath.Base(NormalizePath(p.Path)) == basename`; **does not** stat cwd, `filepath.Abs(name)`, or resolve paths on disk (unlike create-mode basename fallback). **0 matches** → non-zero exit, stderr no-match message, empty stdout. **1 match** → exit 0, stdout = one full absolute path + trailing `\n`. **2+ matches** → exit 0, stdout = all matching full paths sorted lexicographically, one per line, trailing `\n` after last line (no TTY prompt). **Empty/missing arg** (`wrk --where`) → non-zero exit, stderr `wrk: --where requires a path argument`. **Non-basename input** (contains `/` or `\`, or absolute path) → non-zero exit, basename-only rejection. **Mutually exclusive** with all other modes (`--status`, `--list`, `--projects`, create, etc.). **Extra positionals** → non-zero exit, `wrk: unexpected arguments`. No writes (no git ops, no worktree creation, no `projects.json` mutation). Appends event `command: "where"`, `args: ["--where", "<basename>"]`. Auto-record still runs on invocation.
- **RemoveProject** — storage API `RemoveProject(wrkHome, path string) (removed bool, err error)` deletes the `projects.json` entry matching normalized absolute `path`; returns whether an entry was removed.
- **events.jsonl** — one JSON object per line appended on every wrk invocation (success or failure): `ts` (ISO-8601 UTC), `command` (mode: `dashboard`, `create`, `done`, `list`, `status`, `bring`, `merge-back`, `set-task`, `repos`, `projects`, `add`, `rm`, `where`, `cd`, `set-config`, `scan-git-repos`, …), `work_dir` (resolved effective cwd; for `--cd` the resolved absolute target path), `main_repo` (resolved main repo or empty), `args` (remaining CLI flag args, not positionals), `exit_code`.
- **wrk --cd** — standalone mode: `Bool("--cd")` plus **exactly one** path positional. Forms: `wrk --cd <path|basename>` and `wrk <path|basename> --cd`. Resolves path via `resolveDirArg(..., allowBasenameFallback=true, ...)` (local dir under cwd wins; else `projects.json` basename lookup; ambiguous non-TTY lists candidates). **Git not required** for the target. **Branch A (in-place)**: when `WRK_FOLLOWUP_FILE` is non-empty, write `cd /absolute/path\n` via `writeFollowupCD(false, abs)` unconditionally (no create home-gate / done cwd-gate), **empty stdout**, exit 0, do **not** launch a shell. **Branch B (fallback)**: channel unset/empty → stderr warning containing `wrk --bash-integration --install`, stdout = absolute path + trailing `\n`, then `shell/interactive.LoginInteractive(abs, filepath.Base(abs), optional extraEnv...)` and propagate shell exit code. Mutually exclusive with create / `--done` / `--merge-back` / `--list` / `--status` / `--repos` / `--projects` / `--add` / `--rm` / `--where` / `--bring` / `--task` / `--set-task` / spawn target / **`--no-cd`**. Missing path → `wrk: --cd requires a path argument`. Extra positionals → `wrk: unexpected arguments`. Event `command: "cd"`. Doctest harness: `Request.UseFollowupEnv` + `FollowupFile` for Branch A; `installFakeBash` (`FakeShellDir` / `FakeShellLog` / `SHELL`) so Branch B never hangs CI.
- **Request.FollowupFile / UseFollowupEnv** — when `UseFollowupEnv` is true, root `Run` truncates `FollowupFile` and exports `WRK_FOLLOWUP_FILE` (in-place channel for `--cd` and related tests).
- **Request.FakeShellDir / FakeShellLog / FakeShellExit / ShellEnv** — fallback `--cd` harness: prepend fake `bash` on PATH, set `SHELL`, `WRK_FAKE_SHELL_LOG`, and optional non-zero `WRK_FAKE_SHELL_EXIT` so `LoginInteractive` cannot hang.
- **wrk --exec** — long-only cut flag (no `-e`); less-flags `Cut("--exec", &execArgs)`. On seeing `--exec`: if no tokens after → **error** (requires a command); else copy all subsequent tokens into `execArgs` and **stop parsing** (tokens never treated as wrk flags — e.g. `wrk --exec echo --task` runs `echo --task`). Reject equals form `--exec=value`. Absent `--exec` → no post-mode command. After a **successful** allowed mode, run `exec.Command(execArgs[0], execArgs[1:]...)` with `cmd.Dir` set to the mode target absolute directory; inherit stdin/stdout/stderr; non-zero child exit → `ExitCodeError` (same pattern as `--cd` shell / agent-run). **Allowed modes & `cmd.Dir`**: **create** (native) → newly created worktree; **`--cd`** → resolved jump directory; **`--bring`** → external dep worktree (including after soft SKIP); **`--set-task`** → renamed worktree (post-move); **`--done`** → main repo (`MergeBack` `TargetPath`, never the removed worktree). **`--done`**: exec only after successful done (not aborted / failed confirm). **Reject `--exec` with**: `--list`, `--status`, `--repos`, `--projects`, `--add`, `--rm`, `--where`, `--merge-back`, set-config, skill, bash-integration, and other non-allowed modes. Mode path/message lines still print **before** the child command's stdout (create/bring/set-task path; done `worktree removed:…`; in-place `--cd` keeps empty mode stdout so only child output appears). Follow-up shell cd rules unchanged (create home-gate; done/set-task cwd-missing); exec is a child process and does not replace follow-up writes.
- **Request.Args with `--exec`** — leaves pass `--exec` and command tokens as trailing `req.Args` (or after `SetTaskDesc`/`TargetDir` assembly via `buildWrkCLIArgs`); no separate Request field for exec argv.
- **wrk --web** — standalone long-running mode: bind `127.0.0.1` only; optional `--port PORT` (when omitted, pick a free port starting at 8080); serve embedded standalone workflow HTML at `GET`/`HEAD` `/` (`Content-Type: text/html`; body includes identifiable markers: `task`, `Main`, `Remote`, `wrk`, and preferably `worktree`/`changes`/`sync`/`rebase`/`agent-run`/`tag`/`push`); mount existing `wrkserver` under `/api/wrk` (e.g. `GET /api/wrk/projects` → JSON `{"projects":[...]}` with array never null). On successful start print exactly one user-facing stdout line `http://127.0.0.1:<port>/` ending with trailing `\n`; block until SIGINT/SIGTERM. **Mutually exclusive** with all other modes/flags that select behavior (`--list`, `--status`, create, `--done`, `--projects`, …) — prefer `wrk: --web is mutually exclusive with other modes`. **No positionals** (`wrk --web some-dir` → `wrk: unexpected arguments`). **`--port` without `--web`** → non-zero, `wrk: --port is only valid with --web`. Help (`wrk -h`) documents `--web` and `--port`. Event `command: "web"` when recorded (optional for long-lived servers). No new on-disk storage; resolves `WRK_HOME` like the CLI. Page + serve helpers live under `go-pkgs/wrkcli/` (not a separate monorepo-root React app for this iteration).
- **Request.WebProbe / WebPath** — when `WebProbe` is true, root `Run` starts wrk as a long-lived subprocess (isolated `WRK_HOME`), waits up to ~10s for a `http://127.0.0.1:<port>/` listen URL on stdout, HTTP GETs `WebPath` (default `/`), fills `Response.HTTPStatus` + `Response.HTTPBody` (+ `Stdout`/`Stderr`), then SIGTERM (SIGKILL after 2s). Used by `web/serves-page`, `web/mockup-repo-view`, and `web/api-projects-empty`. Error/help leaves leave `WebProbe` false and use normal run-to-completion.
- **Request.SecondRepo** — projects tests: second main repo path for multi-project list assertions.
- **Basename fallback** — shared `resolveDirArg` core (`filepath.Abs` → `stat` → optional `projects.json` lookup via `isBasename` / `resolveBasenameFromProjects` / `pickAmbiguousBasename`). When the user-supplied directory argument is a basename (no path separator, not absolute), `stat(filepath.Abs(<dir>))` fails, and `stat(filepath.Join(cwd, <dir>))` also fails: load `projects.json`, collect entries where `filepath.Base(project.path) == <dir>`. **0** → unchanged `wrk: <candidate> does not exist`; **1** → use that project's `path` as the resolved absolute path; **2+** → TTY prints numbered list (candidates sorted lexicographically by absolute path) and prompts `Select [1-N]:`; non-TTY errors listing all candidates. **Skipped** when: `./<dir>` exists in cwd as a **directory** (even non-git — use cwd path, existing git error); or `<dir>` contains a path separator. **Cwd file collision** (new): when `filepath.Join(cwd, <dir>)` exists and is a **regular file** (not a directory), do not proceed to git-repo resolution; instead load `projects.json` and emit guided stderr. **1** registered match → multi-line stderr: `wrk: <abs-cwd-file> exists and is a file`, `wrk: "<basename>" matches registered project(s):`, one indented project path, `wrk: use \`wrk <concrete-saved-path> <reconstructed-args>\` instead` (hint preserves user flags/args such as `-t`, `--status`, `--bring`, spawn target). **2+** matches → same shape listing all project paths (lex order) and hint `wrk: use \`wrk <full-path> <reconstructed-args>\` instead` (literal `<full-path>` placeholder). **0** matches → single line `wrk: <abs-cwd-file> exists and is a file` only (no registry block, no hint). Exit non-zero; stdout empty; no worktree created. Directory blocking behavior is unchanged. **Enabled** for: create-mode first positional `<dir>` (`wrk <dir>`, `wrk <dir> <target-dir>`) via `resolveSourceWorkDir` with `allowBasenameFallback=createMode`; `--bring <dir>` via bring mode with `allowBasenameFallback=true`; and `wrk <dir> --status` via `resolveSourceWorkDir` with `allowBasenameFallback=status`. **Not enabled** for other modes (`--list`, `--done`, `--projects`, `--add`, `--set-task`, `--merge-back`) — positional basename in those modes still skips lookup. `--where` unchanged (no cwd stat).
- **WRK_BASENAME_CONFIRM** — when set with piped `StdinInput`, bypasses TTY detection for ambiguous-basename prompt tests (same escape hatch pattern as `WRK_SET_TASK_CONFIRM`).
- **Request.BasenameEnv** — basename-fallback tests: extra env var appended when running wrk (e.g. `WRK_BASENAME_CONFIRM=1`).
- **Request.SelectedSavedRepo** — basename-fallback tty-select: absolute path of the saved project chosen via stdin index.
- **Request.FakeHome** — git-lfs-hook tests: temp home directory holding `$HOME/.local/bin/git-lfs` shim.
- **Request.UseMinimalPath** — when true, wrk runs with `PATH=/usr/bin:/bin` and `HOME={FakeHome}`; git-lfs hook failure is expected (exit 1).

## Tree Overview

```
wrk tests
├── create-worktree/              # cwd is a git checkout (success path)
│   ├── main-checkout/            # cwd is the main repo checkout
│   │   ├── basic-create/         # first wrk from main
│   │   ├── sequence-increment/   # second wrk increments -N suffix (always new branch)
│   │   ├── branch-collision/     # branch ref blocks → joint path+branch -1 via -b
│   │   └── slash-branch/         # / sanitized in path token AND branch name
│   ├── from-linked-worktree/     # cwd is linked worktree; basename from main repo
│   ├── from-git-subpath/         # cwd is nested subdir inside checkout; basename from repo root
│   ├── detached-head/            # cwd on detached HEAD → 7-char hash token
│   └── git-lfs-hooks/            # LFS post-checkout hook requires git-lfs on PATH
│       ├── minimal-path-succeeds/  # stripped PATH; git-lfs in $HOME/.local/bin → create fails
│       └── from-other-cwd/         # wrk <repo> from foreign cwd + stripped PATH → create fails
├── set-config/                   # wrk --set-config management for create UX config.json
│   ├── write/                    # --set-config --create mutators
│   │   ├── full-on/              # window+terminal+agent defaults
│   │   ├── new-window-implies-terminal/ # --new-window alone also writes terminal.mode=new
│   │   ├── merge/
│   │   │   ├── terminal-then-agent/     # sequential writes preserve both
│   │   │   └── preserve-extra-key/      # top-level extra:1 preserved
│   │   └── negatives/
│   │       ├── no-open-in-agent/        # agent.enabled=false
│   │       └── no-new-window/           # window cleared/off
│   ├── show/
│   │   └── prints-json/          # --set-config --show → JSON stdout
│   ├── help/                     # nested level-specific -h/--help (RED until implemented)
│   │   ├── set-config/           # dispatcher: --set-config --help|-h
│   │   │   ├── long/             # --help
│   │   │   └── short/            # -h
│   │   ├── create/               # dedicated create usage
│   │   │   ├── long/             # --create --help
│   │   │   ├── short/            # --create -h
│   │   │   ├── help-before-create/ # --help --create order
│   │   │   └── with-ux-flag/     # --create --new-window --help; no write
│   │   └── show/                 # show usage
│   │       ├── long/             # --show --help
│   │       └── short/            # --show -h
│   └── mutual-exclusion/
│       ├── with-list/            # --set-config … --list → non-zero
│       ├── with-create-dir/      # wrk <dir> --set-config … → non-zero; no worktree
│       └── with-no-config/       # --no-config --set-config … → non-zero; no write
├── create-ux/                    # create-mode window/terminal/agent UX (config + flags)
│   ├── bare/
│   │   └── empty-config/         # native create only; mocks silent
│   ├── pipeline/
│   │   ├── flags/                # CLI-only effective UX
│   │   │   ├── new-window-only/  # space + iterm ForceNew; no agent
│   │   │   ├── new-terminal/     # iterm ForceNew; no space
│   │   │   ├── reuse-terminal/   # ModeReuseCurrent
│   │   │   ├── smart-terminal/   # ModeSmart
│   │   │   ├── open-in-agent-only/ # agent-run in current process
│   │   │   ├── terminal-plus-agent/ # iterm follow-up only; outer agent not exec'd
│   │   │   ├── full-pipeline/    # window + terminal + agent follow-up
│   │   │   └── with-exec/        # UX then --exec pwd in worktree
│   │   └── config/               # config defaults ± --no-* override
│   │       ├── defaults-match-flags/
│   │       └── no-open-in-agent-override/
│   ├── errors/
│   │   ├── new-window-no-terminal/
│   │   ├── mutual-terminal-flags/
│   │   └── non-darwin-window/
│   ├── interceptor-ignored/
│   │   └── native-create/        # leftover create.interceptor ignored
│   ├── agent-quoting/
│   │   ├── adversarial-task-quotes/      # argv-safe prompt for agent-in-process
│   │   └── terminal-followup-quotes/     # shell-safe prompt in iterm follow-up
│   ├── agent-full-task/                  # agent always gets full taskDesc
│   │   └── name-budget-trim/             # long basename+task fitted; agent prompt = full text
│   ├── target-dir-config-skipped/ # SpawnDir set → config create.* not applied
│   │   ├── config-ignored/        # full config; no CLI UX → no space/iterm/agent
│   │   ├── flags-still-apply/     # empty config; full CLI UX flags still run
│   │   └── flag-only-no-config-agent/ # config agent on; --new-terminal only → no agent
│   └── no-config/                 # --no-config skips $WRK_HOME/config.json on plain create
│       ├── config-ignored/        # full config + --no-config; no UX flags → mocks silent
│       ├── flags-still-apply/     # full config + --no-config + full CLI UX → flags run
│       └── corrupt-ignored/       # corrupt config.json + --no-config → no parse error
├── bring/                        # wrk --bring sole external-dep worktree mode (best-effort replace)
│   ├── basic/                    # require + --bring → external wt + replace (no SKIP)
│   ├── branch-collision-suffix/  # preferred branch taken → path+branch -1
│   ├── gitignore-already/        # /external already in .gitignore → no duplicate
│   ├── not-a-dependency/         # go.mod but no require → worktree+gitignore + SKIP not a dependency
│   ├── not-go-module/            # dep git without go.mod → worktree+gitignore + SKIP not a go module
│   ├── consumer-no-modules/      # consumer zero go.mod → worktree+gitignore + SKIP consumer has no Go modules
│   ├── dep-sub-module/           # dep module in subdir → replace => <external>/sub
│   ├── dep-multi-sub-module/     # dep multi sub-modules; match correct one → replace => <external>/b
│   ├── consumer-sub-module/      # consumer module in subdir → scan + replace
│   ├── consumer-multi-module/    # two consumer sub-modules both require dep → replace in both
│   ├── consumer-multi-module-selective/ # only one consumer sub-module requires dep
│   ├── both-sub-modules/         # both consumer + dep modules in subdirs
│   ├── cwd-in-sub-module/        # cwd inside consumer sub-module dir
│   ├── external-wt-from-linked-consumer/ # --bring from linked consumer wt → owned by dep main
│   ├── basename-fallback/        # wrk --bring <basename> → projects.json lookup
│   │   ├── single-match/basic/
│   │   ├── cwd-exists/no-fallback/
│   │   ├── path-with-separator/no-fallback/
│   │   ├── no-match/error/
│   │   └── ambiguous/            # tty-select + non-tty (P1 resolve-once; no will bring)
│   ├── multi/                    # repeatable --bring p1 --bring p2
│   │   ├── two-success/          # two abs deps → two external paths
│   │   ├── basename-two/         # two registered basenames
│   │   ├── preflight-ambiguous/  # P2 multi Select one-by-one + will bring plan
│   │   │   ├── two-basenames-select/
│   │   │   └── duplicate-after-select/
│   │   └── reject/               # exec / positionals / exact duplicates
│   ├── exec-after-skip/          # SKIP not-a-dependency + --exec pwd runs in external
│   ├── reuse-same-repo/          # Policy A: reuse live external WT of same depMain
│   │   ├── second-bring/         # second --bring → same path; reuse warning; no -1
│   │   ├── multi-external/       # two external WTs same depMain → lex-smallest + multi warn
│   │   └── two-different-deps/   # dep1 then dep2 → no false reuse
│   ├── no-dep/                   # --bring --no-dep: worktree only (skip replace+tidy)
│   │   ├── skips-replace-tidy/
│   │   ├── with-verbose-no-tidy-log/
│   │   ├── not-a-dep-still-works/
│   │   └── invalid/              # --no-dep only valid with --bring
│   │       ├── bare/
│   │       └── with-list/
│   ├── verbose/                  # -v with --bring
│   ├── not-git-cwd/              # plain non-git cwd soft path
│   ├── help-mentions-no-dep/     # wrk -h documents --no-dep and -v tidy
│   └── removed-flags/            # hard removal of legacy --dep / --all-deps (RED pre-implement)
│       ├── unknown-dep/          # wrk --dep → unknown flag
│       ├── unknown-all-deps/     # wrk --all-deps → unknown flag
│       ├── dry-run-host-list-no-all-deps/ # bare --dry-run hosts exclude --all-deps
│       └── help-no-removed-flags/ # -h has --bring; no --dep / --all-deps mode lines
├── done/                         # wrk --done merge-back --rm from linked worktree
│   ├── already-included/         # branch already merged into main; remove only
│   ├── ahead-confirm/            # ahead + --confirm-from-stdin + Enter
│   ├── ahead-decline/            # ahead + --confirm-from-stdin + decline
│   ├── ahead-non-tty/            # ahead without confirm flag (non-interactive)
│   ├── dirty/                    # uncommitted changes in worktree
│   ├── not-linked/               # cwd is main repo (not linked worktree)
│   ├── from-subpath/             # cwd nested inside linked wt; uses checkout root
│   ├── external-cascade/         # cascade removes external/* wt; guard blocks parent (names go.mod + directive)
│   ├── cascade-merge-base/       # cascade must remove dep wt, not crash "failed to find merge base" (dep branch shares no history with consumer main)
│   ├── cascade-force-removal/    # bug: non-TTY ahead cascade must error, not force-remove
│   │   ├── ahead-non-tty-errors/
│   │   └── ahead-non-tty-with-y-still-errors/
│   ├── cascade-non-tty-rejects-with-confirm-from-stdin/ # option A: --confirm-from-stdin cannot confirm cascade on non-TTY
│   ├── cascade-dep-merge-back/   # regression: ahead dep + --confirm-from-stdin on non-TTY → pre-flight error (no cascade merge)
│   ├── cascade-non-external-linked/ # manual deps/foo linked wt (not under external/) → cascade removes it, consumer merge-back exit 0
│   ├── cascade-external-and-deps/ # external/* + deps/foo both linked → cascade removes both, consumer merge-back exit 0
│   ├── local-replace-blocks/     # extra-repo fs replace (non-existent ./external/foo) → guard blocks + names go.mod + directive
│   ├── intra-replace-warns/      # intra-repo fs replace (./submod, same toplevel) → WARN + proceed (default, exit 0)
│   ├── intra-replace-cross-worktree/ # abs replace to sibling checkout; wrk --done <wt> from outside → extra-repo block
│   ├── intra-replace-strict-blocks/ # intra-repo replace + --no-in-module-replace → block + names go.mod + directive
│   ├── no-in-module-replace-without-done/ # --no-in-module-replace without --done → error
│   ├── no-go-mod/                # linked wt whose checkout has no go.mod → --done merge-back succeeds (guard is no-op)
│   ├── not-linked-no-go-mod/     # main repo without go.mod → "not a linked worktree" (go.mod check must not mask it)
│   └── sub-module-replace-blocks/ # main go.mod clean but sub/go.mod has local replace → guard blocks + names go.mod + directive
├── done-sync/                    # wrk --done --sync post-success composition
│   ├── basic-propagate/          # wtA done (ahead); wtB behind → pass2 distribute after remove
│   ├── no-other-wt/              # only wtA; after done, zero-summary sync
│   ├── aborted-no-sync/          # decline confirm → no synced: line; wt remains
│   └── flag-order-sync-first/    # --sync --done same as --done --sync
├── done-push/                    # wrk --done --push branch push after successful done
│   ├── pushes-main/              # bare origin; after done, origin/main == main HEAD
│   ├── flag-order-push-first/    # --push --done -y same as --done -y --push
│   └── no-remote/                # no origin → non-zero; clear remote error
├── push/                         # bare wrk --push (standalone branch push; option R)
│   ├── pushes-branch/            # main + origin → pushed main → origin/main
│   ├── dry-run/                  # --push --dry-run → would: git push; no mutation
│   ├── no-remote/                # no origin → non-zero
│   ├── from-linked-worktree/     # option R: push worktree branch tip
│   ├── exclusive-with-list/      # --push --list mutually exclusive
│   ├── json-rejected/            # --push --json alone still invalid
│   └── events/                   # events.jsonl command=push
├── done-pipeline/                # done post-pipeline sync → tag-next → push → propagate-tags
│   ├── tag-next-local/           # --done -y --tag-next → local v0.0.2 at main HEAD
│   ├── tag-next-push/            # --done -y --tag-next --push → local+origin tags + branch
│   ├── sync-tag-next/            # --done -y --sync --tag-next → sync then local tags
│   ├── sync-tag-next-push/       # full combo ordered stdout + side effects
│   ├── tag-next-propagate/       # P7: --done -y --tag-next --propagate-tags → tag + consumer bump
│   ├── propagate-existing/       # P7: --done -y --propagate-tags on existing source tags
│   ├── aborted-full-flags/       # decline confirm → no sync/tag/push/propagate
│   ├── flag-order-full/          # --push --tag-next --sync --done -y same as full combo
│   └── dry-run/                  # composition dry-run (zero mutations, no prompts)
│       ├── alone/                # --done --dry-run → primary plan only
│       ├── full-combo/           # --done --sync --tag-next --push --dry-run full plan
│       ├── tag-next-propagate/   # P7: dry-run includes would-propagate (planned next tag)
│       ├── cascade-external/     # cascade would: line; external still present
│       └── no-prompt-without-y/  # dry-run without -y does not require confirm TTY
├── done-output/                  # P2 UX: phase banners + structured cascade Error:
│   ├── dry-run-phases/           # ==> cascade then ==> own (structure, no ANSI)
│   │   ├── with-cascade/         # nested external: headers + would: under cascade
│   │   └── zero-cascade/         # no nested WTs: still prints ==> cascade
│   └── cascade-failure/          # real cascade MergeBack failure framing
│       └── structured-error/     # Error: + path; not bare rebase conflict: alone
├── merge-back-sync/              # wrk --merge-back --sync post-success composition
│   ├── basic-propagate/          # merge-back keeps wtA; wtB gets main tip
│   └── source-wt-stays/          # wtA still on disk after success; no worktree removed
├── merge-back-pipeline/          # merge-back post-pipeline sync → tag-next → push → propagate-tags
│   ├── tag-next-local/           # --merge-back -y --tag-next → local v0.0.2; wt remains
│   ├── tag-next-push/            # --merge-back -y --tag-next --push → branch+tags; wt remains
│   ├── sync-tag-next/            # --merge-back -y --sync --tag-next → sync then local tags
│   ├── sync-tag-next-push/       # full combo ordered stdout + side effects; wt remains
│   ├── tag-next-propagate/       # P7: --merge-back -y --tag-next --propagate-tags; wt kept
│   ├── aborted-full-flags/       # decline confirm → no sync/tag/push; wt remains
│   ├── flag-order-full/          # --push --tag-next --sync --merge-back -y same as full combo
│   └── dry-run/                  # composition dry-run (zero mutations, no prompts)
│       └── tag-next/             # --merge-back --tag-next --dry-run → keep wt; tag planned
├── done-compose/                 # flag matrix: primary + pre gen-commit (P2) + post modifiers
│   ├── allow/                    # flag layer accepts composition (not exclusive with primary)
│   │   ├── done/
│   │   │   ├── with-tag-next/    # --done --tag-next allowed (not mutually exclusive)
│   │   │   ├── with-push/        # --done --push allowed (branch under primary)
│   │   │   ├── with-sync-tag-next-push/ # multi-modifier combo accepted
│   │   │   ├── with-dry-run/     # --done --dry-run allowed (composition host)
│   │   │   ├── with-propagate-tags/ # P7: --done --propagate-tags allowed
│   │   │   ├── with-tag-next-propagate/ # P7: --done --tag-next --propagate-tags allowed
│   │   │   ├── with-gen-commit-msg/ # P2: --gen-commit-msg --commit --done allowed
│   │   │   └── with-gen-commit-msg-sync-tag-next-push/ # P2 pre + posts
│   │   └── merge-back/
│   │       ├── with-tag-next/    # --merge-back --tag-next allowed
│   │       ├── with-propagate-tags/ # P7: --merge-back --propagate-tags allowed
│   │       └── with-gen-commit-msg/ # P2: --gen-commit-msg --commit --merge-back
│   ├── reject/                   # illegal combos
│   │   ├── done-with-json/       # --done --json rejected (json only with bare tag-next)
│   │   ├── merge-back-with-json/ # --merge-back --json rejected
│   │   ├── gen-commit-msg-done-without-commit/ # P2: missing --commit with primary
│   │   ├── gen-commit-msg-model-done-without-commit/
│   │   └── gen-commit-msg-dir-with-done/ # composed --dir rejected
│   ├── still-exclusive/
│   │   ├── tag-next-with-list/   # --tag-next --list remains exclusive
│   │   └── gen-commit-msg-with-sync/ # bare --gen-commit-msg --sync exclusive
│   └── help/                     # wrk --help documents gen-commit pre + post composition
├── list/                         # wrk --list (git worktree list wrapper)
│   ├── main-only/                # single main checkout, no linked worktrees
│   ├── with-linked/              # main + one linked worktree
│   ├── from-subpath/             # cwd nested inside main repo
│   └── non-git/                  # cwd is not a git repo (error)
├── fetch-and-verbose/            # --fetch opt-in refresh + -v verbose git logging + --status Remote:
│   ├── fetch/
│   │   ├── invalid-mode/       # --fetch without --projects/--status → error
│   │   │   ├── bare/
│   │   │   ├── with-list/
│   │   │   └── with-done/
│   │   ├── projects/           # default no-fetch vs --fetch on --projects
│   │   │   ├── default-no-fetch/
│   │   │   │   └── stale-tracking-ref/
│   │   │   └── with-fetch/
│   │   │       ├── behind-upstream/
│   │   │       └── fetch-failure/
│   │   └── status/             # --fetch on --status (main vs linked cwd)
│   │       ├── stale-ref-no-fetch/
│   │       ├── with-fetch-behind/
│   │       ├── from-linked-ignore-fetch/
│   │       ├── from-linked-fetch-not-run/
│   │       └── with-fetch-logs/
│   ├── verbose/                # -v major-git-command stderr logging
│   │   ├── list/
│   │   │   └── no-log/         # worktree list is minor → empty stderr
│   │   ├── create/
│   │   │   ├── basic/          # worktree add pre-command log
│   │   │   ├── streams-output/ # worktree add subprocess output streamed
│   │   │   ├── branch-collision/ # fixed-path pre-existing branch → -b (not --no-checkout)
│   │   │   └── no-minor/       # no rev-parse/status lines
│   │   ├── projects/
│   │   │   ├── no-fetch/       # minor reads only → empty stderr
│   │   │   └── with-fetch/     # fetch logged
│   │   ├── done/
│   │   │   └── merge-back/     # merge/worktree remove logged
│   │   └── off/
│   │       └── no-stderr/      # no -v → empty stderr
│   └── status/
│       └── remote/             # --status Remote: on main checkout root block
│           ├── main-clean/
│           │   └── identical/
│           ├── main-no-upstream/
│           └── from-linked-no-remote/
├── status/                       # wrk --status status-block display
│   ├── valid-git-cwd/            # cwd resolves to a git checkout
│   │   ├── root-clean/           # root checkout shown as "." and clean
│   │   ├── subdir-clean/         # nested cwd Dir via Rel (e.g. ../..) + Remote
│   │   ├── multiple-git-dirs/    # primary main + ---- external ---- + nested
│   │   ├── dirty-counts/         # added/changed/renamed/deleted counts
│   │   ├── untracked-dirty/      # ?? untracked file → dirty (1 added, …)
│   │   └── nested-untracked-dirty/ # primary + header + nested dirty (1 added)
│   ├── invalid-git-cwd/
│   │   └── non-git/              # cwd is not a git repo (error)
│   ├── master-field/             # brief Master: on linked worktrees only (plain pipe)
│   │   ├── linked-ahead/         # Master: needs fast forward(+N commits)
│   │   ├── linked-identical/     # Master: identical
│   │   ├── linked-merge-back/    # Master: needs merge back(+N commits)
│   │   ├── linked-diverged/      # Master: diverged(N commits)
│   │   ├── main-no-compare/      # main checkout omits field
│   │   └── nested-main-no-compare/ # nested after header; omits Master:
│   ├── color-output/             # wrk --status alignment + conditional ANSI (--color)
│   │   ├── force-color-clean/    # --color → green Status: clean
│   │   ├── force-color-dirty/    # --color → granular red/grey dirty status
│   │   ├── force-color-master-identical/   # green Master: identical
│   │   ├── force-color-master-fast-forward/ # orange needs fast forward
│   │   ├── force-color-master-merge-back/   # orange needs merge back
│   │   ├── force-color-master-diverged/   # red diverged
│   │   ├── force-color-header/   # --color → gray ---- external ---- header (P3)
│   │   ├── no-color-header/      # pipe, no --color → plain ---- external ---- header
│   │   └── no-color-pipe/        # pipe without --color → no ANSI, brief Master:
│   ├── basename-fallback/        # wrk <basename> --status → saved projects.json lookup (same core as create/--bring)
│   │   ├── single-match/
│   │   │   └── status/           # one saved project → status block for saved root
│   │   ├── cwd-exists/
│   │   │   └── no-fallback/      # ./myrepo in cwd (not git); saved exists → git error, no fallback
│   │   ├── path-with-separator/
│   │   │   └── no-fallback/      # wrk saved/myrepo --status missing → no fallback
│   │   ├── no-match/
│   │   │   └── error/            # zero matches → does not exist
│   │   └── ambiguous/
│   │       ├── tty-select/       # WRK_BASENAME_CONFIRM + stdin selects saved repo
│   │       └── non-tty/          # error listing candidates
│   ├── invalid-mode/
│   │   └── with-list/            # --status with --list is mutually exclusive
│   ├── nested-broken-linked/     # primary + header + external including broken (non-fatal)
│   │   ├── stale-gitdir/         # stale gitdir on nested linked wt; healthy siblings print
│   │   └── color/                # --color red error + gray ---- external ---- header (P3)
│   ├── section-partition/        # nested DOCTEST: PartitionStatusPaths pure helper (P1 GREEN)
│   ├── section-order/            # P2 CLI: primary-first + ---- external ---- (plain without --color)
│   │   ├── header-present/       # main + nested → header between sections
│   │   ├── header-omitted/       # main + WRK linked only → no header
│   │   ├── primary-before-nested/# linked primary before nested external
│   │   ├── linked-list-order/    # two WRK linked → ListLinked order; no header
│   │   └── mixed-full/           # main + in-tree + out-of-tree + nested
│   ├── main-repo-worktrees/      # nested DOCTEST: primary main-owned linked; Dir=statusDirLine
│   │   ├── no-linked-external/   # clean main only → primary; no header
│   │   ├── external-clean/       # WRK out-of-tree → primary full block (Rel Dir)
│   │   ├── external-dirty/       # out-of-tree dirty counts in primary
│   │   ├── in-tree-only/         # in-tree linked → primary; no header
│   │   ├── mixed-external-in-tree/ # in-tree + out-of-tree ListLinked order; no header
│   │   ├── external-broken/      # alive path, broken git → minimal error block
│   │   ├── external-prunable/    # removed checkout → minimal prunable block
│   │   ├── from-linked-cwd/      # --status inside external wt → no main-repo sectioning
│   │   ├── ordering-two-external/ # two out-of-tree → ListLinked primary order
│   │   ├── color-broken/         # --color red error on primary broken block
│   │   ├── from-main-subdir/     # cwd main/pkg/api → Dir ../.. + Remote
│   │   └── from-deep-subdir/     # cwd main/a/b/c/d → main Dir absolute + Remote
│   └── main-flag/                # wrk --main --status: main content; Dir vs inv cwd
│       ├── happy/                # Dir-aware content match vs --status from main
│       │   ├── from-external-wt/ # cwd = external wrk worktree
│       │   │   ├── main-then-status/   # Args --main --status
│       │   │   └── status-then-main/   # Args --status --main
│       │   ├── already-at-main/  # cwd = main root (Dirs match plain --status)
│       │   └── from-in-tree-linked/ # full main status, not linked-cwd shortcut
│       ├── exclusive/
│       │   └── with-list/        # --main --status --list → mutually exclusive
│       ├── events/
│       │   └── command-status/   # events.jsonl command=status; args include both
│       └── with-fetch/           # --main --status --fetch allowed
├── non-git-cwd/                  # cwd is not a git repo (error, no-args create)
├── stderr-newline/               # hard-error stderr ends with trailing \n (main print path)
├── dir-arg/                      # wrk <dir> optional first positional
│   ├── create/
│   │   └── basic/                # wrk <repoDir> from WorkRoot creates worktree
│   ├── list/
│   │   └── from-dir/             # wrk <repoDir> --list from WorkRoot
│   └── missing-dir/              # wrk <nonexistent> → does not exist
├── target-dir/                   # wrk <dir> <target-dir> custom spawn location
│   ├── target-missing/           # <target-dir> does not exist
│   │   ├── parent-exists/        # spawn exactly at <target-dir> (case 1; SETUP-only group)
│   │   │   ├── basic/            # preferred branch free (always -b)
│   │   │   └── branch-collision/ # preferred branch taken → branch -1 only
│   │   └── parent-missing/       # parent missing → error (case 3)
│   ├── target-exists/            # <target-dir> exists
│   │   ├── basic-subdir/         # spawn default-named sub-dir under it (case 2)
│   │   ├── collision-suffix/     # sub-dir name collides → -N suffix (case 2)
│   │   └── target-is-file/       # target is a file → error (edge)
│   ├── relative-path/            # relative <target-dir> resolved vs shell cwd
│   ├── with-other-mode/          # target-dir + other mode → error
│   │   ├── with-list/            # wrk <dir> <target-dir> --list
│   │   └── with-bring/           # wrk <dir> <target-dir> --bring <dep>
│   └── reuse-same-repo/          # Policy B: named bring avoid duplicate same mainRepo
│       ├── no-prior-linked/      # no linked WT → create as today; no skip prompt
│       └── existing-linked/      # live linked WT(s) of source main
│           ├── tty/              # UseScriptTTY + stdin Y/n
│           │   ├── skip-default/ # Enter → skip; stdout = existing
│           │   ├── proceed-n/    # n → create under target (branch -1 if needed)
│           │   └── multi-skip/   # two linked WTs → lex-smallest skip
│           └── non-tty/
│               └── refuse/       # hard error; empty stdout; no new WT
├── task/                          # wrk --task and wrk --set-task
│   ├── spawn/                     # --task when creating worktree
│   │   ├── basic/                 # wrk --task "fix login bug" → slug in name
│   │   ├── t-alias/               # wrk -t "fix login bug" → slug in name; event args use -t
│   │   ├── special-chars/         # capitals, symbols → sanitized slug
│   │   ├── long-task/             # >64 runes → truncated (soft cap regression)
│   │   ├── name-budget/           # fit path/branch ≤255 bytes (reserve 3 for -99)
│   │   │   ├── long-prefix-fit/   # long basename + long task → create; Base/branch ≤255
│   │   │   └── prefix-alone-too-long/ # prefix exceeds budget → clear error
│   │   ├── empty-task/            # --task "" → error
│   │   ├── empty-slug/            # --task "!!!" → error (slug empty)
│   │   ├── with-done/             # --task + --done → mutually exclusive
│   │   ├── sequence/              # two --task "same" calls → -N suffix
│   │   ├── branch-collision/      # pre-existing branch blocks → suffix
│   │   └── target-dir/            # wrk <dir> <target> --task → branch has slug
│   └── set-task/                  # --set-task inside linked worktree
│       ├── non-tty/               # non-TTY → "requires terminal" error
│       ├── empty-desc/            # --set-task "" → error
│       ├── empty-slug/            # --set-task "!!!" → error
│       ├── not-linked/            # from main repo → error
│       ├── not-wrk-worktree/      # custom branch → cannot parse → error
│       ├── fixed-path-unsupported/ # fixed spawn path dir name → error
│       ├── path-collision-suffix/ # target path exists → suffix walk
│       ├── branch-collision-suffix/ # target branch exists → suffix walk
│       ├── legacy-slash-migrate/  # feature/foo-{date} → sanitized rename
│       ├── rename-succeeds/       # TTY-confirmed rename via WRK_SET_TASK_CONFIRM=1
│       ├── slug-unchanged/        # same slug → no-op "task unchanged"
│       ├── name-budget-fit/       # long basename + long set-task → rename ≤255
│       ├── propagate/             # --set-task updates gitdir for nested repos
│       │   ├── single-external-dep/ # external dep's gitdir updated to new path
│       │   ├── non-external-linked-dep/ # manual deps/foo linked wt gitdir updated to new path
│       │   └── abs-replace-rewritten/ # go.mod abs replace rewritten after consumer rename
│       └── with-dir/              # wrk <dir> --set-task (target via argument)
│           ├── rename-succeeds/   # rename worktree at given <dir>
│           ├── empty-desc/        # empty description → error
│           ├── mutually-exclusive/# with --list → mutual exclusion error
│           └── missing-dir/       # non-existent dir → "does not exist"
├── forgot-task-flag/              # forgot -t: task-like positionals promote / error
│   ├── two-arg/                   # wrk <dir> <arg2> without -t
│   │   ├── task-like/
│   │   │   ├── non-tty/           # hard error + -t hint
│   │   │   │   ├── spaces/        # multi-word second positional
│   │   │   │   ├── over-120-bytes/# length >120, no spaces
│   │   │   │   └── over-255-component/ # component >255 / ENAMETOOLONG class
│   │   │   ├── confirm-y/spaces/  # WRK_TASK_LIKE_CONFIRM + y → promote WRK_HOME
│   │   │   ├── confirm-n/spaces/  # n → keep target-dir path
│   │   │   └── yes-flag/spaces/   # -y auto-promote
│   │   └── not-task-like/
│   │       ├── path-like-dot-slash/ # ./real-target unchanged
│   │       ├── short-token/       # out → target-dir
│   │       └── with-explicit-task/# -t set → second stays target-dir
│   └── one-arg/                   # wrk <arg1>
│       ├── task-like/
│       │   ├── non-tty/spaces/    # error + source-dir hint
│       │   ├── confirm-y/spaces/  # promote create from cwd
│       │   └── yes-flag/spaces/   # -y auto-promote from cwd
│       └── not-task-like/
│           └── existing-source/   # resolvable path → no prompt
├── yes-flag/                     # universal -y / --yes flag
│   ├── done/
│   │   ├── ahead-non-tty/        # wrk --done -y on own ahead wt (non-TTY)
│   │   └── ahead-no-prompt/      # TTY + -y: no Proceed? shown (label: tty)
│   ├── merge-back/
│   │   └── ahead-non-tty/        # wrk --merge-back -y merges, keeps wt
│   ├── set-task/
│   │   └── rename-non-tty/       # wrk --set-task -y renames without TTY error
│   ├── no-op/
│   │   └── create-with-yes/      # wrk -y create same as bare wrk
│   └── cascade/
│       ├── non-tty-rejects/      # ahead external + wrk --done -y → error
│       └── tty-auto-yes/         # TTY + -y auto-confirms cascade merge (label: tty)
├── exec/                         # --exec cut-flag: run command in mode target dir
│   ├── create/                   # native create + --exec
│   │   ├── basic-pwd/            # --exec pwd → path then pwd in wt
│   │   └── args-passthrough/     # --exec echo --task → --task not wrk flag
│   ├── cd/                       # --cd + --exec
│   │   └── with-followup/        # follow-up written; stdout = pwd of jump dir
│   ├── bring/                    # --bring + --exec
│   │   └── basic-pwd/            # external path then pwd in external wt
│   ├── set-task/                 # --set-task + --exec
│   │   └── after-rename/         # pwd = new path after move
│   ├── done/                     # --done + --exec
│   │   └── exec-on-main/         # pwd = main repo (not removed wt)
│   ├── reject/                   # --exec with non-allowed modes
│   │   ├── with-list/            # --list --exec true → error
│   │   └── with-status/          # --status --exec true → error
│   └── empty-flag/               # parse errors for cut flag
│       ├── bare-exec/            # --exec alone → requires command
│       └── equals-form/          # --exec=pwd → reject equals form
└── projects/                     # project persistence + event logging
    ├── auto-record/              # auto-record main repo on every invocation
    │   ├── no-dir/               # effective work dir = process cwd
    │   │   ├── main-cwd/         # cwd is main repo root
    │   │   └── subdir-cwd/       # cwd is nested subpath inside repo
    │   ├── dir-arg/              # effective work dir = <dir> positional
    │   │   ├── main-repo/        # wrk <mainRepo>
    │   │   ├── linked-worktree/  # wrk <linkedWt> → main repo
    │   │   └── nested-subpath/   # wrk <nestedSubpath> → main repo
    │   ├── non-git/              # non-git cwd → no record
    │   ├── missing-dir/          # wrk <nonexistent> → no record
    │   └── fail-after-record/    # dirty --done fails but project recorded
    ├── remote-brief/             # wrk --projects shared Remote: brief labels (plain pipe)
    │   ├── ahead-of-upstream/    # Remote: needs push(+N commit)
    │   ├── behind-upstream/      # Remote: needs pull(N commit(s) behind)
    │   ├── diverged/             # Remote: diverged(N commits)
    │   └── up-to-date/           # Remote: identical
    ├── detailed-status/          # wrk --projects detailed status blocks (plain pipe output)
    │   ├── single-clean-no-wts/  # one project, clean, no linked wts
    │   ├── stale-gitdir-linked/  # stale .git gitdir -> inline error detail, exit 0
    │   ├── broken-main-repo/     # recorded path no longer git -> minimal Dir+Status error block
    │   ├── prunable-worktrees/   # deleted checkout -> 0 total, 1 prune (summary only)
    │   ├── with-linked-mixed/    # Worktrees:    3 total, 1 dirty
    │   ├── ahead-of-upstream/    # Remote: needs push(+N commit)
    │   ├── no-upstream/          # Remote: (no upstream)
    │   ├── multiple-projects/    # two blocks, lex order, blank separator
    │   └── empty/                # exit 0, empty stdout
    ├── color-output/             # wrk --projects alignment + conditional ANSI (--color)
    │   ├── no-color-pipe/        # pipe without --color → no ANSI, aligned Worktrees
    │   ├── force-color-dirty-status/   # granular red/grey dirty Status segments
    │   ├── force-color-dirty-partial/  # 2 changed, zero other counts → grey + red mix
    │   ├── force-color-needs-push/     # orange needs push(...)
    │   ├── force-color-needs-pull/     # orange needs pull(...)
    │   ├── force-color-diverged/       # red diverged(...)
    │   ├── force-color-worktrees-dirty/ # red N dirty portion only
    │   ├── force-color-stale-gitdir-linked/ # red on error summary + detail lines
    │   ├── clean-no-color/       # all clean + --color → no highlights
    │   └── color-with-list/      # --list --color → list unchanged, no ANSI
    ├── list/
    │   └── projects/
    │       ├── empty/            # wrk --projects empty → exit 0, no output
    │       └── after-records/    # sorted detailed blocks after auto-record
    ├── add/
    │   ├── manual/
    │   │   ├── main-repo/        # wrk --add <mainRepo>
    │   │   └── linked-worktree/  # wrk --add <linkedWt> → main repo
    │   └── idempotent/           # auto + manual → single entry
    ├── remove/
    │   ├── manual/
    │   │   ├── main-repo/        # --add then --rm <mainRepo> → gone, stdout path
    │   │   └── linked-worktree/  # --rm <linkedWt> → main repo removed
    │   ├── idempotent/
    │   │   ├── not-recorded/     # never recorded → exit 0, empty stdout
    │   │   └── already-removed/  # remove twice → second empty stdout
    │   ├── missing-path-arg/     # wrk --rm (no path) → error
    │   ├── stale-path/           # record, delete .git, --rm old path → removed
    │   └── invalid-mode/
    │       └── remove-with-list/ # wrk --rm X --list → mutually exclusive
    ├── events/
    │   ├── append-on-success/    # create → event exit_code 0
    │   └── append-on-failure/    # failed command → event exit_code != 0
    ├── invalid-mode/
    │   ├── projects-with-list/   # wrk --projects --list → mutual exclusion
    │   └── add-missing-path/     # wrk --add without path → error
    ├── output-streaming/         # wrk --projects incremental stdout (per-project as ready)
    │   └── fast-before-slow-gather/ # fast aaa block streams before slow zzz gather ends
    ├── perf-profile/             # WRK_PROJECTS_PERF_LOG instrumentation + parallel budgets
    │   ├── emits-events/
    │   │   └── many-worktrees/   # JSONL lifecycle + 12 worktree_status events
    │   ├── budget/
    │   │   └── many-worktrees-parallel/ # worktree_status_all <100ms, run_end <200ms
    │   └── structure/
    │       └── dedup-list-linked/  # single ListLinked per project (not skip+summary)
    └── basename-fallback/        # create-mode basename → saved projects.json lookup
        ├── single-match/
        │   └── create/           # one match → worktree from saved path
        ├── cwd-exists/
        │   └── no-fallback/      # ./basename dir in cwd (non-git) → no lookup
        ├── cwd-file-exists/      # ./basename file in cwd → guided error (not git-repo failure)
        │   ├── single-match/
        │   │   └── guided-error/ # file + 1 saved project → concrete-path hint
        │   ├── ambiguous/
        │   │   └── guided-error/ # file + 2 saved projects → <full-path> hint
        │   └── no-match/
        │       └── short-error/  # file + 0 saved projects → single-line error
        ├── no-match/
        │   └── error/            # zero matches → does not exist
        ├── path-with-separator/
        │   └── no-fallback/      # sub/foo → no lookup
        ├── ambiguous/
        │   ├── tty-select/       # WRK_BASENAME_CONFIRM + stdin index
        │   └── non-tty/          # error listing candidates
        └── other-mode/
            └── no-fallback/      # wrk basename --list → no lookup
└── where/                        # wrk --where basename lookup (projects.json only)
    ├── single-match/
    │   └── basic/                # one match → stdout saved abs path
    ├── ambiguous/
    │   └── two-matches/          # two matches → stdout both paths sorted
    ├── no-match/
    │   └── error/                # zero matches → stderr no-match
    ├── non-basename/
    │   ├── path-with-separator/  # sub/spl → basename-only rejection
    │   └── absolute-path/        # /abs/.../spl → basename-only rejection
    ├── empty-arg/
    │   └── error/                # wrk --where (no value) → requires argument
    ├── mutual-exclusion/
    │   └── with-status/          # --where spl --status → mutually exclusive
    └── cwd-exists/
        └── no-local-fallback/    # ./spl in cwd (non-git) + saved spl → saved path only
└── cd/                           # wrk --cd jump (in-place follow-up | fallback shell)
    ├── in-place/                 # WRK_FOLLOWUP_FILE set → empty stdout + cd follow-up
    │   ├── abs-path/             # wrk --cd /abs
    │   ├── relative-path/        # wrk rel/target --cd
    │   └── basename-expanded/    # wrk --cd myrepo → follow-up uses expanded abs
    ├── fallback/                 # channel closed → stdout abs + install hint + shell
    │   ├── abs-path/             # fake bash; cwd = target
    │   ├── relative-path/
    │   ├── basename-single/      # one projects match
    │   └── shell-exit-nonzero/   # fake bash exit 42 → wrk 42
    ├── syntax-forms/             # flag-then-path vs path-then-flag (in-place)
    │   ├── flag-then-path/       # wrk --cd PATH
    │   └── path-then-flag/       # wrk PATH --cd
    ├── resolution/               # arg/path errors + local basename wins
    │   ├── missing-arg/          # wrk --cd → requires path
    │   ├── no-match/             # basename 0 matches
    │   ├── missing-path/         # abs missing
    │   ├── not-a-directory/      # path is a file
    │   ├── ambiguous-non-tty/    # two saved basenames
    │   └── local-basename/       # ./myrepo wins over projects
    ├── mutual-exclusion/
    │   ├── with-list/
    │   ├── with-no-cd/
    │   └── with-where/
    └── events/
        └── command-cd/           # events.jsonl command "cd"
├── web/                          # wrk --web React SPA (wrk-react)
│   ├── mutual-exclusion/         # --web + other modes
│   │   ├── with-list/            # --web --list → mutually exclusive
│   │   └── with-status/          # --web --status → mutually exclusive
│   ├── port-without-web/         # --port without --web → only valid with --web
│   ├── unexpected-args/          # --web some-dir → unexpected arguments
│   ├── help-mentions-web/        # wrk -h documents --web and --port
│   ├── serves-page/              # WebProbe GET / → SPA shell + markers + listen URL
│   ├── mockup-repo-view/         # WebProbe GET /mockup/repo-view → SPA client-route 200
│   └── api-projects-empty/       # WebProbe GET /api/wrk/projects → {"projects":[]}
└── scan-git-repos/               # wrk --scan-git-repos discover + print-only (no projects.json)
    ├── defaults/                 # bare flag: default root = $HOME (~), not ~/Projects
    │   ├── home-root/            # FakeHome=WorkRoot + repo-a; no Projects → finds main
    │   └── home-unusable/        # FakeHome missing path → non-zero; no Projects msg
    ├── record/                   # success: discover mains under explicit roots
    │   ├── basic-add/            # one main → stdout path; projects.json empty
    │   ├── idempotent/           # already recorded → exit 0; no dup; empty stdout
    │   ├── main-only/            # main + linked wt under root → only main recorded
    │   └── with-no-cache/        # --scan-git-repos --no-cache ROOT still records
    ├── filter-home-subpath/      # P5 two-base home + emit filter (FakeHome)
    │   ├── default-home-universe/    # bare scan → home/repos.json universe=home
    │   ├── projects-shares-home-cache/  # ~/Projects reuses home cache; stdout Projects-only
    │   └── debug-cache-base-filter/  # -v → stderr cache_base + filter
    ├── debug/                    # wire Options.Debug from -v OR WRK_SCAN_DEBUG
    │   ├── via-verbose/          # second scan -v → stderr scan: + mode=
    │   ├── via-env/              # WRK_SCAN_DEBUG=1 no -v → scan: present
    │   └── off/                  # neither → zero scan: markers
    ├── streaming/                # OnRepo record+print as found (discovery order)
    │   ├── discovery-order/      # multi-root CLI order on stdout (not lex sort)
    │   └── first-path-before-finish/  # label:slow; first path bytes before process exit
    ├── interrupt/                # SIGINT/SIGTERM mid-scan
    │   └── sigint-after-first-path/  # exit 130 + warning + partial projects
    ├── mutual-exclusion/
    │   └── with-projects/        # --scan-git-repos --projects → mutually exclusive
    ├── no-cache-without-scan/    # bare --no-cache → only valid with --scan-git-repos
    └── help-mentions-scan-git-repos/  # wrk -h documents --scan-git-repos, --no-cache, default ~
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | create-worktree/main-checkout/basic-create | Fresh git repo on `main`, first `wrk` |
| 2 | create-worktree/main-checkout/sequence-increment | Run `wrk` twice from same repo/branch |
| 3 | create-worktree/main-checkout/branch-collision | Pre-create branch `main-{date}`, no worktree at date path → joint `-1` via always-new `-b` |
| 4 | non-git-cwd | cwd is not a git repository |
| 5 | create-worktree/from-linked-worktree | cwd is existing linked worktree |
| 6 | create-worktree/detached-head | cwd on detached HEAD uses 7-char hash token |
| 7 | create-worktree/main-checkout/slash-branch | Branch `feature/foo` → path+branch `feature-foo-{date}` (no `/`) |
| 8 | create-worktree/from-git-subpath | cwd is nested subdir inside checkout; basename from repo root |
| 8a | create-worktree/git-lfs-hooks/minimal-path-succeeds | LFS hook + stripped PATH; git-lfs in $HOME/.local/bin → create fails (expected) |
| 8b | create-worktree/git-lfs-hooks/from-other-cwd | wrk \<repo\> from foreign cwd + stripped PATH → create fails (expected) |
| 9 | done/already-included | wt branch merged into main; `--done` removes wt + branch |
| 10 | done/ahead-confirm | wt ahead; `--done --confirm-from-stdin` + `\n` → ff-merge + remove |
| 11 | done/ahead-decline | wt ahead; `--confirm-from-stdin` + `n\n` → aborted, wt remains |
| 12 | done/ahead-non-tty | wt ahead; no confirm flag → non-zero (cannot prompt) |
| 13 | done/dirty | uncommitted file in wt → non-zero |
| 14 | done/not-linked | cwd is main repo → `not a linked worktree` |
| 15 | done/from-subpath | cwd is subdir inside linked wt; `--done` uses checkout root |
| 16 | list/main-only | single main repo; stdout matches `git worktree list` |
| 17 | list/with-linked | main + linked worktree; stdout lists both paths |
| 18 | list/from-subpath | cwd nested in main repo; stdout same as from repo root |
| 19 | list/non-git | cwd is not a git repository |
| 34a | bring/basic | Consumer requires dep; `--bring` creates external wt + replace + gitignore; no SKIP |
| 34b | bring/not-a-dependency | Consumer has go.mod without require; `--bring` → exit 0, worktree+gitignore, SKIP not a dependency |
| 34c | bring/not-go-module | Dep git without go.mod; `--bring` → exit 0, worktree+gitignore, SKIP not a go module |
| 34d | bring/consumer-no-modules | Consumer zero go.mod; `--bring` → exit 0, worktree+gitignore, SKIP consumer has no Go modules |
| 34f | bring/exec-after-skip | not-a-dependency + `--exec pwd` → SKIP on stderr; pwd runs in external wt |
| 34g | bring/reuse-same-repo/second-bring | second `--bring` same dep → same path; reuse warning; no `-1` |
| 34h | bring/reuse-same-repo/multi-external | two external WTs same depMain → reuse lex-smallest + multi/also-present warnings |
| 34i | bring/reuse-same-repo/two-different-deps | bring dep1 then dep2 → no false reuse of other dep's external |
| 34k | bring/no-dep/skips-replace-tidy | `--bring --no-dep` matching → external wt; go.mod unchanged; no tidy |
| 34l | bring/no-dep/with-verbose-no-tidy-log | `--bring --no-dep -v` → may log git; never `mod tidy` |
| 34m | bring/no-dep/not-a-dep-still-works | no require + `--no-dep` → wt; no SKIP (analyse skipped) |
| 34n | bring/no-dep/invalid/bare | bare `--no-dep` → only valid with dep/bring/all-deps |
| 34o | bring/no-dep/invalid/with-list | `--list --no-dep` → only-valid-with message |
| 34p | bring/verbose/tidy-pre-line | `--bring -v` → timestamped `$ go -C … mod tidy` on stderr |
| 34q | bring/verbose/no-verbose-silent-tidy | `--bring` without `-v` → no tidy pre-line |
| 34r | bring/verbose/streams-worktree-add | `--bring -v` streams `Preparing worktree` |
| 34s | bring/help-mentions-no-dep | `wrk -h` documents `--no-dep` and `-v` tidy logging |
| 48 | done/external-cascade | `--done` cascades to `external/*` dep wt first; parent errors on local replace (names go.mod + directive) |
| 49 | done/local-replace-blocks | extra-repo `replace => ./external/foo` (non-existent) blocks `--done` at guard (names go.mod + directive) |
| 50 | dir-arg/create/basic | `wrk <repoDir>` from WorkRoot creates worktree |
| 51 | dir-arg/list/from-dir | `wrk <repoDir> --list` matches `git worktree list` |
| 52 | dir-arg/missing-dir | `wrk <nonexistent>` → does not exist |
| 53 | target-dir/target-missing/parent-exists/basic | `wrk <dir> <target>` spawns exactly at `<target>` (parent exists; branch free) |
| 53a | target-dir/target-missing/parent-exists/branch-collision | fixed path + pre-existing branch → path fixed, branch `-1` via `-b` |
| 54 | target-dir/target-missing/parent-missing | `wrk <dir> <target>` parent missing → error |
| 55 | target-dir/target-exists/basic-subdir | `wrk <dir> <target>` existing dir → default-named sub-dir |
| 56 | target-dir/target-exists/collision-suffix | existing dir + colliding sub-dir → `-N` suffix |
| 57 | target-dir/target-exists/target-is-file | `<target>` is a file → error |
| 58 | target-dir/relative-path | relative `<target>` resolved against shell cwd |
| 59 | target-dir/with-other-mode/with-list | `wrk <dir> <target> --list` → unexpected arguments |
| 60 | target-dir/with-other-mode/with-bring | `wrk <dir> <target> --bring <dep>` → unexpected arguments |
| 60a | target-dir/reuse-same-repo/no-prior-linked | named bring, no prior linked WT → create under target; no skip prompt |
| 60b | target-dir/reuse-same-repo/existing-linked/tty/skip-default | TTY + Enter → skip; stdout = existing path (label: tty) |
| 60c | target-dir/reuse-same-repo/existing-linked/tty/proceed-n | TTY + `n` → create new under target (label: tty) |
| 60d | target-dir/reuse-same-repo/existing-linked/tty/multi-skip | two linked WTs + Enter → skip lex-smallest (label: tty) |
| 60e | target-dir/reuse-same-repo/existing-linked/non-tty/refuse | non-TTY + existing linked → hard error; empty stdout; no new WT |
| 61 | done/no-go-mod | linked wt whose checkout has no go.mod; `--done` merge-back succeeds (guard no-op) |
| 62 | done/not-linked-no-go-mod | main repo without go.mod; `--done` → `not a linked worktree` (not `no go.mod found`) |
| 63 | done/sub-module-replace-blocks | main go.mod clean but `sub/go.mod` has local replace → guard blocks `--done` (names go.mod + directive) |
| 64 | done/cascade-merge-base | cascade must remove dep wt, not crash `failed to find merge base` (dep branch vs consumer main share no history) |
| 65 | done/cascade-force-removal/ahead-non-tty-errors | ahead external dep + non-TTY `--done` → error; dep wt + commits preserved (no force-remove) |
| 65a | done/cascade-force-removal/ahead-non-tty-with-y-still-errors | same + `-y` → still errors (cascade guard) |
| 65b | done/cascade-non-tty-rejects-with-confirm-from-stdin | ahead external + `--confirm-from-stdin` on non-TTY → error before mutations |
| 65c | done/cascade-dep-merge-back | regression: ahead dep + `--confirm-from-stdin` on non-TTY → error (option A; no cascade merge) |
| 120 | yes-flag/done/ahead-non-tty | `wrk --done -y` merges own ahead wt on non-TTY |
| 121 | yes-flag/done/ahead-no-prompt | TTY + `wrk --done -y` shows no `Proceed?` (label: tty) |
| 122 | yes-flag/merge-back/ahead-non-tty | `wrk --merge-back -y` merges, keeps worktree |
| 123 | yes-flag/set-task/rename-non-tty | `wrk --set-task -y` renames on non-TTY |
| 124 | yes-flag/no-op/create-with-yes | `wrk -y` create same as bare `wrk` |
| 125 | yes-flag/cascade/non-tty-rejects | ahead external + `wrk --done -y` on non-TTY → error |
| 126 | yes-flag/cascade/tty-auto-yes | TTY + `wrk --done -y` merges cascade + consumer (label: tty) |
| 65a | done/cascade-non-external-linked | manual `deps/foo` linked wt (not under `external/`) → cascade removes it, consumer `--done` exit 0 |
| 65b | done/cascade-external-and-deps | `external/*` + `deps/foo` both linked → cascade removes both, consumer `--done` exit 0 |
| 66 | done/intra-replace-warns | intra-repo `replace example.com/foo => ./submod` (existing, same toplevel) → WARN, exit 0, merge-back proceeds |
| 66b | done/intra-replace-cross-worktree | abs replace to main-checkout `submod` from linked wt; `wrk --done <wt>` from outside → extra-repo block |
| 67 | done/intra-replace-strict-blocks | intra-repo replace + `--no-in-module-replace` → block, names go.mod + directive |
| 68 | done/no-in-module-replace-without-done | `wrk --list --no-in-module-replace` → non-zero, `--no-in-module-replace is only valid with --done` |
| 68a | done-sync/basic-propagate | `--done -y --sync`: wtA removed; wtB pass2 FF from main; blank line before sync block |
| 68b | done-sync/no-other-wt | `--done -y --sync` with only wtA → primary message + blank + zero-summary sync |
| 68c | done-sync/aborted-no-sync | decline confirm → `merge-back aborted`; no `synced:`; wt remains |
| 68d | done-sync/flag-order-sync-first | `--sync --done -y` same composition as `--done --sync` |
| 68e | merge-back-sync/basic-propagate | `--merge-back -y --sync`: wtA kept; wtB gets main tip |
| 68f | merge-back-sync/source-wt-stays | `--merge-back -y --sync`: wtA still on disk; no `worktree removed:` |
| 68g | done-compose/allow/done/with-tag-next | `--done --tag-next` not mutually exclusive at flag layer |
| 68h | done-compose/allow/done/with-push | `--done --push` not rejected as tag-next-only |
| 68i | done-compose/allow/done/with-sync-tag-next-push | `--done --sync --tag-next --push` flag layer accepts |
| 68j | done-compose/allow/done/with-dry-run | `--done --dry-run` not rejected as only-valid-with host |
| 68j2 | done-compose/allow/done/with-propagate-tags | P7: `--done --propagate-tags` flag layer accepts |
| 68j3 | done-compose/allow/done/with-tag-next-propagate | P7: `--done --tag-next --propagate-tags` flag layer accepts |
| 68k | done-compose/allow/merge-back/with-tag-next | `--merge-back --tag-next` not mutually exclusive |
| 68k2 | done-compose/allow/merge-back/with-propagate-tags | P7: `--merge-back --propagate-tags` flag layer accepts |
| 68l | done-compose/reject/done-with-json | `--done --json` → non-zero; names json + done |
| 68m | done-compose/reject/merge-back-with-json | `--merge-back --json` → non-zero; names both |
| 68n | *(retired)* done-compose/reject/bare-push | bare `--push` is standalone — see `push/` |
| 68o | done-compose/still-exclusive/tag-next-with-list | `--tag-next --list` still mutually exclusive |
| 68o2 | done-compose/help | `wrk --help` → exit 0; `--done`/`--merge-back` list optional `--tag-next`/`--push`/`--gen-commit-msg`; `--push` dual meaning (not only tag-next); no tag-next exclusive-with-done claim |
| 68o3 | done-compose/allow/done/with-gen-commit-msg | P2: `--gen-commit-msg --commit --model=m --done` not mutually exclusive (Classic RED) |
| 68o4 | done-compose/allow/done/with-gen-commit-msg-sync-tag-next-push | P2: gen-commit pre + posts flag layer accepts (Classic RED) |
| 68o5 | done-compose/allow/merge-back/with-gen-commit-msg | P2: `--gen-commit-msg --commit --merge-back` not mutually exclusive (Classic RED) |
| 68o6 | done-compose/reject/gen-commit-msg-done-without-commit | P2: missing `--commit` with primary → require-commit error (Classic RED) |
| 68o7 | done-compose/reject/gen-commit-msg-model-done-without-commit | P2: `--model` without `--commit` + primary rejected (Classic RED) |
| 68o8 | done-compose/reject/gen-commit-msg-dir-with-done | P2: composed `--dir` with primary rejected (Classic RED) |
| 68o9 | done-compose/still-exclusive/gen-commit-msg-with-sync | bare `--gen-commit-msg --sync` still exclusive (GREEN pin) |
| 68p | done-push/pushes-main | `--done -y --push`: origin/main == post-merge main HEAD; push confirmation line |
| 68q | done-push/flag-order-push-first | `--push --done -y` same composition as `--done -y --push` |
| 68r | done-push/no-remote | `--done -y --push` without origin → non-zero; stderr mentions remote |
| 68s | done-pipeline/tag-next-local | `--done -y --tag-next`: local v0.0.2 at main HEAD; event command done |
| 68t | done-pipeline/tag-next-push | `--done -y --tag-next --push`: local+origin tags; origin/main == main |
| 68u | done-pipeline/sync-tag-next | `--done -y --sync --tag-next`: sync then local tags; no push |
| 68v | done-pipeline/sync-tag-next-push | full combo ordered: primary → sync → tag-next → push |
| 68w | done-pipeline/aborted-full-flags | decline with full flags incl. `--propagate-tags` → no post stages; wt remains |
| 68w2 | done-pipeline/tag-next-propagate | P7: `--done -y --tag-next --propagate-tags`: tag + consumer bump; event done |
| 68w3 | done-pipeline/propagate-existing | P7: `--done -y --propagate-tags` on existing tags; event done |
| 68x | done-pipeline/flag-order-full | `--push --tag-next --sync --done -y` same as full combo |
| 68y | done-pipeline/dry-run/alone | `--done --dry-run`: MergeBack plan only; zero mutations; no post stages |
| 68y2 | done-pipeline/dry-run/with-gen-commit-msg | P2: gen-commit dry plan + done dry plan; HEAD/subject unchanged (Classic RED) |
| 68z | done-pipeline/dry-run/full-combo | `--done --sync --tag-next --push --dry-run`: full plan; zero mutations; no `-y` |
| 68z2 | done-pipeline/dry-run/tag-next-propagate | P7: dry-run plans tag + would-propagate; zero mutations |
| 68za | done-pipeline/dry-run/cascade-external | cascade `would: cascade merge-back`; external still present |
| 68zb | done-pipeline/dry-run/no-prompt-without-y | dry-run without `-y` exits 0; no confirm TTY error |
| 68ux1 | done-output/dry-run-phases/with-cascade | P2: `==> cascade` then would: then `==> own` (Classic RED) |
| 68ux2 | done-output/dry-run-phases/zero-cascade | P2: zero targets still prints `==> cascade` + `==> own` (Classic RED) |
| 68ux3 | done-output/cascade-failure/structured-error | P2: cascade fail → `Error:` + path; not bare `rebase conflict:` (Classic RED) |
| 68zc | merge-back-pipeline/tag-next-local | `--merge-back -y --tag-next`: local v0.0.2; source wt kept; event merge-back |
| 68zd | merge-back-pipeline/tag-next-push | `--merge-back -y --tag-next --push`: branch+tags; wt kept |
| 68ze | merge-back-pipeline/sync-tag-next | `--merge-back -y --sync --tag-next`: sync then local tags; wt kept |
| 68zf | merge-back-pipeline/sync-tag-next-push | full combo ordered: primary → sync → tag-next → push; wt kept |
| 68zf2 | merge-back-pipeline/tag-next-propagate | P7: `--merge-back -y --tag-next --propagate-tags`; wt kept; event merge-back |
| 68zg | merge-back-pipeline/aborted-full-flags | decline with full flags → no sync/tag/push; wt remains |
| 68zh | merge-back-pipeline/flag-order-full | `--push --tag-next --sync --merge-back -y` same as full combo |
| 68zi | merge-back-pipeline/dry-run/tag-next | `--merge-back --tag-next --dry-run`: keep wt; tag planned; no mutations |
| 69 | task/spawn/basic | `wrk --task "fix login bug"` → dir/branch include `-fix-login-bug` |
| 69a | task/spawn/t-alias | `wrk -t "fix login bug"` → same slug behavior; event `args: ["-t", "fix login bug"]` |
| 70 | task/spawn/special-chars | Task with capitals, symbols, unicode → sanitized slug |
| 71 | task/spawn/long-task | >64 runes → truncated to 64 (soft cap regression) |
| 71a | task/spawn/name-budget/long-prefix-fit | long basename + long task → create; Base/branch ≤255; slug fitted |
| 71b | task/spawn/name-budget/prefix-alone-too-long | prefix alone over budget → clear non-zero error; no basename chop |
| 72 | task/spawn/empty-task | `--task ""` → error |
| 73 | task/spawn/empty-slug | `--task "!!!"` → error (slug empty after sanitization) |
| 74 | task/spawn/with-done | `--task` + `--done` → mutually exclusive |
| 75 | task/spawn/sequence | Two `wrk --task "same"` calls → `-N` suffix on second |
| 76 | task/spawn/branch-collision | Pre-existing branch with task-slug name → suffix increment |
| 77 | task/spawn/target-dir | `wrk <dir> <target> --task "desc"` → branch has slug, dir is user-specified |
| 78 | task/set-task/non-tty | `--set-task` in non-TTY → error "requires terminal" |
| 79 | task/set-task/empty-desc | `--set-task ""` → error |
| 80 | task/set-task/empty-slug | `--set-task "!!!"` → error |
| 81 | task/set-task/not-linked | `--set-task` from main repo → error |
| 82 | task/set-task/not-wrk-worktree | `--set-task` on custom-branch worktree → cannot parse → error |
| 82a | task/set-task/fixed-path-unsupported | Fixed spawn path dir name → unsupported / cannot parse directory |
| 82b | task/set-task/path-collision-suffix | Target path occupied → suffix walk to `-1` |
| 82c | task/set-task/branch-collision-suffix | Target branch occupied → suffix walk to `-1` |
| 82d | task/set-task/legacy-slash-migrate | Legacy `feature/foo-{date}` → sanitized rename with slug |
| 83 | task/set-task/rename-succeeds | `--set-task "new task"` with WRK_SET_TASK_CONFIRM=1 → worktree renamed, branch renamed |
| 84 | task/set-task/slug-unchanged | `--set-task` with same slug → no-op, prints "task unchanged" |
| 84a | task/set-task/name-budget-fit | long basename + long `--set-task` → rename; Base/branch ≤255 |
| 85 | task/set-task/propagate/single-external-dep | `--set-task` with external dep → consumer renamed, dep gitdir updated to new path |
| 85a | task/set-task/propagate/non-external-linked-dep | `--set-task` with manual `deps/foo` linked wt → consumer renamed, dep gitdir updated to new path |
| 85b | task/set-task/propagate/abs-replace-rewritten | `--set-task` with `wrk --bring` abs replace → go.mod replace rewritten to new consumer path |
| 86 | task/set-task/with-dir/rename-succeeds | `wrk <dir> --set-task "new task"` renames worktree at `<dir>` |
| 87 | task/set-task/with-dir/empty-desc | `wrk <dir> --set-task ""` → error |
| 88 | task/set-task/with-dir/mutually-exclusive | `wrk <dir> --set-task "task" --list` → mutual exclusion error |
| 89 | task/set-task/with-dir/missing-dir | `wrk <nonexistent> --set-task "task"` → does not exist |
| 89a | forgot-task-flag/two-arg/task-like/non-tty/spaces | `wrk <dir> "fix the login bug"` non-TTY → error + `-t` hint |
| 89b | forgot-task-flag/two-arg/task-like/non-tty/over-120-bytes | second positional >120 bytes non-TTY → error + hint |
| 89c | forgot-task-flag/two-arg/task-like/non-tty/over-255-component | second positional >255 component non-TTY → error + hint |
| 89d | forgot-task-flag/two-arg/task-like/confirm-y/spaces | WRK_TASK_LIKE_CONFIRM + y → promote to WRK_HOME task create |
| 89e | forgot-task-flag/two-arg/task-like/confirm-n/spaces | confirm n → fixed multi-word target-dir path |
| 89f | forgot-task-flag/two-arg/task-like/yes-flag/spaces | `-y` auto-promotes multi-word second positional |
| 89g | forgot-task-flag/two-arg/not-task-like/path-like-dot-slash | `./real-target` → target-dir; no treat-as-task |
| 89h | forgot-task-flag/two-arg/not-task-like/short-token | short `out` → target-dir |
| 89i | forgot-task-flag/two-arg/not-task-like/with-explicit-task | `-t` set → second stays target-dir |
| 89j | forgot-task-flag/one-arg/task-like/non-tty/spaces | one-arg multi-word non-TTY → error + source hint |
| 89k | forgot-task-flag/one-arg/task-like/confirm-y/spaces | one-arg confirm y → create from cwd with task |
| 89l | forgot-task-flag/one-arg/task-like/yes-flag/spaces | one-arg `-y` → create from cwd with task |
| 89m | forgot-task-flag/one-arg/not-task-like/existing-source | existing source path → normal create; no prompt |
| 83 | status/valid-git-cwd/root-clean | `wrk --status` from repo root shows `Dir: .` and clean status |
| 84 | status/valid-git-cwd/subdir-clean | `wrk --status` from nested subdir shows `Dir: ../..` + Remote |
| 85 | status/valid-git-cwd/multiple-git-dirs | root + nested independent git repo produce two status blocks |
| 86 | status/valid-git-cwd/dirty-counts | status counts one added, changed, renamed, and deleted entry |
| 86a | status/valid-git-cwd/untracked-dirty | clean repo + one untracked `??` file → dirty `(1 added, 0 changed, 0 renamed, 0 deleted)` |
| 86b | status/valid-git-cwd/nested-untracked-dirty | root clean; nested `tools/child` with only untracked → child dirty `(1 added, …)` |
| 87 | status/invalid-git-cwd/non-git | `wrk --status` outside git fails with `is not a git repository` |
| 88 | status/invalid-mode/with-list | `wrk --status --list` fails as mutually exclusive |
| 88a | status/master-field/linked-ahead | linked wt `Master: needs fast forward(+1 commit)` |
| 88b | status/master-field/linked-identical | linked wt `Master: identical` |
| 88c | status/master-field/linked-merge-back | linked wt `Master: needs merge back(+1 commit)` |
| 88d | status/master-field/linked-diverged | linked wt `Master: diverged(2 commits)` |
| 88e | status/master-field/main-no-compare | main checkout block has no Master: line |
| 88f | status/master-field/nested-main-no-compare | nested independent repo has no Master: line |
| 88g | status/color-output/force-color-clean | `--color` → green `Status: clean` |
| 88h | status/color-output/force-color-dirty | `--color` → granular red/grey dirty status |
| 88i | status/color-output/force-color-master-identical | `--color` → green `Master: identical` |
| 88j | status/color-output/force-color-master-fast-forward | `--color` → orange needs fast forward |
| 88k | status/color-output/force-color-master-merge-back | `--color` → orange needs merge back |
| 88l | status/color-output/force-color-master-diverged | `--color` → red diverged |
| 88m | status/color-output/no-color-pipe | pipe `--status` → no ANSI, brief Master: |
| 88m1 | status/color-output/force-color-header | `--status --color` with nested → gray full-line `---- external ----` (P3) |
| 88m2 | status/color-output/no-color-header | pipe `--status` with nested → plain `---- external ----` (no ANSI) |
| 88n | status/main-repo-worktrees/no-linked-external | clean main only → primary [main]; no header |
| 88o | status/main-repo-worktrees/external-clean | main + WRK out-of-tree → primary both; no header; statusDirLine Dir + Master |
| 88p | status/main-repo-worktrees/external-dirty | out-of-tree primary linked `Status: dirty (...)` |
| 88q | status/main-repo-worktrees/in-tree-only | in-tree linked → primary both; no header |
| 88r | status/main-repo-worktrees/mixed-external-in-tree | in-tree + out-of-tree → primary ListLinked order; no header |
| 88s | status/main-repo-worktrees/external-broken | primary minimal `Status: error: …`, exit 0; no header |
| 88t | status/main-repo-worktrees/external-prunable | primary minimal `Status: prunable`; no header |
| 88u | status/main-repo-worktrees/from-linked-cwd | `--status` from external wt; no main-repo primary/external sections |
| 88v | status/main-repo-worktrees/ordering-two-external | two out-of-tree → primary order = ListLinked; no header |
| 88w | status/main-repo-worktrees/color-broken | `--status --color` red `error:` on primary broken block; no header |
| 88w1 | status/main-repo-worktrees/from-main-subdir | cwd `main/pkg/api`; main Dir `../..` + Remote |
| 88w2 | status/main-repo-worktrees/from-deep-subdir | cwd `main/a/b/c/d`; main Dir absolute + Remote |
| 88x | status/nested-broken-linked/stale-gitdir | nested linked wt stale gitdir → primary + plain header + external; exit 0 |
| 88y | status/nested-broken-linked/color | `--status --color` red `error:` + gray `---- external ----` header (P3) |
| 88y1 | status/section-order/header-present | main + nested → plain `---- external ----` between sections |
| 88y2 | status/section-order/header-omitted | main + WRK linked only → no header (WRK linked is primary) |
| 88y3 | status/section-order/primary-before-nested | ListLinked primary before nested external (vs legacy scan order) |
| 88y4 | status/section-order/linked-list-order | two WRK linked → ListLinked porcelain order; no header |
| 88y5 | status/section-order/mixed-full | main + in-tree + out-of-tree + nested → primary then header then nested |
| 88y6 | status/section-partition/no-external/main-only | PartitionStatusPaths: primary=[main], external=[] |
| 88y7 | status/section-partition/no-external/linked-order | primary preserves ListLinked order after main |
| 88y8 | status/section-partition/no-external/prunable-linked | dead/prunable linked still in primary |
| 88y9 | status/section-partition/no-external/scan-dup-linked | linked also in scan → once in primary; external empty |
| 88y10 | status/section-partition/has-external/main-plus-nested | nested only in external |
| 88y11 | status/section-partition/has-external/multiple-external-sorted | external path-sorted |
| 88y12 | status/section-partition/has-external/mixed-full | primary ListLinked; external nesteds only, sorted |
| 88z1 | status/main-flag/happy/from-external-wt/main-then-status | `--main --status` from external: content match; Dir via inv cwd |
| 88z2 | status/main-flag/happy/from-external-wt/status-then-main | `--status --main` same Dir-aware content as main-then-status |
| 88z3 | status/main-flag/happy/already-at-main | `--main --status` at main root matches plain `--status` Dirs |
| 88z4 | status/main-flag/happy/from-in-tree-linked | full main status Dir-aware; not linked-cwd shortcut |
| 88z5 | status/main-flag/exclusive/with-list | `--main --status --list` mutually exclusive |
| 88z6 | status/main-flag/events/command-status | event command=`status`; args include `--main` and `--status` |
| 88z7 | status/main-flag/with-fetch | `--main --status --fetch` allowed; exit 0 |
| 90 | projects/auto-record/no-dir/main-cwd | `wrk --list` from main repo cwd records main repo |
| 91 | projects/auto-record/no-dir/subdir-cwd | `wrk --list` from nested subdir records main repo |
| 92 | projects/auto-record/dir-arg/main-repo | `wrk <mainRepo> --list` records main repo |
| 93 | projects/auto-record/dir-arg/linked-worktree | `wrk <linkedWt> --list` records main repo, not worktree |
| 94 | projects/auto-record/dir-arg/nested-subpath | `wrk <nestedSubpath> --list` records main repo |
| 95 | projects/auto-record/non-git | non-git cwd → no project record |
| 96 | projects/auto-record/missing-dir | `wrk <nonexistent> --list` → no project record |
| 97 | projects/auto-record/fail-after-record | dirty `--done` fails but project auto-recorded + event logged |
| 98 | projects/list/projects/empty | `wrk --projects` empty → exit 0, no output |
| 99 | projects/list/projects/after-records | `wrk --projects` prints sorted detailed blocks after auto-record |
| 99a | projects/detailed-status/single-clean-no-wts | one project block with remote compare + `0 total, 0 dirty` |
| 99b | projects/detailed-status/with-linked-mixed | `Worktrees: 3 total, 1 dirty` |
| 99b2 | projects/detailed-status/stale-gitdir-linked | stale `.git` gitdir → `2 total, 0 dirty, 1 error` + detail line, exit 0 |
| 99b2a | projects/detailed-status/broken-main-repo | broken main repo → minimal `Dir` + `Status: error: ...` only, exit 0 |
| 99b2b | projects/detailed-status/prunable-worktrees | deleted checkout → `0 total, 0 dirty, 1 prune`, no per-path lines, exit 0 |
| 99c | projects/detailed-status/ahead-of-upstream | `Remote:` shows `needs push(+N commit)` |
| 99d | projects/detailed-status/no-upstream | `Remote: (no upstream)` |
| 99e | projects/detailed-status/multiple-projects | two lex-ordered blocks with blank separator |
| 99e2 | projects/output-streaming/fast-before-slow-gather | fast project stdout before slow project gather completes |
| 99f | projects/detailed-status/empty | empty projects → exit 0, no stdout |
| 99g | projects/color-output/no-color-pipe | pipe `--projects` → no ANSI, aligned `Worktrees:    ` |
| 99h | projects/color-output/force-color-dirty-status | `--color` → granular red/grey dirty status segments |
| 99h2 | projects/color-output/force-color-dirty-partial | `--color` → grey zero segments, red `2 changed` |
| 99i | projects/color-output/force-color-needs-push | `--color` → orange around `needs push(...)` |
| 99j | projects/color-output/force-color-needs-pull | `--color` → orange around `needs pull(...)` |
| 99k | projects/color-output/force-color-diverged | `--color` → red around `diverged(...)` |
| 99o | projects/remote-brief/ahead-of-upstream | plain `Remote: needs push(+1 commit)` |
| 99p | projects/remote-brief/behind-upstream | plain `Remote: needs pull(1 commit behind)` |
| 99q | projects/remote-brief/diverged | plain `Remote: diverged(2 commits)` |
| 99r | projects/remote-brief/up-to-date | plain `Remote: identical` |
| 99l | projects/color-output/force-color-worktrees-dirty | `--color` → red on `N dirty` only |
| 99l2 | projects/color-output/force-color-stale-gitdir-linked | `--color` → red on `1 error` summary + per-path `error: ...` |
| 99m | projects/color-output/clean-no-color | all clean + `--color` → no red/orange on values |
| 99n | projects/color-output/color-with-list | `--list --color` → git worktree list unchanged |
| 100 | projects/add/manual/main-repo | `wrk --add <mainRepo>` records + stdout path |
| 101 | projects/add/manual/linked-worktree | `wrk --add <linkedWt>` resolves to main repo |
| 102 | projects/add/idempotent | duplicate auto + manual → single entry (source stays auto) |
| 103 | projects/events/append-on-success | create appends event with `exit_code` 0 |
| 104 | projects/events/append-on-failure | failed command appends event with `exit_code` != 0 |
| 105a | projects/perf-profile/emits-events/many-worktrees | perf log JSONL with run/project/phase/worktree events for 12 wts |
| 105b | projects/perf-profile/budget/many-worktrees-parallel | parallel gather: worktree_status_all <100ms, run_end <200ms |
| 105c | projects/perf-profile/structure/dedup-list-linked | one list_linked phase per project (dedup ListLinked) |
| 105 | projects/invalid-mode/projects-with-list | `wrk --projects --list` → mutually exclusive error |
| 106 | projects/invalid-mode/add-missing-path | `wrk --add` without path → error |
| 106a | projects/remove/manual/main-repo | `--add` then `--rm <mainRepo>` → gone from json, stdout path |
| 106b | projects/remove/manual/linked-worktree | `--rm <linkedWt>` resolves to main repo, removes entry |
| 106c | projects/remove/idempotent/not-recorded | `--rm` never-recorded path → exit 0, empty stdout |
| 106d | projects/remove/idempotent/already-removed | remove twice → second call exit 0, empty stdout |
| 106e | projects/remove/missing-path-arg | `wrk --rm` without path → error |
| 106f | projects/remove/stale-path | record repo, delete `.git`, `--rm <old-path>` → removed |
| 106g | projects/remove/invalid-mode/remove-with-list | `wrk --rm X --list` → mutually exclusive error |
| 107 | projects/basename-fallback/single-match/create | Saved project; cwd elsewhere; `wrk myrepo` creates wt from saved path |
| 108 | projects/basename-fallback/cwd-exists/no-fallback | `./myrepo` exists in cwd (not git); `wrk myrepo` → git error, no fallback |
| 109 | projects/basename-fallback/no-match/error | No cwd entry, no saved project → `does not exist` |
| 110 | projects/basename-fallback/path-with-separator/no-fallback | `wrk sub/foo` missing → no fallback, normal error |
| 111 | projects/basename-fallback/ambiguous/tty-select | Two saved projects same basename; TTY + stdin selects one |
| 112 | projects/basename-fallback/ambiguous/non-tty | Two saved projects same basename; non-TTY → error listing candidates |
| 113 | projects/basename-fallback/other-mode/no-fallback | `wrk myrepo --list` with saved project → no fallback, `does not exist` |
| 113a | projects/basename-fallback/cwd-file-exists/single-match/guided-error | File `./myrepo` in cwd + one saved project → guided stderr with concrete path + `-t` hint |
| 113b | projects/basename-fallback/cwd-file-exists/ambiguous/guided-error | File `./spl` in cwd + two saved `spl` projects → guided stderr + `<full-path> --status` hint |
| 113c | projects/basename-fallback/cwd-file-exists/no-match/short-error | File `./foo` in cwd, no saved project → single-line `exists and is a file` only |
| 120 | status/basename-fallback/single-match/status | Saved project; neutral cwd; `wrk myrepo --status` → one clean block for saved root |
| 121 | status/basename-fallback/cwd-exists/no-fallback | `./myrepo` in cwd (not git); saved exists → `is not a git repository`, no fallback |
| 122 | status/basename-fallback/no-match/error | No cwd entry, no saved project → `does not exist` |
| 123 | status/basename-fallback/path-with-separator/no-fallback | `wrk saved/myrepo --status` missing → no fallback, normal error |
| 124 | status/basename-fallback/ambiguous/tty-select | Two saved projects same basename; TTY + stdin selects one → status for chosen repo |
| 125 | status/basename-fallback/ambiguous/non-tty | Two saved projects same basename; non-TTY → error listing candidates |
| 126 | where/single-match/basic | One saved `spl`; `wrk --where spl` → stdout saved abs path |
| 127 | where/ambiguous/two-matches | Two saved `spl`; `wrk --where spl` → both paths sorted, exit 0 |
| 128 | where/no-match/error | No saved `spl`; `wrk --where spl` → stderr no-match, empty stdout |
| 129 | where/non-basename/path-with-separator | Saved `spl`; `wrk --where sub/spl` → basename-only rejection |
| 130 | where/non-basename/absolute-path | Saved `spl`; `wrk --where <abs-path>` → basename-only rejection |
| 131 | where/empty-arg/error | `wrk --where` (no value) → requires argument |
| 132 | where/mutual-exclusion/with-status | Saved `spl`; `wrk --where spl --status` → mutually exclusive |
| 133 | where/cwd-exists/no-local-fallback | `./spl` in cwd (non-git) + saved `spl` → stdout saved path only |
| 134 | cd/in-place/abs-path | `WRK_FOLLOWUP_FILE` + `wrk --cd /abs` → empty stdout; follow-up `cd /abs` |
| 135 | cd/in-place/relative-path | in-place `wrk rel/target --cd` → follow-up expanded abs |
| 136 | cd/in-place/basename-expanded | in-place `wrk --cd myrepo` → follow-up uses saved abs not basename |
| 137 | cd/fallback/abs-path | channel closed; stdout abs; install hint; fake shell cwd=abs |
| 138 | cd/fallback/relative-path | channel closed; relative path resolves under cwd |
| 139 | cd/fallback/basename-single | one projects match → stdout saved abs + shell |
| 140 | cd/fallback/shell-exit-nonzero | fake shell exit 42 → wrk exit 42 |
| 141 | cd/syntax-forms/flag-then-path | `wrk --cd PATH` in-place success |
| 142 | cd/syntax-forms/path-then-flag | `wrk PATH --cd` in-place success (equivalent) |
| 143 | cd/resolution/missing-arg | `wrk --cd` → requires path argument |
| 144 | cd/resolution/no-match | basename 0 matches → does not exist |
| 145 | cd/resolution/missing-path | missing abs → does not exist |
| 146 | cd/resolution/not-a-directory | file path → not a directory |
| 147 | cd/resolution/ambiguous-non-tty | two saved basenames → error listing candidates |
| 148 | cd/resolution/local-basename | local `./myrepo` wins over projects entry |
| 149 | cd/mutual-exclusion/with-list | `--cd` + `--list` → mutually exclusive |
| 150 | cd/mutual-exclusion/with-no-cd | `--cd` + `--no-cd` → error |
| 151 | cd/mutual-exclusion/with-where | `--cd` + `--where` → mutually exclusive |
| 152 | cd/events/command-cd | success → `events.jsonl` `command: "cd"`, args include `--cd` |
| 153 | set-config/write/full-on | `--set-config --create --new-window --new-terminal --open-in-agent` → full defaults |
| 154 | set-config/write/new-window-implies-terminal | `--new-window` alone also writes `terminal.mode=new` |
| 155 | set-config/write/merge/terminal-then-agent | sequential terminal then agent writes; both present |
| 156 | set-config/write/merge/preserve-extra-key | top-level `extra: 1` preserved after set-config |
| 157 | set-config/write/negatives/no-open-in-agent | agent.enabled=false after prior on |
| 158 | set-config/write/negatives/no-new-window | window cleared/off |
| 159 | set-config/show/prints-json | `--set-config --show` prints JSON |
| 160 | set-config/mutual-exclusion/with-list | `--set-config … --list` → non-zero |
| 161 | set-config/mutual-exclusion/with-create-dir | `wrk <dir> --set-config …` → non-zero; no worktree |
| 161b | set-config/mutual-exclusion/with-no-config | `--no-config --set-config --show` → non-zero; no write |
| 162 | create-ux/bare/empty-config | bare create; no space/iterm/agent |
| 163 | create-ux/pipeline/flags/new-window-only | space + iterm ForceNew; no agent |
| 164 | create-ux/pipeline/flags/new-terminal | iterm ForceNew; no space |
| 165 | create-ux/pipeline/flags/reuse-terminal | ModeReuseCurrent script |
| 166 | create-ux/pipeline/flags/smart-terminal | ModeSmart script |
| 167 | create-ux/pipeline/flags/open-in-agent-only | agent-run in-process; cwd=wt |
| 168 | create-ux/pipeline/flags/terminal-plus-agent | iterm follow-up only; outer agent not exec'd |
| 169 | create-ux/pipeline/flags/full-pipeline | window + terminal + agent follow-up |
| 170 | create-ux/pipeline/flags/with-exec | UX then `--exec pwd` |
| 171 | create-ux/pipeline/config/defaults-match-flags | config-only full UX matches flags |
| 172 | create-ux/pipeline/config/no-open-in-agent-override | config agent on + `--no-open-in-agent` |
| 173 | create-ux/errors/new-window-no-terminal | `--new-window --no-new-terminal` → error |
| 174 | create-ux/errors/mutual-terminal-flags | two terminal mode flags → error |
| 175 | create-ux/errors/non-darwin-window | mocked linux GOOS → platform error |
| 176 | create-ux/interceptor-ignored/native-create | leftover interceptor ignored; native create |
| 177 | create-ux/agent-quoting/adversarial-task-quotes | argv-safe prompt for agent-in-process |
| 178 | create-ux/agent-quoting/terminal-followup-quotes | shell-safe prompt in iterm follow-up |
| 178a | create-ux/agent-full-task/name-budget-trim | long basename+task fitted; agent-run prompt = full taskDesc |
| 179a | create-ux/target-dir-config-skipped/config-ignored | full config + SpawnDir; no CLI UX → mocks silent |
| 179b | create-ux/target-dir-config-skipped/flags-still-apply | empty config + SpawnDir + full CLI UX → pipeline runs |
| 179c | create-ux/target-dir-config-skipped/flag-only-no-config-agent | config agent on + SpawnDir + `--new-terminal` only → no agent |
| 179d | create-ux/no-config/config-ignored | full config + `--no-config`; no UX flags → mocks silent |
| 179e | create-ux/no-config/flags-still-apply | full config + `--no-config` + full CLI UX → flags drive pipeline |
| 179f | create-ux/no-config/corrupt-ignored | corrupt config.json + `--no-config` → no parse error; bare create |
| 179 | exec/create/basic-pwd | create + `--exec pwd` → path then pwd in new wt |
| 180 | exec/create/args-passthrough | `--exec echo --task` → child sees `--task`; no task slug on path |
| 190 | exec/cd/with-followup | `--cd` + follow-up + `--exec pwd` → follow-up + stdout pwd |
| 191 | exec/bring/basic-pwd | `--bring` + `--exec pwd` → external path then pwd there |
| 192 | exec/set-task/after-rename | `--set-task` + `--exec pwd` → new path then pwd |
| 193 | exec/done/exec-on-main | `--done -y --exec pwd` → last line = main repo path |
| 194 | exec/reject/with-list | `--list --exec true` → non-zero |
| 195 | exec/reject/with-status | `--status --exec true` → non-zero |
| 196 | exec/empty-flag/bare-exec | bare `--exec` → requires command; no wt |
| 197 | exec/empty-flag/equals-form | `--exec=pwd` → reject equals form |
| 198 | stderr-newline | unrecognized flag → exit 1; stderr body + trailing `\n` |
| 199 | web/mutual-exclusion/with-list | `wrk --web --list` → non-zero; mutually exclusive; empty stdout |
| 200 | web/mutual-exclusion/with-status | `wrk --web --status` → non-zero; mutually exclusive; empty stdout |
| 201 | web/port-without-web | `wrk --port 18080` → non-zero; `--port is only valid with --web` |
| 202 | web/unexpected-args | `wrk --web some-dir` → non-zero; unexpected arguments |
| 203 | web/help-mentions-web | `wrk -h` → exit 0; help mentions `--web` and `--port` |
| 204 | web/serves-page | `wrk --web --port <free>`; GET `/` → 200 HTML markers; stdout listen URL + `\n` |
| 205 | web/api-projects-empty | same server; GET `/api/wrk/projects` empty WRK_HOME → 200 `{"projects":[]}` |
| 206 | web/mockup-repo-view | same server; GET `/mockup/repo-view` → 200 SPA shell (client route fallback) |
| 207 | scan-git-repos/record/basic-add | one main under root; stdout abs path; projects empty |
| 208 | scan-git-repos/record/idempotent | pre-seeded scan entry; second scan → still one entry; always-print |
| 209 | scan-git-repos/record/main-only | main + linked wt under root; only main printed; projects empty |
| 210 | scan-git-repos/record/with-no-cache | `--scan-git-repos --no-cache ROOT` still discovers + prints |
| 211 | scan-git-repos/mutual-exclusion/with-projects | `--scan-git-repos --projects` → mutually exclusive |
| 212 | scan-git-repos/no-cache-without-scan | bare `--no-cache` → only valid with `--scan-git-repos` |
| 213 | scan-git-repos/help-mentions-scan-git-repos | `wrk -h` → mentions `--scan-git-repos`, `--no-cache`, default root `~` |
| 214 | scan-git-repos/streaming/discovery-order | multi-root CLI order on stdout (not lex path sort); projects empty |
| 215 | scan-git-repos/streaming/first-path-before-finish | `label:slow`; first path bytes before process exit (pad walk) |
| 216 | scan-git-repos/interrupt/sigint-after-first-path | SIGINT after first path → exit 130; `warning:`; projects unchanged |
| 217 | scan-git-repos/defaults/home-root | bare `--scan-git-repos`; FakeHome=WorkRoot + `repo-a`; no Projects → scan home |
| 218 | scan-git-repos/defaults/home-unusable | bare `--scan-git-repos`; FakeHome missing → non-zero; no Projects error |
| 219 | scan-git-repos/debug/via-verbose | seed then `--scan-git-repos -v ROOT`; stderr `scan:` + `mode=` (FakeHome cache) |
| 220 | scan-git-repos/debug/via-env | seed then `WRK_SCAN_DEBUG=1` without `-v`; stderr `scan:` + `mode=` |
| 221 | scan-git-repos/debug/off | seed then quiet second scan; zero `scan:` markers |
| 222 | scan-git-repos/filter-home-subpath/default-home-universe | bare scan; product `home/repos.json` universe=home lists home mains |
| 223 | scan-git-repos/filter-home-subpath/projects-shares-home-cache | `~/Projects` reuses home universe cache; stdout only under Projects |
| 224 | scan-git-repos/filter-home-subpath/debug-cache-base-filter | `-v` Projects root → stderr `cache_base` + `filter` |


| R1 | bring/removed-flags/unknown-dep | `wrk --dep` → non-zero unknown flag (RED pre-implement) |
| R2 | bring/removed-flags/unknown-all-deps | `wrk --all-deps` → non-zero unknown flag (RED pre-implement) |
| R3 | bring/removed-flags/dry-run-host-list-no-all-deps | bare `--dry-run` host list excludes `--all-deps` (RED) |
| R4 | bring/removed-flags/help-no-removed-flags | `-h` has `--bring`; no `--dep`/`--all-deps` mode lines (RED) |
| R5 | bring/no-dep/invalid/bare | `--no-dep` alone → only valid with `--bring` (RED until error text drops old hosts) |
| M1 | bring/dep-sub-module | dep module in subdir → replace at sub (migrated from dep/) |
| M2 | bring/consumer-multi-module | multi consumer modules → replace in each (migrated) |
| M3 | bring/basename-fallback/single-match/basic | `--bring mydep` basename fallback (migrated) |
| M4 | bring/branch-collision-suffix | preferred branch taken → `-1` (migrated) |
| M5 | exec/bring/basic-pwd | `--bring` + `--exec pwd` |
| M6 | bash-integration/complete/bring | complete after `--bring` basenames |
| M7 | target-dir/with-other-mode/with-bring | target-dir + `--bring` → unexpected arguments |

## How to Run

```sh
# Verify tree structure (no test execution)
doctest vet ./tests

# Fast discovery run (skips labeled leaves — e2e / slow / tty / flaky)
doctest test ./tests

# True binary integration leaves (process boundary; pilot: dir-arg, list, sync, …)
doctest test --label e2e ./tests

# Slow / perf leaves only (often also e2e; 12-worktree perf, multi-repo --projects, …)
doctest test --label slow ./tests

# Full CI: discovery then e2e (and optionally slow/tty)
doctest test ./tests && doctest test --label e2e ./tests

# Flaky timing budget (subset of slow)
doctest test --label flaky ./tests/projects/perf-profile/budget/many-worktrees-parallel

# Run a specific leaf
doctest test ./tests/create-worktree/main-checkout/basic-create

# Run a done leaf
doctest test ./tests/done/ahead-confirm

# Composition: flag matrix, push, done/merge-back post-pipeline (+ dry-run)
doctest vet ./tests/done-compose
doctest test ./tests/done-compose
doctest test ./tests/done-compose/help
doctest test ./tests/done-push
doctest test ./tests/done-pipeline
doctest test ./tests/done-pipeline/dry-run
# P2 done UX phases + structured cascade errors (Classic RED until implementer)
doctest vet ./tests/done-output
doctest test ./tests/done-output
doctest test ./tests/merge-back-pipeline
doctest test ./tests/merge-back-pipeline/dry-run

# Run a list leaf
doctest test ./tests/list/main-only

# Run status leaves
doctest test ./tests/status
doctest test ./tests/status/valid-git-cwd/dirty-counts
# Untracked-as-added leaves (expect RED until PorcelainUntracked includes ?? for --status)
doctest test ./tests/status/valid-git-cwd/untracked-dirty
doctest test ./tests/status/valid-git-cwd/nested-untracked-dirty
doctest vet ./tests/status/master-field
doctest test ./tests/status/master-field
doctest vet ./tests/status/color-output
doctest test ./tests/status/color-output

# Always-new-branch + naming policy (expect RED until wrkcli create/set-task/external changes land)
doctest vet ./tests/create-worktree
doctest test ./tests/create-worktree/main-checkout/slash-branch
doctest test ./tests/create-worktree/main-checkout/branch-collision
doctest test ./tests/create-worktree/main-checkout/sequence-increment
doctest vet ./tests/fetch-and-verbose/verbose/create
doctest test ./tests/fetch-and-verbose/verbose/create/branch-collision

# Run --status basename-fallback leaves (expect RED until resolveSourceWorkDir enables status fallback)
doctest vet ./tests/status/basename-fallback
doctest test ./tests/status/basename-fallback
doctest test ./tests/status/basename-fallback/single-match/status
doctest test ./tests/status/basename-fallback/ambiguous/tty-select

# Run nested-broken-linked leaves (GREEN: non-fatal broken + P3 gray header when --color)
doctest vet ./tests/status/nested-broken-linked
doctest test ./tests/status/nested-broken-linked
doctest test ./tests/status/nested-broken-linked/stale-gitdir
doctest test ./tests/status/nested-broken-linked/color

# Run main-repo-worktrees (GREEN: primary = main + ListLinked; WRK linked not external section)
doctest vet ./tests/status/main-repo-worktrees
doctest test ./tests/status/main-repo-worktrees
doctest test ./tests/status/main-repo-worktrees/external-clean

# Run section-order + section-partition (GREEN: P1 partition + P2 primary/external CLI order)
doctest vet ./tests/status/section-partition
doctest test ./tests/status/section-partition
doctest vet ./tests/status/section-order
doctest test ./tests/status/section-order
doctest test ./tests/status/color-output/force-color-header
doctest test ./tests/status/color-output/no-color-header

# Run a dep leaf

# Run --bring best-effort leaves (expect RED until --bring implemented)
doctest vet ./tests/bring
doctest test ./tests/bring
doctest test ./tests/bring/basic
doctest test ./tests/bring/not-a-dependency
doctest test ./tests/bring/exec-after-skip

# Run --no-dep + -v tidy/worktree stream leaves (expect RED until --no-dep / verbose tidy/stream landed)
doctest vet ./tests/bring/no-dep ./tests/bring/verbose ./tests/bring/removed-flags
doctest test ./tests/bring/no-dep/...
doctest test ./tests/bring/verbose/...
doctest test ./tests/bring/help-mentions-no-dep

# Run --bring basename-fallback leaves (expect RED until resolveDirArg wired in runDep)

# Run an all-deps leaf

# Run a dry-run leaf

# Run yes-flag / cascade guard leaves (expect RED until -y + option A implemented)
doctest vet ./tests/yes-flag
doctest test ./tests/yes-flag
doctest test ./tests/done/cascade-force-removal
doctest test ./tests/done/cascade-non-tty-rejects-with-confirm-from-stdin

# Run --exec cut-flag leaves (expect RED until less-flags Cut + wrk --exec implemented)
doctest vet ./tests/exec
doctest test ./tests/exec
doctest test ./tests/exec/create/basic-pwd
doctest test ./tests/exec/done/exec-on-main
doctest test ./tests/exec/empty-flag/bare-exec

# Run a done cascade leaf
doctest test ./tests/done/external-cascade
doctest test ./tests/done/cascade-non-external-linked
doctest test ./tests/done/cascade-external-and-deps

# Run a local-replace guard leaf
doctest test ./tests/done/local-replace-blocks
doctest test ./tests/done/sub-module-replace-blocks

# Run an intra-replace (lenient/strict) leaf
doctest test ./tests/done/intra-replace-warns
doctest test ./tests/done/intra-replace-cross-worktree
doctest test ./tests/done/intra-replace-strict-blocks

# Run the --no-in-module-replace validation leaf
doctest test ./tests/done/no-in-module-replace-without-done

# Run a dir-arg leaf
doctest test ./tests/dir-arg/create/basic
doctest test ./tests/dir-arg/list/from-dir
doctest test ./tests/dir-arg/missing-dir

# Run a target-dir leaf
doctest vet ./tests/target-dir
doctest test ./tests/target-dir/target-missing/parent-exists/basic
doctest test ./tests/target-dir/target-missing/parent-exists/branch-collision
doctest test ./tests/target-dir/target-exists/collision-suffix
doctest test ./tests/target-dir/relative-path

# Same-repo reuse (Policy A auto + Policy B named) — expect RED until implemented
doctest vet ./tests/bring/reuse-same-repo
doctest test ./tests/bring/reuse-same-repo
doctest vet ./tests/target-dir/reuse-same-repo
doctest test ./tests/target-dir/reuse-same-repo
doctest test --label tty ./tests/target-dir/reuse-same-repo/existing-linked/tty

# Run a task spawn leaf
doctest test ./tests/task/spawn/basic
doctest test ./tests/task/spawn/t-alias
doctest test ./tests/task/spawn/empty-task
doctest test ./tests/task/spawn/sequence
doctest test ./tests/task/spawn/long-task
doctest test ./tests/task/spawn/name-budget

# Run a task set-task leaf (non-TTY, expects error)
doctest test ./tests/task/set-task/non-tty
doctest test ./tests/task/set-task/empty-desc
doctest test ./tests/task/set-task/not-linked
doctest test ./tests/task/set-task/fixed-path-unsupported
doctest test ./tests/task/set-task/path-collision-suffix
doctest test ./tests/task/set-task/branch-collision-suffix
doctest test ./tests/task/set-task/legacy-slash-migrate
doctest test ./tests/task/set-task/name-budget-fit

# Run a set-task with-dir leaf
doctest test ./tests/task/set-task/with-dir/rename-succeeds
doctest test ./tests/task/set-task/with-dir/missing-dir

# Forgot -t / task-like positionals (expect RED until implemented)
doctest vet ./tests/forgot-task-flag
doctest test ./tests/forgot-task-flag
doctest test ./tests/forgot-task-flag/two-arg/task-like/non-tty
doctest test ./tests/forgot-task-flag/two-arg/task-like/confirm-y
doctest test ./tests/forgot-task-flag/two-arg/task-like/yes-flag
doctest test ./tests/forgot-task-flag/one-arg

# Name budget + agent full taskDesc (expect RED until implemented)
doctest test ./tests/create-ux/agent-full-task/name-budget-trim

# Run projects leaves (expect RED until project persistence is implemented)
doctest vet ./tests/projects
doctest test ./tests/projects
doctest test ./tests/projects/auto-record/no-dir/main-cwd
doctest test ./tests/projects/list/projects/after-records
doctest vet ./tests/projects/detailed-status
doctest test ./tests/projects/detailed-status
doctest vet ./tests/projects/remote-brief
doctest test ./tests/projects/remote-brief
doctest vet ./tests/projects/color-output
doctest test ./tests/projects/color-output
doctest test ./tests/projects/add/manual/main-repo
doctest vet ./tests/projects/remove
doctest test ./tests/projects/remove
doctest test ./tests/projects/remove/manual/main-repo
doctest test ./tests/projects/remove/idempotent/already-removed
doctest test ./tests/projects/events/append-on-success

# Run basename-fallback leaves (expect RED until basename fallback is implemented)
doctest vet ./tests/projects/basename-fallback
doctest test ./tests/projects/basename-fallback
doctest test ./tests/projects/basename-fallback/single-match/create
doctest test ./tests/projects/basename-fallback/ambiguous/tty-select

# Run cwd-file-exists guided-error leaves (expect RED until file-collision hint is implemented)
doctest vet ./tests/projects/basename-fallback/cwd-file-exists
doctest test ./tests/projects/basename-fallback/cwd-file-exists
doctest test ./tests/projects/basename-fallback/cwd-file-exists/single-match/guided-error
doctest test ./tests/projects/basename-fallback/cwd-file-exists/ambiguous/guided-error
doctest test ./tests/projects/basename-fallback/cwd-file-exists/no-match/short-error

# Run --where leaves (expect RED until runWhere is implemented)
doctest vet ./tests/where
doctest test ./tests/where
doctest test ./tests/where/single-match/basic
doctest test ./tests/where/ambiguous/two-matches

# Run --cd leaves (expect RED until runCd + shell/interactive wired)
doctest vet ./tests/cd
doctest test ./tests/cd
doctest test ./tests/cd/in-place
doctest test ./tests/cd/fallback
doctest test ./tests/cd/in-place/abs-path
doctest test ./tests/cd/fallback/abs-path
doctest test ./tests/cd/fallback/shell-exit-nonzero
doctest test ./tests/cd/resolution/missing-arg
doctest test ./tests/cd/events/command-cd

# Run set-config + create-ux leaves (expect RED until create UX / --set-config implemented)
doctest vet ./tests/set-config
doctest vet ./tests/create-ux
doctest test ./tests/set-config
doctest test ./tests/create-ux
doctest test ./tests/set-config/write/full-on
doctest test ./tests/create-ux/bare/empty-config
doctest test ./tests/create-ux/pipeline/flags/full-pipeline
doctest test ./tests/create-ux/errors/new-window-no-terminal

# Run --web leaves (expect RED until wrk --web + wrkcli page serve is implemented)
doctest vet ./tests/web
doctest test ./tests/web
doctest test ./tests/web/mutual-exclusion/with-list
doctest test ./tests/web/serves-page
doctest test ./tests/web/mockup-repo-view
doctest test ./tests/web/api-projects-empty

# Run --scan-git-repos leaves (record + defaults + filter-home-subpath + debug + streaming + interrupt; skip labeled slow by default)
doctest vet ./tests/scan-git-repos
doctest test ./tests/scan-git-repos
doctest test ./tests/scan-git-repos/defaults
doctest test ./tests/scan-git-repos/defaults/home-root
doctest test ./tests/scan-git-repos/defaults/home-unusable
doctest test ./tests/scan-git-repos/record/basic-add
doctest test ./tests/scan-git-repos/record/idempotent
doctest test ./tests/scan-git-repos/record/main-only
doctest test ./tests/scan-git-repos/record/with-no-cache
doctest test ./tests/scan-git-repos/filter-home-subpath
doctest test ./tests/scan-git-repos/filter-home-subpath/default-home-universe
doctest test ./tests/scan-git-repos/filter-home-subpath/projects-shares-home-cache
doctest test ./tests/scan-git-repos/filter-home-subpath/debug-cache-base-filter
doctest test ./tests/scan-git-repos/debug
doctest test ./tests/scan-git-repos/debug/via-verbose
doctest test ./tests/scan-git-repos/debug/via-env
doctest test ./tests/scan-git-repos/debug/off
doctest test ./tests/scan-git-repos/streaming/discovery-order
doctest test ./tests/scan-git-repos/interrupt/sigint-after-first-path
doctest test ./tests/scan-git-repos/mutual-exclusion/with-projects
doctest test ./tests/scan-git-repos/no-cache-without-scan
doctest test ./tests/scan-git-repos/help-mentions-scan-git-repos
# Slow pad-walk streaming timing leaf (label: slow)
doctest test --label slow ./tests/scan-git-repos/streaming/first-path-before-finish
```



```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

type Request struct {
	WorkRoot   string
	WrkHome    string
	RepoDir    string   // process cwd when running wrk
	TargetDir  string   // optional first positional <dir>; prepended to Args when set
	SpawnDir   string   // optional second positional <target-dir>; appended after TargetDir when set
	HashToken  string   // detached-head: 7-char short commit hash
	Args       []string // CLI args after <dir>; empty + no TargetDir/TaskDesc → bare no-args (dashboard); create uses --new or positionals/task
	StdinInput string   // piped to stdin when set
	MainRepo      string // done tests: main checkout path
	WtDir         string // done/dep tests: linked worktree path
	WtBranch      string // done tests: worktree branch name
	Wt2Dir        string // done-sync/merge-back-sync: second linked worktree path
	Wt2Branch     string // done-sync/merge-back-sync: second linked worktree branch
	OriginBare    string // done-push/tag-next push fixtures: bare origin path
	DepPath          string // dep tests: path-to-repo argument
	ConsumerTop      string // dep tests: consumer git toplevel
	ConsumerModDir   string // dep tests: consumer go.mod directory (may differ from repo root for sub-modules)
	ConsumerModDir2  string // dep tests: second consumer go.mod directory for multi-module tests
	ExternalWtDir    string // dep/done tests: external worktree path
	DepsLinkedWtDir  string // done tests: manual linked worktree under deps/ (or other non-external path)
	DepsDepPath      string // done tests: dep main repo that owns DepsLinkedWtDir
	DepModulePath    string // dep tests: module path from dep go.mod
	TaskDesc           string // task tests: task description passed to --task or -t
	TaskFlag           string // task tests: flag form for TaskDesc ("-t" or "--task"; default "--task")
	SetTaskDesc        string // task tests: new task description for --set-task
	SetTaskEnv         string // task tests: extra env vars for --set-task (e.g., WRK_SET_TASK_CONFIRM=1)
	OldExternalGitdir  string // propagate tests: old gitdir content before rename
	ExternalWtDir2    string // propagate tests: second external worktree path
	SecondRepo         string // projects tests: second main repo path
	BasenameEnv        string // basename-fallback tests: e.g. WRK_BASENAME_CONFIRM=1
	SelectedSavedRepo  string // basename-fallback tty-select: chosen saved project path
	ProjectsPerfLog    string // perf-profile tests: WRK_PROJECTS_PERF_LOG path
	FakeHome           string // git-lfs-hook tests: temp home with .local/bin/git-lfs
	UseMinimalPath     bool   // git-lfs-hook tests: run wrk with PATH=/usr/bin:/bin
	UseScriptTTY       bool   // yes-flag tests: run wrk under `script` fake TTY (darwin/linux)

	// --cd / follow-up channel / fake interactive shell (cd/ leaves)
	FollowupFile   string // path for WRK_FOLLOWUP_FILE content checks
	UseFollowupEnv bool   // export WRK_FOLLOWUP_FILE to wrk (in-place channel open)
	FakeShellDir   string // bin dir prepended to PATH containing fake "bash"
	FakeShellLog   string // path where fake bash records cwd/args
	FakeShellExit  int    // exit code of fake bash (default 0; set via env for non-zero)
	ShellEnv       string // when set, export SHELL=<value> (detect.Shell basename)

	// PathPrepend / ExtraEnv / InterceptorLog — create-ux (fake agent-run + mock env) and shared harness
	PathPrepend    string   // bin dir prepended to PATH (fake agent-run / tools)
	ExtraEnv       []string // additional KEY=VAL env entries for wrk (UX mocks, etc.)
	InterceptorLog string   // optional path for fake argv logs (create-ux uses WorkRoot helpers)

	// --web long-running server probe (web/ leaves)
	WebProbe bool   // when true: start wrk, wait for listen URL, GET WebPath, kill process
	WebPath  string // HTTP path to probe (default "/"); only used when WebProbe

	// InProcess runs via wrkcli.Capture (L2 short path) instead of the product binary.
	// Prefer for help / mutual-exclusion / early reject leaves that do not need a
	// process boundary. Leave false (default) for true L3 e2e integration.
	InProcess bool
}

type Response struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	HTTPStatus int    // set when WebProbe; status of GET WebPath
	HTTPBody   string // set when WebProbe; response body of GET WebPath
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	adoptDoctestContext(d)
	args := buildWrkCLIArgs(req)

	if err := prepareFollowupFile(req); err != nil {
		return nil, err
	}

	if req.InProcess {
		return runWrkInProcess(req, args)
	}

	bin := getWrkBin(t)

	if req.UseScriptTTY {
		return execScriptTTYWrk(t, req, bin, args)
	}

	if req.WebProbe {
		return runWebProbe(t, req, bin, args)
	}

	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = wrkEnv(req)
	return captureCommandOutput(cmd, req.StdinInput)
}

// runWrkInProcess is L2 short-path execution via wrkcli.Capture (no product binary).
// Mirrors wrkEnv extras needed for demoted leaves (ExtraEnv, PathPrepend, follow-up,
// SetTask/Basename, FakeHome) without os.Environ(). Skip UseScriptTTY / WebProbe /
// UseMinimalPath — those stay binary e2e.
func runWrkInProcess(req *Request, args []string) (*Response, error) {
	env := []string{"WRK_HOME=" + req.WrkHome, "WRK_DATE=" + wrkDate}
	if req.FakeHome != "" {
		env = append(env, "HOME="+req.FakeHome)
	}
	if req.SetTaskEnv != "" {
		env = append(env, req.SetTaskEnv)
	}
	if req.BasenameEnv != "" {
		env = append(env, req.BasenameEnv)
	}
	if req.ProjectsPerfLog != "" {
		env = append(env, "WRK_PROJECTS_PERF_LOG="+req.ProjectsPerfLog)
	}
	env = appendExtraEnv(env, req)
	env = appendCDEnv(env, req)
	res := wrkcli.Capture(wrkcli.CaptureOpts{
		Args:  args,
		Dir:   req.RepoDir,
		Env:   env,
		Stdin: req.StdinInput,
	})
	return &Response{
		Stdout:   res.Stdout,
		Stderr:   res.Stderr,
		ExitCode: res.ExitCode,
	}, nil
}

// prepareFollowupFile truncates FollowupFile so in-place --cd leaves can detect writes.
func prepareFollowupFile(req *Request) error {
	if req.FollowupFile == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(req.FollowupFile), 0o755); err != nil {
		return err
	}
	return os.WriteFile(req.FollowupFile, nil, 0o644)
}
```
