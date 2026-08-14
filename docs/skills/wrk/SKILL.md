---
name: wrk
description: >-
  Git worktree helper for isolated feature branches. Use when creating
  linked worktrees, merging back, checking status, linking deps, PRs,
  or looking up registered projects by basename.
---

# When to use

Isolated git worktree without disturbing the main checkout.
Flow: **dashboard** or create → work → `--done` / `--merge-back` → optional PR.

# Lifecycle

| Intent | Command |
|--------|---------|
| Dashboard (no create) | `wrk` |
| Create | `wrk --new` · `wrk <basename> [-t 'task']` |
| Finish | `wrk --done` (merge + remove) · `wrk --merge-back` (merge only) |

`--new` is explicit create; create also runs with create-selecting args (`<dir>`, `-t` / `--task`). Bare `wrk` alone is **dashboard**, not create. `-t` / `--task` slugs branch/dir names.

# Compose pipeline

Fixed stage order (flag order free):

```text
# [pre]  --gen-commit-msg --commit [--model …]   # --commit required with primary
# [main] --done | --merge-back
# [post] --sync → --tag-next → --push → --propagate-tags
# [tail] --reinstall-local   # main tip after merge; empty plan OK
```

```sh
# finish, sync, tag, push main
wrk --done --sync --tag-next --push
# full pre→post→tail
wrk --gen-commit-msg --commit --model=MODEL --done --sync --tag-next --push --reinstall-local
# manual commit + create/attach PR
wrk --add-all --commit -m '…' --pr --title '…' --comment '…'
# plan tag + consumer bumps only
wrk --tag-next --propagate-tags --dry-run
```

- Pre-stage: source worktree; `--dir` invalid when composed.
- `--propagate-tags`: bump consumer `go.mod` requires to source release tags (with `--tag-next` uses new tags; alone uses existing).
- `--json` only for bare `--tag-next` (not with primary / `--propagate-tags`).
- Default auto-yes on `--done` / `--merge-back` / `--set-task`; `--confirm` for Y/n; `-y` still valid.

# Pull request

Multi-mode from a linked worktree (`gh` + github.com origin). **Not every mode needs `--title`/`--comment`.**

| Mode | Flags | Notes |
|------|-------|-------|
| Show | `--pr` | Open PR URL, or empty |
| Status | `--pr --status` | Metadata + checks/reviews; not with title/comment/push |
| Comment-only | `--pr --comment C` (no title) | Existing open PR; error if none |
| Push-existing | `--pr --push` (no title) | Require open PR; full-push tip; optional `--comment` |
| Create/attach | `--pr --title T --comment C` | Both required for this mode; new PR body = comment; existing: title ignored, comment additive |

```sh
# show open PR URL (or empty)
wrk --pr
# PR metadata + checks rollup
wrk --pr --status
# comment on existing open PR
wrk --pr --comment 'Follow-up'
# push tip then create/attach
wrk --push --pr --title 'Fix login' --comment 'Details'
# gen commit → push → PR
wrk --gen-commit-msg --commit --model=MODEL --push --pr --title '…' --comment '…'
```

Compose order: gen-commit → push → pr. With `--push` on create/attach, always full-pushes tip first.

# Inspect & deps

```sh
# status for repos under this checkout
wrk --status
# list worktrees (alias --list)
wrk -l
# recorded main repository paths
wrk --projects
# only github.com origins
wrk --projects --github
# module dep graph across projects
wrk --projects-dep-graph
# lookup project path(s) by basename
wrk --where <basename>
# nested shell at main repo root
wrk --main
# dep worktree(s) under ./external (one or more paths; repeatable)
wrk --bring <path>
wrk --bring p1 p2
# create + bring into the new worktree
wrk src --bring p1 p2
# worktree only; skip replace/tidy
wrk --bring <path> --no-dep
```

# Gotchas

- **Replace guard**: `--done` may refuse local `replace` in `go.mod`; `--no-in-module-replace` for strict block.
- **Basename fallback**: `wrk myrepo` uses registered projects when no local `./myrepo`.

<!-- WRK_SKILL_DOCTEST_MARKER -->
