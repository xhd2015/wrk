# Scenario

**Feature**: successful wrk --main --status yields main-repo status output

```
# cwd anywhere under a checkout → status of that checkout's main repo
linked or main cwd + --main --status -> runStatus(mainRepo)
  content == (cd mainRepo && wrk --status); Dir may differ by inv cwd
```

## Preconditions

- Fixture is a valid git checkout with a resolvable main repo.
- No nested shell is launched (status path only).

## Steps

- Descendants set cwd (main / external wt / in-tree linked) and Args for the pair.
- Assert exit 0 and Dir-aware content equivalence to `wrk --status` from main.

## Context

- Flag order variants live under `from-external-wt/`.
- In-tree linked must not collapse to the linked-cwd shortcut shape.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Happy leaves own fixture layout; keep main-flag helpers referenced.
	ensureMainFlagHelpersUsed()
	return nil
}
```
