# Scenario

**Feature**: --dry-run host allowlist includes --reinstall-local after P2

```
# bare --dry-run (no host mode) -> non-zero; stderr host list names --reinstall-local
wrk --dry-run -> error
```

## Steps

- Descendants run bare `--dry-run` without `--reinstall-local` (or other hosts).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: bare dry-run host-list errors.
	ensureReinstallLocalCLIHelpersUsed()
	return nil
}
```
