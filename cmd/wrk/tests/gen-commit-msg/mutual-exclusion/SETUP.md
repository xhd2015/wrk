# Scenario

**Feature**: wrk --gen-commit-msg exclusives vs allowed pipeline compose

```
# still exclusive with non-pipeline modes:
wrk --gen-commit-msg --status -> non-zero; mutually exclusive

# allowed multi-stage (activeRoot pipeline; no --done required):
wrk --gen-commit-msg --sync [--dry-run] -> not mutually exclusive
# covered by with-sync/ (exit 0 dry-run path)
```

## Steps

1. Descendants either reject list/status-style modes or allow pipeline stages with gen-commit.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: mutex leaves share root helpers; with-sync needs git (leaf sets RepoDir).
	ensureGenCommitMsgHelpersUsed()
	return nil
}
```
