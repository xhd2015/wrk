# Scenario

**Feature**: real TTY bare `wrk` runs Bubble Tea dashboard (stays alive) via tty-watch

```
linked-wt -> tty-watch run --detach -- wrk   # empty args = dashboard
  -> snapshot shows dashboard while process alive
  -> send q / CANCEL -> process exits; no compose mutations
non-TTY harness path still one-shot snapshot (regression)
```

## Preconditions

- `tty-watch` on PATH or common install paths (`lookPathTTYWatch`).
- Set `TERM=dumb` for the child (helpers do this) so Bubble Tea OSC color query does not hang.
- Isolated `TTY_WATCH_HOME` under `WorkRoot`.

## Steps

- Leaves use `runDashboardTTYWatch` from parent `dashboard/SETUP.md`.
- Do **not** set `WRK_DASHBOARD_ACTION` on TTY leaves (tea path, not hermetic ACTION).

## Keys / env

| Input | Effect |
|-------|--------|
| `q` / `Esc` / CANCEL+Enter | cancel; exit 0; event dashboard; no compose |
| RUN ALL+Enter | quit tea frame; compose; **re-open TUI** (stay until CANCEL/`q`) |
| `WRK_DASHBOARD_DRY_RUN=1` | inject `--dry-run` on TTY RUN |
| `WRK_DASHBOARD_COMPOSE_ARGV_LOG` | argv log on TTY RUN |

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping only; leaves call parent helpers (runDashboardTTYWatch) by name.
	// Do not bare-ref parent helpers here — dual-mode child packages cannot
	// resolve free names like lookPathTTYWatch as values.
	return nil
}
```
