# Scenario

**Feature**: wrk --dep-update --all inventory pull over the unwind stack

```
# stack consumer set + registered owner modules in WRK_HOME
cwd consumer (or linked worktree)
  -> wrk --dep-update --all [--dry-run]
  -> CollectStackInventory + BuildInventory ownership + latest tags
  -> pin outdated inventory-owned requires on every stack checkout; same tidy helper
  -> skip external / same-checkout filesystem replace; warn no-tag
```

## Preconditions

- Leaves under this node use `req.Args` containing `--dep-update` and `--all`.
- Consumer is a **git** checkout (ShowToplevel); need not be in projects.json.
- Owner project(s) registered via `writeProjectsJSON`.
- Apply leaves seed `file://` modproxy for offline tidy when pins run.

## Steps

1. Grouping marks inventory-pull scenarios.
2. Subtrees split dry-run / apply / soft outcomes.

## Context

- Default partner flags only; leaves append `--dry-run` when needed.
- Blast radius: mutate go.mods under the stack member Paths (linked worktree Path, not MainRepo).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	ensureDepUpdateHelpersUsed()
	return nil
}
```
