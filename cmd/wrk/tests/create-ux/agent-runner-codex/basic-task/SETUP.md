# Scenario

**Feature**: `--agent-runner codex` uses `$brainstorm` prompt instead of `/brainstorm`

```
wrk -t 'fix bug' --open-in-agent --agent-runner codex
  -> create; agent-run argv uses --agent-runner=codex-tty
  -> prompt = $brainstorm fix bug (not /brainstorm fix bug)
```

## Steps

1. Run with task + `--open-in-agent` + `--agent-runner codex`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TaskDesc = "fix bug"
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent", "--agent-runner", "codex"}
	return nil
}
```
