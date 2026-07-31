---
name: wrk
description: >-
  Git worktree helper for isolated feature branches. Use when creating
  linked worktrees, merging back, checking status, linking deps, or
  looking up registered projects by basename.
---

# wrk

WRK_SKILL_DOCTEST_MARKER

## When to use

Use wrk when you need an isolated git worktree without disturbing your main
checkout. Typical flow: open the **dashboard** (bare `wrk`) or create with
`wrk --new`, work in the worktree, finish with `--done` or `--merge-back`.

## Dashboard

```sh
wrk                              # dashboard mode (does not create a worktree)
```

Bare no-args `wrk` opens the **dashboard** stage snapshot (Pre / Main / After).
It does **not** create a worktree. Interactive TTY / hermetic RUN can compose
`--done` / `--merge-back` pipelines; cancel leaves the tree unchanged.
Hint in the UI: create with `wrk --new`.

## Create

```sh
wrk --new                        # create a worktree from cwd (explicit create entry)
wrk --new -t 'fix login bug'    # create with task slug in branch/dir names
wrk myrepo -t 'fix login bug'   # create via basename + task (no --new required)
wrk myrepo                       # create from registered/local dir positional
```

**`--new`** is the explicit create entry (former bare create). Create also runs
when create-selecting args are present (`<dir>`, `-t` / `--task`, create UX flags).
Bare `wrk` alone is **dashboard**, not create.

`-t` / `--task` appends a task slug to worktree and branch names.

## Finish

```sh
wrk --done                       # merge back and remove worktree
wrk --merge-back                 # merge back without removing
wrk --done --sync --tag-next --push   # after success: sync, tag, push main (+ tags)
wrk --done --tag-next --propagate-tags   # after success: tag then bump consumer require lines
wrk --merge-back --tag-next --push --dry-run  # plan post-pipeline only
wrk --tag-next --propagate-tags --dry-run     # plan tag then consumer bumps
wrk --propagate-tags --dry-run                # plan consumer go.mod bumps from source release tags
```

### Compose pipeline (fixed stage order; flag order free)

```text
# [pre]  --gen-commit-msg --commit [--model …]   # requires --commit with primary
# [main] --done | --merge-back
# [post] --sync → --tag-next → --push → --propagate-tags
# [tail] --reinstall-local   # from main tip after merge; empty plan succeeds
```

Fluent recipes:

```sh
wrk --done --sync --tag-next --push
wrk --done --sync --tag-next --push --reinstall-local
wrk --gen-commit-msg --commit --model=MODEL --done --sync --tag-next --push
# full:
wrk --gen-commit-msg --commit --model=MODEL --done --sync --tag-next --push --reinstall-local
# optional: --confirm restores Y/n; -y still auto-yes
```

Optional pre-stage on `--done` / `--merge-back`: `--gen-commit-msg --commit …` (on the source worktree; with primary, `--commit` is required; `--dir` is not valid when composed).
Optional post-modifiers: `--sync`, `--tag-next`, `--push`, `--propagate-tags`, `--reinstall-local`, `--dry-run`.
Post-pipeline order is fixed: **sync → tag-next → push → propagate-tags → reinstall-local** (then exec/land on `--done` only).
`--propagate-tags` bumps consumer `go.mod` require versions to source release tags (compose with `--tag-next` to use newly planned/created tags; alone uses existing source tags).
`--push` with a primary pushes the main branch (and tags when combined with `--tag-next`).
`--reinstall-local` after a successful primary scans modules from the main tip (empty reinstall plan still succeeds).
`--json` is only for bare `--tag-next`, not with `--done` / `--merge-back` or `--propagate-tags`.

## Pull request

Multi-mode from a linked worktree (requires `gh` and a github.com origin):

```sh
wrk --pr                                          # show open PR URL (or empty)
wrk --pr --status                                 # open PR metadata + checks/reviews rollup
wrk --pr --comment 'Follow-up note'               # comment-only on existing open PR
wrk --pr --push                                   # full-push tip when open PR exists (error if none)
wrk --pr --push --comment 'Pushed tip'            # push then comment on open PR
wrk --pr --title 'Fix login' --comment 'Details'  # create or attach
wrk --push --pr --title 'Fix login' --comment 'Details'
wrk --gen-commit-msg --commit --model=MODEL --push --pr --title 'Fix login' --comment 'Details'
```

**`--pr`** modes (flag combinations; not every mode needs title+comment):
- **Show** (bare `--pr`): print open PR URL for the current branch, or empty if none.
- **Status** (`--pr --status`): open PR metadata + checks/reviews rollup (not with `--title`/`--comment`/`--push`).
- **Comment-only** (`--pr --comment C`, no `--title`): add a comment on an existing open PR (error if none).
- **Push-existing** (`--pr --push` without `--title`): require open PR, full-push branch tip, print URL; optional `--comment` after push.
- **Create/attach** (`--pr --title T --comment C`): both required and non-empty for this mode. Ensures remote head branch (push only if missing). New PR: title + comment as **initial body**. Existing PR: `--title` ignored; `--comment` is additive.

Compose: `[--gen-commit-msg --commit …] [--push] --pr` (order: gen-commit → push → pr).
With `--push` on create/attach compose, always full-pushes the branch tip before the PR path.

## Inspect

```sh
wrk --status                     # status for git repos under this checkout
wrk -l                           # list worktrees (alias: --list)
wrk --projects                   # recorded main repository paths
wrk --projects --github          # only projects whose origin is github.com
wrk --projects-dep-graph         # module-level dep graph across registered projects
wrk --where <basename>           # look up saved project path(s) by basename
wrk --main                       # nested shell at main repository root
```

## Dependencies

```sh
wrk --bring <path>               # spawn a dependency worktree under ./external; soft-skip replace when not a module dep
wrk --bring <path> --no-dep      # worktree only; skip replace and tidy
```

## Gotchas

- **Replace guard**: `--done` may refuse when `go.mod` has local `replace` directives;
  use `--no-in-module-replace` for strict blocking.
- **Basename fallback**: `wrk myrepo` resolves via registered projects when cwd has no
  local `./myrepo` directory.
- **Default auto-yes**: bare `--done` / `--merge-back` / `--set-task` skip plan
  prompts (including cascade deps). Use `--confirm` for interactive Y/n;
  `-y` remains valid for compatibility.