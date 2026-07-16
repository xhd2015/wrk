# Scenario

**Feature**: wrk --gen-commit-msg is mutually exclusive with other standalone modes

```
wrk --gen-commit-msg combined with other modes -> error
```

## Steps

1. Descendants combine `--gen-commit-msg` with another exclusive mode (e.g. `--status`).

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: mutex leaves share root helpers; no git required for mode selection.
	ensureGenCommitMsgHelpersUsed()
	return nil
}
```
