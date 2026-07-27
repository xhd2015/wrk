# Scenario

**Feature**: wrk --reinstall-local (no --dry-run) executes planned installs

```
# module + GOBIN stubs -> go install|go run for Action=install; skip otherwise
# continue on failure; summary reinstalled N, skipped M, failed F
# exit 1 iff failed > 0
mod/ + gobin/ -> wrk --reinstall-local
```

## Preconditions

- Leaves under this branch write a valid `go.mod` and discovery fixtures.
- Args are `--reinstall-local` **without** `--dry-run`.
- Go toolchain available for real `go install` / compile failures.
- GOBIN is the isolated leaf `BinDir` (set by root `Run`).

## Steps

1. Leaves write go.mod, package mains (buildable or deliberately broken), optional GOBIN stubs.
2. Set `Args` to `--reinstall-local` only.
3. Assert exit code, execute summary, and GOBIN side effects (or lack of go runs).

## Context

- Group default: real execute path (P3). Currently production returns
  `not implemented yet` → leaves are RED until implementer wires apply.
- Progress lines: `go install <RelPath>` / `go run <RelPath>` (no `would:`).
- Skip lines unchanged: `skip: <bin> (not in <bindir>)`.
- Summary always: `reinstalled N, skipped M, failed F\n`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: execute path defaults (no --dry-run).
	req.Args = []string{"--reinstall-local"}
	return nil
}
```
