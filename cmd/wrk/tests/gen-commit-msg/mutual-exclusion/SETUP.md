# Scenario

**Feature**: wrk --gen-commit-msg is mutually exclusive with other standalone modes (no primary)

```
wrk --gen-commit-msg combined with other modes -> error
# still exclusive after P2 primary compose:
wrk --gen-commit-msg --status
wrk --gen-commit-msg --sync   # --sync is post-only with --done/--merge-back
```

## Steps

1. Descendants combine `--gen-commit-msg` with another exclusive mode (e.g. `--status`, bare `--sync`).

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: mutex leaves share root helpers; no git required for mode selection.
	ensureGenCommitMsgHelpersUsed()
	return nil
}
```
