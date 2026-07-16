# Scenario

**Feature**: wrk --main composes with --reinstall-local (scan main repo modules; no nested shell)

```
# P4 product compose (Classic TDD — currently RED while flags are exclusive)
# linked worktree cwd + --main --reinstall-local [--dry-run]
#   -> PlanLocalReinstallsFromWorkDir(cwd, binDir, useMain=true)
#   -> scan root = main repo path (ResolveMainRepo)
#   -> dry-run plan from main's modules; no nested shell
#
# flag order free:
#   wrk --main --reinstall-local --dry-run
#   wrk --reinstall-local --main --dry-run
#   -> same plan
#
# without --main from linked WT:
#   wrk --reinstall-local --dry-run
#   -> scan root = ShowToplevel(linked WT) → worktree checkout modules
#
# still exclusive:
#   wrk --main --reinstall-local --list
#   -> non-zero, mutually exclusive
```

## Preconditions

- Nested under `reinstall-local-cli/` (inherits root `Request`/`Response`/`Run`,
  GOBIN isolation, session wrk binary). **Do not** seal-break existing
  `dry-run/*`, `dry-run-multi/*`, `execute/*`, `events/*`, or flag/error leaves.
- Git required for linked-worktree fixtures (`from-linked-wt/`).
- Pure API `useMain` / `ResolveReinstallScanRoot` already covered under
  `reinstall-local/scan/`; this branch locks the **CLI flag compose** surface.

## Steps

1. Descendants build linked-worktree fixtures and/or set compose Args.
2. Happy leaves run dry-run only (no real install; GOBIN stubs unchanged).
3. Exclusive leaf combines compose with another mode (`--list`).

## Context

- **Compose contract**: `--main` + `--reinstall-local` is allowed; `--dry-run`
  remains a valid modifier. Scan uses `useMain=true` so planning targets the
  **main repository** of the checkout, not the linked worktree path.
- **No nested shell**: success path prints reinstall plan and exits (same as
  bare `--reinstall-local --dry-run`); must not open interactive shell.
- **Flag order free**: `--main` before or after `--reinstall-local` is equivalent.
- **Without --main**: still plans the **worktree checkout** (ShowToplevel), so a
  diverged linked WT with different modules proves the compose difference.
- **Still exclusive with --list**: stacking `--list` on the compose remains a
  mutual-exclusion error (mirrors `--main --status --list`).

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: --main + --reinstall-local compose (P4 CLI).
	ensureReinstallLocalCLIHelpersUsed()
	return nil
}
```
