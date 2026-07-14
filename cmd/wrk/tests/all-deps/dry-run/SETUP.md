# Scenario

**Feature**: wrk --dry-run flag validation (without --all-deps)

```
# bare --dry-run without --all-deps -> error before any planning
wrk --dry-run -> error (--dry-run is only valid with --all-deps)
```

## Preconditions

- `--dry-run` is valid ONLY with `--all-deps`.
- Registered-project dry-run planning leaves live under `registered/dry-run/`.

## Steps

- Descendants invoke `wrk --dry-run` without `--all-deps` and assert stderr.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()
	return nil
}
```