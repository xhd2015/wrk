---
name: wrk
description: wrapping git worktree, used when user wants to commit, merge back into main repo, creating PRs
---


# Create PR

To commit and create PR:
```sh
wrk --add-all --commit -m '...' --pr --title '...' --comment '...'
```