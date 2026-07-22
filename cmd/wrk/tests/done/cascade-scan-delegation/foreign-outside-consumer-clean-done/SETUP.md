# Scenario

**Feature**: dirty git tree **outside** consumer must not block `wrk --done`
(cascade Scan roots consumer only; base-path filter + no foreign cascade target)

```
# layout
WorkRoot/
  myrepo/                         # consumer main
  .wrk/worktrees/myrepo-…/        # clean consumer linked wt (cwd)
  other/external/agent-pro/       # dirty independent main — NOT under consumer

# under test
clean consumer linked wt (zero nested cascade under consumer)
  + dirty foreign sibling on disk
  -> wrk --done
  -> exit 0; own merge-back / remove succeeds
  -> stderr/stdout never name foreign path; no dirty Error: for outside tree
  -> foreign dirty tree still present on disk (untouched)
```

## Preconditions

- Coverage backfill (P1 filter): **GREEN OK** — Scan(consumer) must not return
  foreign; dirty preflight must not see outside path. Regression guard if
  warm-cache / walk-log leak returns.
- No nested linked worktrees under consumer (zero cascade targets).

## Steps

1. Build clean consumer linked wt + dirty foreign sibling via group helper.
2. Run bare `wrk --done`.

```go
func Setup(t *testing.T, req *Request) error {
	setupCascadeScanForeignIsolation(t, req)
	req.Args = []string{"--done"}
	return nil
}
```
