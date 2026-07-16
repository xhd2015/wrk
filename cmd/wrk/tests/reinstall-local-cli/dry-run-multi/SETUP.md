# Scenario

**Feature**: wrk --reinstall-local --dry-run multi-module plan output (P3)

```
# multi-module tree (git toplevel or walk-up scan root) + GOBIN stubs
#   -> PlanLocalReinstallsFromWorkDir(useMain=false)
#   -> grouped # module headers + would:/skip: lines
#   -> would: reinstall N binaries (M skipped) across K modules  (K>1 only)
cwd + GOBIN -> wrk --reinstall-local --dry-run
```

## Preconditions

- Leaves build multi-module fixtures under WorkRoot (nested go.mod dirs and/or
  git checkout). Process cwd is `req.ModuleRoot` (may be repo root or a subdir).
- Args always include `--reinstall-local` and `--dry-run`.
- **K=1** single-module dry-run format is owned by sealed `dry-run/` leaves
  (no `# module` headers; summary without `across K modules`). These multi
  leaves only lock **K>1** (and multi-module collision errors).
- Cross-module install×install collision is a hard plan error (stderr, non-zero).

## Steps

1. Leaves write multi-module trees, optional git init, GOBIN stubs; set cwd via
   `req.ModuleRoot`.
2. Group sets default Args to dry-run reinstall-local.
3. Assert multi dry-run stdout vocabulary or collision stderr.

## Context

- Group default: multi dry-run (exit 0 when plan succeeds).
- Module headers: `# module <ModulePath> (<RelDir>)` where RelDir is relative
  to the scan root (`.` for the scan-root module itself).
- Module blocks ordered by absolute ModuleRoot lex; items within a module by
  BinName (same as pure multi plan).
- Summary last line when K>1:
  `would: reinstall N binaries (M skipped) across K modules\n`
  N/M are totals across all modules; K = number of modules in the multi plan.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: multi-module dry-run defaults.
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
