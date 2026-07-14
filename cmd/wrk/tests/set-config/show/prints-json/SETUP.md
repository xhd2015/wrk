# Scenario

**Feature**: show prints create section JSON after a write

```
seed create UX -> wrk --set-config --show
  -> stdout includes window/terminal/agent fields as JSON
```

## Steps

1. Seed a full create section.
2. Run `--set-config --show`.

```go
func Setup(t *testing.T, req *Request) error {
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
	req.Args = setConfigArgs("--show")
	return nil
}
```
