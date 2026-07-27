# Scenario

**Feature**: dirty git tree **outside** consumer must not block
`wrk --done --dry-run` (same Scan isolation as apply path)

```
# layout (same as foreign-outside-consumer-clean-done)
WorkRoot/
  myrepo/
  .wrk/worktrees/myrepo-…/        # clean consumer linked wt
  other/external/agent-pro/       # dirty foreign main outside consumer

# under test
clean consumer linked wt + dirty foreign sibling
  -> wrk --done --dry-run
  -> exit 0; plan only (zero mutations)
  -> no foreign path in plan/Error:; no dirty preflight fail for outside tree
  -> consumer wt still linked; foreign still dirty
```

## Preconditions

- Coverage backfill (P1): **GREEN OK**.
- Zero nested cascade under consumer → no `would: cascade` for foreign.

## Steps

1. Build clean consumer + dirty foreign sibling via group helper.
2. Run `wrk --done --dry-run`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadeScanForeignIsolation(t, req)
	req.Args = []string{"--done", "--dry-run"}
	return nil
}
```
