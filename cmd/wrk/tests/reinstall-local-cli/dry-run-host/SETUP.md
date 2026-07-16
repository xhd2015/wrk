# Scenario

**Feature**: --dry-run host allowlist includes --reinstall-local after P2

```
# bare --dry-run (no host mode) -> non-zero; stderr host list names --reinstall-local
wrk --dry-run -> error
```

## Steps

- Descendants run bare `--dry-run` without `--reinstall-local` (or other hosts).

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: bare dry-run host-list errors.
	ensureReinstallLocalCLIHelpersUsed()
	return nil
}
```
