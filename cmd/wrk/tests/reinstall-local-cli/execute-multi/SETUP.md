# Scenario

**Feature**: wrk --reinstall-local executes multi-module reinstall plans (P5)

```
# multi-module tree + GOBIN stubs
#   -> PlanLocalReinstallsFromWorkDir(useMain=false)
#   -> executeMultiLocalReinstalls: Dir=each ModuleRoot
#   -> progress lines + reinstalled N, skipped M, failed F
# continue after failures; install×install collision fails before any go install
cwd + GOBIN -> wrk --reinstall-local
```

## Preconditions

- Leaves build multi-module fixtures under WorkRoot (nested go.mod dirs).
  Process cwd is `req.ModuleRoot` (scan root for walk-up / non-git multi).
- Args are `--reinstall-local` **without** `--dry-run`.
- Go toolchain available for real `go install` and compile failures.
- GOBIN is the isolated leaf `BinDir` (set by root `Run`).
- Sealed single-mod `execute/*` and multi dry-run `dry-run-multi/*` leaves are
  not modified by this group.

## Steps

1. Leaves write multi-module trees, package mains (buildable or deliberately
   broken), GOBIN stubs; set cwd via `req.ModuleRoot` when needed.
2. Group sets default Args to bare reinstall-local (execute).
3. Assert exit code, execute summary / progress, GOBIN side effects, or
   plan-time collision stderr without mutation.

## Context

- Group default: multi-module **execute** path (P5). Production already wires
  `executeMultiLocalReinstalls` after the multi plan (coverage backfill; GREEN
  expected if P3 execute multi landed).
- Progress lines are the same as single-mod execute (no `# module` headers;
  no `would:` prefix): `go install <RelPath>` / `go run <RelPath>` with
  `Dir=<that module's ModuleRoot>`.
- Module order: multi-plan order (lex absolute ModuleRoot); items within a
  module by BinName.
- Summary last line (totals across all modules):
  `reinstalled N, skipped M, failed F\n`
  (no `across K modules` suffix — that is dry-run-only).
- Exit **1** iff `failed > 0`; plan-time collision → non-zero before any install.
- Cross-module install×install collision is a hard plan error (stderr, non-zero);
  stubs must remain unchanged.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: multi-module execute defaults (no --dry-run).
	req.Args = []string{"--reinstall-local"}
	return nil
}
```
