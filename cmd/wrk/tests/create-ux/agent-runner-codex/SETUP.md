# Scenario

**Feature**: `--agent-runner codex` (or `codex-tty`) uses `$brainstorm` skill prompt

```
wrk --open-in-agent --agent-runner codex
  -> agent-run prompt uses $brainstorm instead of /brainstorm
```

## Steps

- Empty config; flags only.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	return nil
}
```
