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
```

## Inspect

```sh
wrk --status                     # status for git repos under this checkout
wrk -l                           # list worktrees (alias: --list)
wrk --projects                   # recorded main repository paths
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