---
name: edit-dependency-code
description: >-
  Before writing code in another project, reuse ./external/ if already brought,
  otherwise wrk --bring into the primary consumer worktree's external/ (never
  nested), then edit only that SSOT path. Use when modifying dependency code,
  sibling/upstream repos, registered wrk projects (e.g. dot-pkgs), go.mod
  replace targets, or any tree outside the current checkout. Triggers: edit
  dependency code, other project, upstream package, out-of-tree, registered
  project, sibling module, wrk --bring, bring into scope. Slash:
  /edit-dependency-code. Do not use for in-repo-only edits, or version-only
  bumps with no source edit in the dependency.
---

# Edit dependency code

Change source in another project only after it is in **this** session's scope
under the primary consumer's `./external/`, as **one** SSOT path.

## When to use

- About to **create, edit, or delete** files in another project (dependency,
  sibling, upstream, registered wrk project, module outside this tree)
- Task is "fix / implement / change code in dependency X" or similar
- Target is a basename (`dot-pkgs`), another git root, or a `go.mod` replace path

## When not to use

- All writes stay inside the **current** worktree / repo
- Pure **read-only** inspection (bring only before a write)
- Version-only bumps in **this** repo's manifests with no dependency source edit
- Merge-back / finish worktrees (`wrk --done`, etc.) — not this skill's job

## Procedure

1. **Identify** the target: basename, path, or module.
2. **Primary consumer** = the worktree whose `go.mod` (or ship) you are changing.
   Run all bring/reuse steps from that root.
3. **Reuse check** — before any `--bring`:

   ```bash
   ls external/
   rg -n 'replace .*external/|external/<basename>' go.mod
   ```

   - If `go.mod` already `replace`s to `./external/…` → that path is SSOT.
   - Else if `./external/<basename>` (or an existing `./external/<basename>-N`)
     is present → that path is SSOT.
   - **Do not** run `wrk --bring` when a usable tree already exists.
4. **Bring only if missing**, from the primary consumer root:

   ```bash
   wrk --bring dot-pkgs
   ```

   Use `--bring` (not `--dep`). Note the path wrk prints under `./external/…`.
5. **Dep-of-dep** (e.g. consumer needs `agent-pro`, which needs `dot-pkgs`):
   bring both into the **same** consumer `./external/` as siblings. Point any
   nested module's `replace` at that sibling path — never
   `./external/<dep>/external/…`.
6. **Edit and test only the SSOT path.** If `wrk --bring` fails, stop and
   report; do not invent a path or fall back to an unregistered main checkout.

## Anti-patterns

```bash
# Wrong: bring again when ./external/dot-pkgs already exists
# → creates ./external/dot-pkgs-1 (brainsplit)
wrk --bring dot-pkgs

# Wrong: bring from inside a nested brought project
cd external/agent-pro && wrk --bring dot-pkgs
# → creates external/agent-pro/external/dot-pkgs

# Wrong: edit the dependency's main checkout while replace → ./external/…

# Wrong: edit two trees in one session (e.g. workspace/external/ and
# consumer/external/)
```

## Pitfalls

- Brainsplit = two writable trees of the same project (see **Anti-patterns**).
- "Already brought this session" is not enough — trust on-disk `./external/`
  and `go.mod` `replace` first (see **Procedure**).
- Version bump in this repo ≠ source edit in the dependency.
