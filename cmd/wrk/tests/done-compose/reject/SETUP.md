# Scenario

**Feature**: illegal flag combinations with primary are rejected

```
wrk --done|--merge-back --json -> non-zero; --json not valid with primary
```

## Preconditions

- `--json` is not a valid post-modifier of `--done` / `--merge-back`.
- Bare `wrk --push` is a **standalone mode** (see `cmd/wrk/tests/push/`); not rejected here.

## Steps

- Leaves set the illegal combo and assert non-zero + stderr policy.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: illegal combos still run from a git main repo when useful.
	skipIfNoGit(t)
	return nil
}
```
