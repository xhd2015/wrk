# Scenario

**Feature**: primary (`--done` / `--merge-back`) composes with pre-stage gen-commit and post modifiers at the flag layer

```
# primary + allowed modifiers pass flag validation (no mutex / only-valid-with for those pairs)
wrk --done|--merge-back [--sync] [--tag-next] [--push] [--propagate-tags] [--reinstall-local] [--dry-run]
  -> flag layer accepts composition
  -> later stage may still error (e.g. not a linked worktree on main)
  -> stage order after primary success: sync → tag-next → push → propagate-tags → reinstall-local → exec → land

# P2 pre-stage: --gen-commit-msg --commit on source worktree before primary
wrk --gen-commit-msg --commit [--model M] --done|--merge-back [post…]
  -> flag layer accepts (not mutually exclusive)
  -> missing --commit with primary → clear error; no primary
  -> composed --dir with primary → reject (workDir wins)
  -> event command stays "done" / "merge-back" (not gen-commit-msg)

# illegal: --json with primary; non-composed exclusives
wrk --done --json / wrk --tag-next --list / wrk --reinstall-local --sync / wrk --reinstall-local --list
wrk --gen-commit-msg --sync   # no primary → still exclusive
  -> non-zero, clear stderr
# bare wrk --push is standalone (cmd/wrk/tests/push/), not rejected here

# user-facing help: composition documented in usage()
wrk --help
  -> --done/--merge-back list optional pre --gen-commit-msg and post modifiers;
     --push dual meaning
```

## Preconditions

- Reuses root `cmd/wrk/tests` harness (`Request` / `Response` / `Run`).
- Git available for flag-layer leaves; **`help/`** leaf is help-only (no git).
- Flag-layer leaves use a **main** repo checkout so mutex checks fire **before** heavy merge-back / tagscope work.
- Flag validation leaves: not full merge+tag+push+propagate+reinstall e2e; `help/` asserts `usage()` substrings only.
- **P7**: `--propagate-tags` is an allowed post modifier with primary.
- **P1 reinstall tail**: `--reinstall-local` is an allowed post-success tail after primary only.
  Bare `--reinstall-local --sync` / `--list` stay exclusive.
- **P2 gen-commit pre**: `--gen-commit-msg --commit` may compose with primary (± post). Classic RED
  until unlocked (today early `runGenCommitMsg` rejects `--done`/`--merge-back` as mutex).
  Bare `--gen-commit-msg` (no primary) and library path stay covered under `gen-commit-msg/`.
  Pipeline dry-run pre+primary: `done-pipeline/dry-run/with-gen-commit-msg/`.
- **P3 full integration**: flag-layer full ship under
  `allow/done/with-gen-commit-msg-sync-tag-next-push-reinstall-local/`;
  ordered dry-run under `done-pipeline/dry-run/full-combo-gen-commit-reinstall/`;
  help documents all three bands (pre + primary + posts/tail).

## Steps

- Grouping only. Leaves set `req.Args` (flag leaves: minimal main repo via `initGitRepoOnMain`; `help/`: `--help` only).

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```
