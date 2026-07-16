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
checkout. Typical flow: create a worktree, work in it, finish with `--done` or
`--merge-back`.

## Create

```sh
wrk                              # create from cwd repo
wrk myrepo -t 'fix login bug'    # basename + task slug in branch/dir names
```

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

Optional post-modifiers on `--done` / `--merge-back`: `--sync`, `--tag-next`, `--push`, `--propagate-tags`, `--dry-run`.
Post-pipeline order is fixed: **sync → tag-next → push → propagate-tags** (then exec/land on `--done` only).
`--propagate-tags` bumps consumer `go.mod` require versions to source release tags (compose with `--tag-next` to use newly planned/created tags; alone uses existing source tags).
`--push` with a primary pushes the main branch (and tags when combined with `--tag-next`).
`--json` is only for bare `--tag-next`, not with `--done` / `--merge-back` or `--propagate-tags`.

## Inspect

```sh
wrk --status                     # status for git repos under this checkout
wrk -l                           # list worktrees (alias: --list)
wrk --projects                   # recorded main repository paths
wrk --projects-dep-graph         # module-level dep graph across registered projects
wrk --where <basename>           # look up saved project path(s) by basename
wrk --main                       # nested shell at main repository root
```

## Dependencies

```sh
wrk --dep <path>                 # spawn a dependency worktree under ./external
wrk --bring <path>               # like --dep; soft-skip replace when not a module dep
wrk --all-deps                   # link required deps from registered projects
wrk --all-deps --dry-run         # plan only, no writes
```

## Gotchas

- **Replace guard**: `--done` may refuse when `go.mod` has local `replace` directives;
  use `--no-in-module-replace` for strict blocking.
- **Basename fallback**: `wrk myrepo` resolves via registered projects when cwd has no
  local `./myrepo` directory.
- **Non-TTY cascade**: `-y` auto-confirms on a TTY; in scripts prefer explicit flags.