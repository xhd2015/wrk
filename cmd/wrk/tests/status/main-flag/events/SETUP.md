# Scenario

**Feature**: successful --main --status records events.jsonl command "status"

```
wrk --main --status -> events.jsonl last:
  command=status, exit_code=0, args include --main and --status
```

## Preconditions

- Isolated `WRK_HOME`; successful composition from a git checkout.

## Steps

1. Create checkout (external wt cwd is fine).
2. Run `--main --status`.
3. Assert last event fields (do not run a reference status before reading events).

## Context

- `command` is `"status"` (not `"main"`); shell is not launched.

```go
func Setup(t *testing.T, req *Request) error {
	ensureMainFlagHelpersUsed()
	return nil
}
```