---
name: wrk-pipeline-publish-and-bump
description: >-
  Publish a depended-upon repo (merge-back, tag-next, push, sync, reinstall),
  bump every consumer to the new tag with a dep: commit, then merge back and
  clean up. Use when other projects import the repo you are shipping. Triggers:
  publish and bump, dependent repo release, bump consumers, dep: module vA -> vB.
  Not for a repo nothing imports (plain merge-back is enough).
---

# Publish and bump

Ship a repo that **other projects import**: land it, publish its next tag, then
move every consumer onto that tag. Ends with merge-back and worktree cleanup.

## When to use

- The repo you finished is imported by at least one other registered project
- A new tag must reach every consumer's `go.mod` / `go.sum`
- User asks to publish a dependency and bump consumers, or a `dep:` bump commit

## When not to use

- Nothing imports the repo → merge-back only, no tag, no bump
- Only editing dependency source, not shipping → `edit-dependency-code`
- Full flag reference → `wrk --help`

## 1. Publish and Bump

**Know the edges first.** From any project with a wrk registry:

```bash
wrk --projects-dep-graph
```

Incoming edges into the module you are shipping are its **consumers**. Work
leaves-inward: publish a dependency before bumping the projects that import it,
and if a bumped consumer is itself imported, publish it too before bumping its
own consumers.

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

### 1b. Bump each consumer

**cwd = the consumer's worktree** (create one with `wrk <basename>` when the
consumer has no worktree yet).

```bash
go get <module>@<new-tag>
go mod tidy
wrk --add-all --commit -m "dep: <module> vA -> vB"
```

- One commit per module: `dep: <module> vA -> vB` (old tag → new tag).
- `go mod tidy` is not optional: `go.sum` and indirect requirements must move
  with the requirement.
- Repeat for every incoming edge; if no file changed after `go mod tidy`, the
  consumer was already on that tag — skip its commit.
- A consumer bump commit is regular work: land it in **2. Merge Back**.

## 2. Merge Back

**cwd = the repo whose branch carries the commits** — the dependency worktree
after 1a, or each consumer worktree after 1b.

```bash
# consumer: nothing imports it, so no tag and no push. Its dep: commit already
# exists, so the tree is clean — no --commit here.
wrk --merge-back --sync --reinstall-local

# dependency: publish its next tag as part of the land
wrk --merge-back --tag-next --push --sync --reinstall-local
```

`--merge-back` lands onto that repo's main; `--sync` fans the tip into its other
worktrees. The worktree still exists until cleanup.

## 3. Cleanup

**cwd = the worktree you just landed.** Only after merge-back looks good and the
user agrees.

```bash
wrk --done
```

Removes the worktree and branch; main is the source of truth after merge-back.
Clean up each worktree you used, including consumer worktrees created for bumps.

Report: published tag, per-consumer old → new version, commit SHAs, any consumer
that needed no change.

## Pitfalls

- Bumping consumers before the dependency tag exists **and** is pushed
- Missing a consumer because the graph was not checked (`wrk --projects-dep-graph`)
- `--tag-next` / `--push` on a consumer with no dependents
- Forgetting `go mod tidy`, leaving `go.sum` stale
- Bumping a consumer in its main checkout while a worktree holds the real branch
- `--done` before merge-back, dropping commits that only exist on the worktree
