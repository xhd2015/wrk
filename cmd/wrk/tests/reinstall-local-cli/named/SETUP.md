# Scenario

**Feature**: exclusive `--reinstall-local name…` selects bins and skips the binDir gate

```
# named mode (exclusive path only)
mod/ -> wrk --reinstall-local NAME [--dry-run]
  -> plan/install only NAME; ActionInstall even when GOBIN lacks NAME
```

## Preconditions

- Inherits reinstall-local-cli root harness (GOBIN isolation, InProcess).
- Names are illegal with `--done` / pipeline partners (asserted under `error/`).

## Steps

1. Leaves write module fixtures and set `Args` including names.
2. Root `Run` executes wrk as usual.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	_ = req
	return nil
}
```
