# Scenario

**Feature**: wrk --reinstall-local --dry-run plan output

```
# module + GOBIN stubs -> would:/skip: lines + summary; exit 0; no installs
mod/ + gobin/ -> wrk --reinstall-local --dry-run
```

## Preconditions

- Leaves under this branch write a valid `go.mod` and discovery fixtures.
- Args always include `--reinstall-local` and `--dry-run`.
- Expect exit 0 and exact stdout vocabulary.

## Steps

1. Leaves write go.mod, package mains, optional GOBIN stubs.
2. Set `Args` to `--reinstall-local --dry-run`.
3. Assert stdout plan + summary; stub bins unchanged.

## Context

- Group default: successful dry-run (exit 0).

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: dry-run happy path defaults.
	req.Args = []string{"--reinstall-local", "--dry-run"}
	return nil
}
```
