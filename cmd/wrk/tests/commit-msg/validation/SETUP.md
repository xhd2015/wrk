# Scenario

**Feature**: early flag validation for manual -m/--message and --commit message source

```
# pure flag-layer rejects (neutral cwd; no git required unless noted)
-m / --message without --commit
--gen-commit-msg with -m
--commit alone (no gen, no -m)
empty / whitespace-only -m value
```

## Steps

1. Inherit root Setup.
2. Leaves set Args and InProcess for Capture.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: validation leaves share root helpers.
	ensureCommitMsgHelpersUsed()
	return nil
}
```
