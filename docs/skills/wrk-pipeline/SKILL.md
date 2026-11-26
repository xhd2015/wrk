---
name: wrk-pipeline
description: >-
  Three-phase wrk lifecycle (PR optional): prepare (bring), PR (commit + open
  PR), publish (merge-back + tag/push/sync/reinstall). Removing the worktree
  (`wrk --done`) is a separate fourth step the user must ask for — never part
  of the default plan. Skip PR and jump to publish when no review PR is needed.
  Triggers: wrk pipeline, ship worktree, wrk --pr, merge-back reinstall,
  /wrk-pipeline. Not for in-tree-only edits without a wrk worktree, or
  multi-agent fan-out (use orchestrate-agent-runners).
---

# wrk pipeline

Canonical **bring → (optional PR) → land** lifecycle after work on a wrk
worktree is verified. Removing the worktree is **not** the finish line of this
lifecycle: `--done` is a separate, explicitly user-requested action (see §4).

## When to use

- Ship or finish a **brought / feature worktree** (commit, PR, merge-back, tag, reinstall)
- User asks for the wrk ship line, `wrk --pr`, or merge-back + tag-next
- User **explicitly** asks to remove/delete/clean up a worktree (`wrk --done`) — see §4
- Out-of-tree dependency work that already used (or will use) `wrk --bring`

## When not to use

- All edits stay in the **current** repo/worktree with no wrk bring/ship
- Multi-agent fan-out / keep-alive land → `orchestrate-agent-runners`
- Full flag semantics → `wrk --help`

## Pipeline overview

| Phase | When | Command (cwd noted per phase) |
|-------|------|-------------------------------|
| **1 Prepare** | Before out-of-tree writes | `wrk --bring <name-or-path>` |
| **2 PR** | Verified; user wants remote/review PR | see **PR Phase** |
| **3 Publish** | User wants land on main (after PR **or** skip PR) | see **Publish Phase** |
| **4 Cleanup** (never automatic) | Only when a user message asks to remove that worktree; cwd = main checkout | `wrk --done <worktree-dir>` |

**Skip PR:** if the user does not need a PR → jump **1 → work → 3**. Stop there:
phase 4 is not part of the default path and must not be appended to a plan.

**Between Prepare and PR/Publish:** implement and verify on the brought tree. This skill does not own that work.

**Commit:** PR always uses `--add-all --commit -m` (real message; spelling **`--commit`**). Publish commit only when dirty — see **§3**.

## Gates

| Gate | Rule |
|------|------|
| Prepare → work | Brought path exists and is used for all edits/tests |
| Work → **PR** | Checks green; user wants PR |
| Work → **Publish** | Checks green; user **skips PR** (jump to 3) |
| PR → Publish | Not automatic — run land only when user asks |
| Publish → Cleanup | Ship exit 0 **and** a user message explicitly asks to remove/delete that worktree. "Merge", "ship", "land", "install", "finish" are **not** removal orders |
| Cleanup | `--done` deletes the worktree dir immediately and interrupts any agent whose cwd is inside it — run it from the main checkout with the path, and only when the user explicitly allows. Never add it to a todo list or chain it in the same turn as Publish |
| Never | `wrk --done` before land if that would drop unmerged work still only on the worktree |

## 1. Prepare Phase

From the **consumer** worktree root:

```bash
wrk --bring <name-or-path>
```

- Note the path wrk prints (usually `./external/…`).
- **All** later read/write/test/ship for that project use that path — not the bare main checkout unless the user overrides.
- Edits stay inside that brought path; do not touch the dependency main checkout.
- If bring **fails**, stop and report; do not invent a path.

## 2. PR Phase

**cwd = brought worktree.** Optional — skip when the user does not need a PR (go to **Publish**).

```bash
wrk --add-all --commit -m "<summary>" \
  --pr --title "<PR title>" --comment "<PR body>"
```

| Flag | Role |
|------|------|
| `--add-all --commit -m` | Stage + commit; real message (see overview **Commit**) |
| `--pr` | Open PR (wrk semantics) |
| `--title` / `--comment` | PR title + body; fill real text |

- Do **not** merge-back here — land is **Publish**.
- Do **not** auto-run Publish or Cleanup after PR.
- Report: commit SHA, PR URL if printed.

## 3. Publish Phase

**cwd = brought worktree.** Run when the user wants land on main (after PR, or when PR was skipped).

Pick the row that matches what depends on the repo you are shipping:

| Repo kind | Command |
|-----------|---------|
| **Consumer** (no project imports it) — clean | `wrk --merge-back --sync --reinstall-local` |
| **Consumer** — dirty | `wrk --add-all --commit -m "<summary>" --merge-back --sync --reinstall-local` |
| **Depended upon** (an in-context consumer imports it) | load **`wrk-pipeline-publish-and-bump`**: tag this repo, then bump only those in-context consumers |

| Flag | Role | Needed by |
|------|------|-----------|
| `--add-all --commit -m` | Dirty only; real message (see overview **Commit**) | both |
| `--merge-back` | Land branch onto main; worktree may remain until Cleanup | both |
| `--tag-next` | Tag the landed tip with the next version | depended upon |
| `--push` | Push main and the new tag; also lease-updates `origin/<worktree-branch>` when that ref already exists | depended upon |
| `--sync` | Fan main tip into other worktrees | both |
| `--reinstall-local` | Reinstall local bins from the landed tree | both |

- Optional: `--dry-run` first if the user is unsure, then re-run without it.
- Report: commit SHA, merge-back result, new tag(s) if any, reinstall summary.
- A diverged branch lands by rebasing onto main. When that rebase conflicts, the
  landing repository's `script/git/hooks/rebase/on-conflict` resolver runs (shell
  or Go, committed in the repository) and the rebase continues if it clears every
  conflict; a conflict it leaves stops the land with `rebase conflict:` as before.
  The result line then says `(N conflicted path(s) resolved by the repository's
  on-conflict hook)`. Report that sentence rather than "merged cleanly".

## 4. Cleanup Phase (explicit user request only)

**cwd = main checkout.** Never part of an agent's plan or todo list, and never
chained in the same turn as Publish/reinstall. Run it only when the user's own
message asks to remove that worktree (done / cleanup / delete / 收尾). After
Publish, **stop and report**: leaving the worktree in place is the correct end
state, not unfinished work.

```bash
wrk --done <worktree-dir>     # from main: merge + remove that worktree
```

`--done` deletes the worktree directory immediately (auto-yes unless `--confirm`), so running it with the shell — or a running agent, this one included — inside that directory interrupts the caller: interactive bash only gets a follow-up `cd`, an agent does not. Never run it from inside the worktree, and only when the user explicitly allows. Inside the worktree use `wrk --merge-back` to land without removing.

Removes the worktree (and branch per wrk). Source of truth is **main** after merge-back.

## Pitfalls

- Editing the dependency **main** checkout instead of the brought `external/…` path
- PR / Publish from the **wrong cwd** (not the worktree); `--done` from **inside** the worktree it removes
- Shipping or opening PR before verify (`--add-all` will commit whatever is dirty)
- Empty `--title` / `--comment` on PR
- Auto-chaining Publish → Cleanup without the user asking
- Treating Cleanup as the pipeline's default last step, adding it to a todo list, or running `--done` because the code work is "finished"
- Reading "merge / ship / land / install" as permission to remove the worktree
- Tagging or pushing a **consumer** that nothing imports (it needs merge-back only)
- `--done` after PR-only while land was still intended, or before merge-back with unmerged work only on the worktree
- Editing dependency source anywhere other than the brought `external/…` path
