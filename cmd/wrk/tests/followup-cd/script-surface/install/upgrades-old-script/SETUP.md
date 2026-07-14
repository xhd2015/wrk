# Scenario

**Feature**: install overwrites completion-only bash.sh with wrapper script

```
pre-seed bash.sh with only _wrk + complete
wrk --bash-integration --install
  -> bash.sh rewritten with wrk() + WRK_FOLLOWUP_FILE
```

## Steps

1. Pre-seed `integration/bash.sh` with completion-only stub.
2. Run install.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "install")
	req.PreExistingBashSh = minimalCompletionOnlyBashSh()
	return nil
}
```
