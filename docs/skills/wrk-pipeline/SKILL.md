---
name: wrk-pipeline
description: >-
  Four-phase wrk lifecycle (PR optional): prepare (bring), PR (commit + open
  PR), publish (merge-back + tag/push/sync/reinstall), cleanup (done). Skip PR
  and jump to publish when no review PR is needed. Triggers: wrk pipeline, ship
  worktree, wrk --pr, merge-back reinstall, wrk --bring then --done,
  /wrk-pipeline. Not for in-tree-only edits without a wrk worktree, or
  multi-agent fan-out (use orchestrate-agent-runners).
---

# wrk pipeline

Canonical **bring → (optional PR) → land → remove worktree** lifecycle after
work on a wrk worktree is verified.

## When to use

- Ship or finish a **brought / feature worktree** (commit, PR, merge-back, tag, reinstall, done)
- User asks for the wrk ship line, `wrk --pr`, merge-back + tag-next, or `wrk --done`
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
| **4 Cleanup** | After land confirmed | `wrk --done` |

**Skip PR:** if the user does not need a PR → jump **1 → work → 3 → 4**.

**Between Prepare and PR/Publish:** implement and verify on the brought tree. This skill does not own that work.

**Commit:** PR always uses `--add-all --commit -m` (real message; spelling **`--commit`**). Publish commit only when dirty — see **§3**.

## Gates

| Gate | Rule |
|------|------|
| Prepare → work | Brought path exists and is used for all edits/tests |
| Work → **PR** | Checks green; user wants PR |
| Work → **Publish** | Checks green; user **skips PR** (jump to 3) |
| PR → Publish | Not automatic — run land only when user asks |
| Publish → Cleanup | Ship exit 0; user confirms (or already ordered land **and** done) |
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

## 4. Cleanup Phase

**cwd = brought worktree.** Only after **Publish** looks good (and user OK, unless they already ordered full land+done). Not the default after PR-only.

```bash
wrk --done
```

Removes the worktree (and branch per wrk). Source of truth is **main** after merge-back.

## Pitfalls

- Editing the dependency **main** checkout instead of the brought `external/…` path
- PR / Publish / `--done` from the **wrong cwd** (not the worktree)
- Shipping or opening PR before verify (`--add-all` will commit whatever is dirty)
- Empty `--title` / `--comment` on PR
- Auto-chaining PR → Publish → Cleanup without the user asking
- Tagging or pushing a **consumer** that nothing imports (it needs merge-back only)
- `--done` after PR-only while land was still intended, or before merge-back with unmerged work only on the worktree
- Editing dependency source anywhere other than the brought `external/…` path
