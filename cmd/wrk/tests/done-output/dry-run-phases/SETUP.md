# Scenario

**Feature**: dry-run phase headers only when cascade targets ≥ 1; zero-cascade omits both

```
# zero cascade: no phase headers; primary dry-run plan only
linked wt (no nested cascade)
  -> wrk --done --dry-run
  -> (no ==> cascade, no ==> own)
  -> primary MergeBack dry-run plan
  -> zero mutations

# with cascade: ==> cascade then plan lines then ==> own then primary plan
linked wt + nested cascade
  -> wrk --done --dry-run
  -> ==> cascade
  -> would: cascade merge-back <path>
  -> ==> own
  -> primary dry-run plan
  -> zero mutations
```

## Preconditions

- Parent `done-output/` helpers for phase header asserts (`assertDonePhaseHeaders` /
  `assertNoDonePhaseHeaders`).
- Fixtures reuse done-pipeline dry-run patterns (local ahead wt; optional external cascade).

## Steps

- Grouping: leaves set fixtures + `req.Args = ["--done", "--dry-run"]`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
