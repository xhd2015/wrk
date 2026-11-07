---
name: edit-dependency-code
description: >-
  Editing another project's source always starts with wrk --bring <repo> from the
  primary consumer root: it reuses or creates a linked worktree under ./external/,
  repoints the go.mod replace at it, and prints the one SSOT path. Edit only that
  path; never the dependency's main worktree (~/Projects/...). Use when modifying
  dependency code, sibling/upstream repos, registered wrk projects (e.g. dot-pkgs),
  go.mod replace targets outside ./external/, or any tree outside the current
  checkout. Triggers: edit dependency code, other project, upstream package,
  out-of-tree, registered project, sibling module, wrk --bring, bring into scope.
  Slash: /edit-dependency-code. Do not use for in-repo-only edits, or version-only
  bumps with no source edit in the dependency.
---

# Edit dependency code

Dependency source is edited in exactly one place: the linked worktree that
`wrk --bring` prints under the primary consumer's `./external/`.

The dependency's **main worktree** (`~/Projects/...`) is read-only for this
session. An edit there is a lost edit: no worktree branch, no merge-back cascade,
and it dirties a checkout other sessions share.

## When to use

- About to **create, edit, or delete** files in another project (dependency,
  sibling, upstream, registered wrk project, module outside this tree)
- Task is "fix / implement / change code in dependency X" or similar
- Target is a basename (`dot-pkgs`), another git root, or a `go.mod` replace path
  pointing outside `./external/`

## When not to use

- All writes stay inside the **current** worktree / repo
- Pure **read-only** inspection (bring only before a write)
- Version-only bumps in **this** repo's manifests with no dependency source edit
- Merge-back / finish worktrees (`wrk --done`, etc.) — not this skill's job
- Working in a dedicated dep worktree of your own — use **wrk-pipeline**

## Procedure

1. **Identify** the target: basename, path, or module.
2. **Primary consumer** = the worktree whose `go.mod` (or ship) you are changing.
   `cd` to its root and run every step from there. If cwd is the dependency
   itself (its main checkout, or its own worktree), you are in the wrong tree:
   stop and switch.
3. **Bring — always, unconditionally:**

   ```bash
   wrk --bring dot-pkgs        # a registered basename or an absolute path
   ```

   Reuse is success — `wrk: warning: dot-pkgs already exists under external/;
   reusing …` is the expected output when this session already brought it.
   Repeating a bring never duplicates a live worktree.

   Record `P` = the **absolute path wrk prints**. Never guess it, and never
   hand-roll `external/<name>` (no clone, no `cp`, no symlink, no bare
   `git worktree add`).
4. **Verify `P` before the first write:**

   ```bash
   test -f "$P/.git"                          # .git is a file in a linked worktree, a dir in a main checkout
   git -C "$P" rev-parse --abbrev-ref HEAD    # wrk-created branch, not master/main
   git -C "$P" rev-parse --git-common-dir     # <dep main checkout>/.git
   ```

   `go.mod` must now `replace` to `P`; bring rewrites that directive.
5. **Dep-of-dep** (e.g. consumer needs `agent-pro`, which needs `dot-pkgs`):
   bring both into the **same** consumer `./external/` as siblings. Point any
   nested module's `replace` at that sibling path — never
   `./external/<dep>/external/…`.
6. **Edit and test only `P`.** If `wrk --bring` fails, stop and report; do not
   invent a path, and do not fall back to the dependency's main checkout.

`--dep-replace` / `--pin-locals` point a consumer at an existing tree for build
verification only. They are not an editing entry: dependency source still goes
through `wrk --bring`.

## Anti-patterns

```bash
# Wrong: edit the dependency's main worktree directly
$EDITOR ~/Projects/xhd2015/dot-pkgs/pkg/x.go
# even when go.mod replace points there — run `wrk --bring dot-pkgs` and it
# repoints the replace at ./external/dot-pkgs; edit that path only

# Wrong: hand-roll the tree under external/ (clone / cp / symlink /
# git worktree add). It is not a linked worktree, so the next --bring
# silently creates ./external/dot-pkgs-1 (brainsplit)

# Wrong: keep both the old slot and a new -1 slot

# Wrong: bring from inside a nested brought project
cd external/agent-pro && wrk --bring dot-pkgs
# → creates external/agent-pro/external/dot-pkgs

# Wrong: edit two trees of the same project in one session
```

## Pitfalls

- Repeating `wrk --bring` is safe: it warns and reuses the live linked
  worktree, and refreshes the `replace`. "Skip the bring to avoid a duplicate"
  is what actually creates duplicates.
- A `./external/<name>-1` slot means the occupant was not a live linked
  worktree (plain dir, dead registration, foreign repo). Stop and report;
  inspect `git -C <dep-main> worktree list` before pruning anything.
- `git worktree add: exit status 128 … missing but already registered` is a
  stale registration: stop and report, never force it away.
- Version bump in this repo ≠ source edit in the dependency.
