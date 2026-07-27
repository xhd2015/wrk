# Scenario

**Feature**: cascade MergeBack failure surfaces as structured `Error:` with path context (not bare git framing alone)

```
# diverged external dep wt → cascade merge-back fails (cascade ≥ 1)
consumer wt + diverged external/dep
  -> wrk --done
  -> ==> cascade  (required: non-empty cascade before fail)
  -> non-zero exit mid-cascade
  -> stderr includes Error: + external path / path field
  -> ==> own may be absent (own phase not reached)
  -> must not lead with bare "rebase conflict:" as sole framing
```

## Preconditions

- Parent `setupDivergedExternalForCascadeFail` builds conflicting main/external `dep.go`.
- Real apply (not dry-run) so `MergeBack` executes rebase and fails.

## Steps

- Grouping: leaf sets Args for `--done` (default auto-yes).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
