# Scenario

**Feature**: `wrk --done --sync` runs sync from main after successful done

```
# primary --done succeeds (not aborted) then runSync(main TargetPath)
linked wtA (+ optional wtB behind)
  -> wrk --done [-y|--confirm-from-stdin] --sync
  -> merge-back --rm (message on stdout)
  -> blank line
  -> runSync(main): pass2 distribute / zero-summary
```

## Preconditions

- Git available; monotree root helpers (`setupWrkWorktreeFromMain`, `setupCompositionTwoWTs`, sync stdout builders).
- Composition not implemented yet → leaves **RED** (mutual exclusive or no sync block).

## Steps

- Grouping only: leaves set fixture + `req.Args` / `req.StdinInput`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
