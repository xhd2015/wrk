# Scenario

**Feature**: successful wrk --main appends events.jsonl with command "main"

```
cwd = main subdir; fake shell exit 0
wrk --main -> last event command=main, exit_code=0, args include --main
```

## Preconditions

- Uses launch success path so shell and main-mode handling both run.
- Fake bash required (nested shell).

## Steps

1. Create main repo + subdir; install fake bash (exit 0).
2. Run `wrk --main`.
3. Assert last `events.jsonl` event.

## Context

- Event `command` is the mode name `"main"`; `args` record CLI flags including `--main`.

```go
func Setup(t *testing.T, req *Request) error {
	// Leaves configure layout; keep helpers referenced.
	ensureMainHelpersUsed()
	return nil
}
```
