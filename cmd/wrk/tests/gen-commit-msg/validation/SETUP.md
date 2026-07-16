# Scenario

**Feature**: wrk --gen-commit-msg flag validation (library + wire)

```
# library validation surfaces through wrk binary stderr
--no-verify without --commit -> error
--agent-runner codex (even with --dry-run) -> unsupported agent runner
```

## Steps

1. Inherit root Setup.
2. Leaves set Args (and optionally stage a repo for dry-run validation paths).

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: validation leaves share root helpers (stage for runner path).
	ensureGenCommitMsgHelpersUsed()
	return nil
}
```
