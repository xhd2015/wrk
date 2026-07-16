# Scenario

**Feature**: primary + allowed post modifiers are accepted at flag validation

```
# --done/--merge-back with --tag-next / --push / --sync / --propagate-tags / --dry-run
# must not fail as "mutually exclusive" or "only valid with --tag-next"
main repo -> wrk <primary> <modifiers> -> past flag layer
```

## Preconditions

- Allowed modifiers with primary: `--sync`, `--tag-next`, `--push` (branch under primary; does not require `--tag-next`), `--propagate-tags` (P7 post stage; may pair with or without `--tag-next`), `--dry-run` (composition host). Full multi-stage apply/plan is covered under `done-pipeline/` / `merge-back-pipeline/` (not this flag-matrix tree).
- From a main-repo cwd, post-flag errors like `not a linked worktree` prove the flag check passed (flag-layer leaves only).

## Steps

- Descendants set primary + modifier `req.Args` on a main repo.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: ensure git for main-repo flag fixtures in descendants.
	skipIfNoGit(t)
	return nil
}
```
