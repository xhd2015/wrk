# Scenario

**Feature**: wrk --reinstall-local appends events.jsonl with command "reinstall-local"

```
# successful reinstall-local (dry-run, execute, or --main compose) -> events.jsonl
mod/ + gobin/ -> wrk [--main] --reinstall-local [--dry-run]
  -> last event: command=reinstall-local, exit_code=0
  -> args include --reinstall-local (and --main / --dry-run when passed)
```

## Preconditions

- Auto-record and event logging run on every wrk invocation (success path).
- Leaves use a minimal go.mod + package main fixture so plan/execute succeed.
- Compose with `--main` requires a git checkout (useMain scan root).
- GOBIN isolation via root harness (`BinDir`); skip-only fixtures avoid real installs.

## Steps

1. Leaves write go.mod + package main (git repo when `--main`); set Args.
2. Run wrk; assert last `events.jsonl` event.

## Context

- Event `command` is the mode name `"reinstall-local"` (from `resolveCommand`),
  including when composed with `--main` (not `"main"`).
- Event `args` record CLI flags (not positionals): always `--reinstall-local`;
  dry-run also records `--dry-run`; compose records `--main`.
- Event `work_dir` is the process cwd (module/repo root under test isolation).
- Help skips events; these leaves are success paths that must append.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: events.jsonl command identity for reinstall-local.
	ensureReinstallLocalCLIHelpersUsed()
	return nil
}
```
