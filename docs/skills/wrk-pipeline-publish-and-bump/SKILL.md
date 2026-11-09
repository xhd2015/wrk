---
name: wrk-pipeline-publish-and-bump
description: >-
  Publish a depended-upon repo (merge-back, tag-next, push, sync, reinstall),
  bump only in-context consumers to the new tag with a dep: commit, then merge
  back and clean up. Use when this session's consumer(s) import the repo you are
  shipping. Triggers: publish and bump, dependent repo release, bump consumers,
  dep: module vA -> vB. Not for a repo nothing in context imports (plain
  merge-back is enough).
---

# Publish and bump

Ship a repo that **this session's consumer(s) import**: land it, publish its
next tag, then bump **only those in-context consumers**. Ends with merge-back
and worktree cleanup. Do not scan the wrk project registry for other importers.

## When to use

- The repo you finished is imported by a consumer already in this session
- That consumer needs the new tag in `go.mod` / `go.sum`
- User asks to publish a dependency and bump the consumer(s) in context, or a `dep:` bump commit

## When not to use

- Nothing in this session imports the repo → merge-back only, no tag, no bump
- Only editing dependency source, not shipping → `edit-dependency-code`
- Do not use this skill to bump unrelated registered importers
- Full flag reference → `wrk --help`

## In-context consumers

Bump this set and nothing else:

1. The worktree being published (the dependency).
2. The checkout that `wrk --bring`'d it (parent of `./external/<dep>`), when this session edited that dep for that consumer.
3. Any other repo the user **named** in this conversation.

Never invent consumers from a project inventory. Never open a worktree for an importer that is not in the set above.

## 1. Publish and Bump

Publish the dependency first, then bump the in-context consumers that import it.
If a bumped consumer is itself imported **and** that importer is also in context,
publish the consumer before bumping that importer. Stop there.

### 1a. Publish the dependency

**cwd = that repo's worktree** (the one with the finished work).

```bash
# clean tree
wrk --merge-back --tag-next --push --sync --reinstall-local

# dirty tree: commit first, same land flags after
wrk --add-all --commit -m "<summary>" \
  --merge-back --tag-next --push --sync --reinstall-local
```

| Flag | Role |
|------|------|
| `--add-all --commit -m` | Dirty only; stage + commit with a real message |
| `--merge-back` | Land the branch onto main |
| `--tag-next` | Tag the landed tip with the next version |
| `--push` | Push main **and** the new tag, so consumers can resolve it |
| `--sync` | Fan main tip into other worktrees |
| `--reinstall-local` | Reinstall local bins from the landed tree |

Read the new tag (`vX.Y.Z`) from the output before bumping. Never bump a
consumer to a tag that is not yet pushed — `go get` would fail or resolve the
old version.

### 1b. Bump each in-context consumer

**cwd = that consumer's worktree** (this session's checkout, or `wrk <basename>`
only when the user named a consumer that has no worktree yet).

```bash
go get <module>@<new-tag>
go mod tidy
wrk --add-all --commit -m "dep: <module> vA -> vB"
```

- One commit per module: `dep: <module> vA -> vB` (old tag → new tag).
- `go mod tidy` is not optional: `go.sum` and indirect requirements must move
  with the requirement.
- Repeat only for in-context consumers; if no file changed after `go mod tidy`,
  that consumer was already on the tag — skip its commit.
- A consumer bump commit is regular work: land it in **2. Merge Back**.

## 2. Merge Back

**cwd = the repo whose branch carries the commits** — the dependency worktree
after 1a, or each in-context consumer worktree after 1b.

```bash
# consumer: nothing in context imports it, so no tag and no push. Its dep:
# commit already exists, so the tree is clean — no --commit here.
wrk --merge-back --sync --reinstall-local

# dependency: publish its next tag as part of the land
wrk --merge-back --tag-next --push --sync --reinstall-local
```

`--merge-back` lands onto that repo's main; `--sync` fans the tip into its other
worktrees. The worktree still exists until cleanup. Do not merge-back a session
worktree the user did not ask to land.

## 3. Cleanup

**cwd = the worktree you just landed.** Only after merge-back looks good and the
user agrees.

```bash
wrk --done
```

Removes the worktree and branch; main is the source of truth after merge-back.
Clean up worktrees created for in-context bumps. Do not `--done` the session
checkout.

Report: published tag, per in-context consumer old → new version, commit SHAs,
any in-context consumer that needed no change.

## Pitfalls

- Bumping consumers before the dependency tag exists **and** is pushed
- Bumping a repo that is not in this session
- `--tag-next` / `--push` on a consumer with no dependents
- Forgetting `go mod tidy`, leaving `go.sum` stale
- Bumping a consumer in its main checkout while a worktree holds the real branch
- `--done` before merge-back, dropping commits that only exist on the worktree
