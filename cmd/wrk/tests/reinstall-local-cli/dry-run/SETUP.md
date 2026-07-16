# Scenario

**Feature**: wrk --reinstall-local --dry-run plan output (+ optional diagnostics)

```
# module + GOBIN stubs -> would:/skip: lines + summary; exit 0; no installs
# conflict/ambiguity -> diagnostics on stderr (plain under pipe; --color prefixes)
mod/ + gobin/ -> wrk --reinstall-local --dry-run [--color]
```

## Preconditions

- Leaves under this branch write a valid `go.mod` and discovery fixtures.
- Args include `--reinstall-local` and `--dry-run` (leaves may add `--color`).
- Expect exit 0 and exact stdout vocabulary; stderr locked when diagnostics apply.

## Steps

1. Leaves write go.mod, package mains, optional GOBIN stubs.
2. Set `Args` to `--reinstall-local --dry-run` (or with `--color`).
3. Assert stdout plan + summary; optional stderr diagnostics; stub bins unchanged.

## Context

- Group default: successful dry-run (exit 0). Diagnostics are non-fatal.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: dry-run happy path defaults.
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
