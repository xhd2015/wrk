# Scenario

**Feature**: primary (`--done` / `--merge-back`) composes with post modifiers at the flag layer

```
# primary + allowed modifiers pass flag validation (no mutex / only-valid-with for those pairs)
wrk --done|--merge-back [--sync] [--tag-next] [--push] [--propagate-tags] [--dry-run]
  -> flag layer accepts composition
  -> later stage may still error (e.g. not a linked worktree on main)

# illegal: --json with primary; non-composed exclusives
wrk --done --json / wrk --tag-next --list
  -> non-zero, clear stderr
# bare wrk --push is standalone (cmd/wrk/tests/push/), not rejected here

# user-facing help: composition documented in usage()
wrk --help
  -> --done/--merge-back list optional post modifiers; --push dual meaning
```

## Preconditions

- Reuses root `cmd/wrk/tests` harness (`Request` / `Response` / `Run`).
- Git available for flag-layer leaves; **`help/`** leaf is help-only (no git).
- Flag-layer leaves use a **main** repo checkout so mutex checks fire **before** heavy merge-back / tagscope work.
- Flag validation leaves: not full merge+tag+push+propagate e2e; `help/` asserts `usage()` substrings only.
- **P7**: `--propagate-tags` is an allowed post modifier with primary (Classic RED until unlocked).

## Steps

- Grouping only. Leaves set `req.Args` (flag leaves: minimal main repo via `initGitRepoOnMain`; `help/`: `--help` only).

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```
