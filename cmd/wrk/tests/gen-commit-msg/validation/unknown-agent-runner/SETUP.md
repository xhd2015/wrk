# Scenario

**Feature**: unknown --agent-runner is rejected even with --dry-run

```
# runner validation before pure-plan success
repo/ (1 staged) -> wrk --gen-commit-msg --dry-run --agent-runner codex
  -> non-zero; unsupported agent runner / codex
```

## Preconditions

- Stage one file so a successful dry-run would otherwise proceed.
- Runner name `codex` is not supported by gen-commit-msg (only `opencode`).

## Steps

1. Stage one text file.
2. Run with `--dry-run --agent-runner codex`.

```go
func Setup(t *testing.T, req *Request) error {
	stageOneTextFile(t, req)
	req.Args = []string{"--gen-commit-msg", "--dry-run", "--agent-runner", "codex"}
	return nil
}
```
