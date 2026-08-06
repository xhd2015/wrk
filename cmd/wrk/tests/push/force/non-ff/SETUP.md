# Scenario

**Feature**: diverged origin tip — bare `--push` fails; `--push -f` force-with-lease succeeds

```
# shared fixture: local main and origin/main diverged
setupPushDivergedMainWithOrigin
  -> wrk --push           # control: non-zero; origin unchanged
  -> wrk --push -f        # force: exit 0; origin == local HEAD; pushed … confirm
```

## Preconditions

- Uses `setupPushDivergedMainWithOrigin` from parent `force/SETUP.md`.
- Origin tip snapshotted at `{WorkRoot}/origin-main-before` before Run.

## Steps

- Grouping: leaves call diverged fixture and set Args with or without force.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	req.InProcess = true
	return nil
}
```
