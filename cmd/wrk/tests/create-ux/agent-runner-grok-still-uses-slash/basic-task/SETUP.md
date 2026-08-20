# Scenario

**Feature**: `--agent-runner grok` still uses `/brainstorm` prompt (regression guard)

```
wrk -t 'fix bug' --open-in-agent --agent-runner grok
  -> create; agent-run argv uses --agent-runner=grok-tty
  -> prompt = /brainstorm fix bug (unchanged default)
```

## Steps

1. Run with task + `--open-in-agent` + `--agent-runner grok`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TaskDesc = "fix bug"
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent", "--agent-runner", "grok"}
	return nil
}
```
