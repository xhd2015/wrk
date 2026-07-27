# Scenario

**Feature**: TTY keys from default MERGE BACK focus run batch dry-run (DONE path after select)

```
linked-wt (ahead, dirty) + DRY_RUN + argv log
  -> tty-watch: default focus MERGE BACK
  -> j to DONE, space select DONE, j×5 to RUN ALL, Enter
  -> compose argv log = DONE recipe + --dry-run + --add-all
  -> terminal exits; worktree still present
```

## Steps

1. Linked worktree ahead + untracked dirty (add changes on → `--add-all`).
2. tty-watch bare `wrk` with `WRK_DASHBOARD_DRY_RUN=1` and argv log.
3. Focus order: add, gen, commit, **MERGE BACK** (default), DONE, sync, tag, push, reinstall, **RUN ALL**, CANCEL.
4. Keys: `j` (DONE), `space` (select DONE), five `j` (to RUN ALL), Enter.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	linked := setupDashboardLinkedAhead(t, req)
	markDashboardDirtyUntracked(t, linked)
	req.Args = nil
	return nil
}
```
