# Scenario

**Feature**: `--no-open-in-agent` disables agent after it was enabled

```
seed agent on -> wrk --set-config --create --no-open-in-agent
  -> agent.enabled=false; other create keys preserved
```

## Steps

1. Seed full create UX with agent on.
2. Run `--no-open-in-agent`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	writeSetConfigRaw(t, req.WrkHome, `{
  "version": 1,
  "create": {
    "window": { "mode": "new" },
    "terminal": { "mode": "new" },
    "agent": {
      "enabled": true,
      "runner": "grok-tty",
      "prompt_template": "/brainstorm ${task}",
      "args": ["--session-id-from-prompt", "--no-submit", "--open"]
    }
  }
}
`)
	req.Args = setConfigArgs("--create", "--no-open-in-agent")
	return nil
}
```
