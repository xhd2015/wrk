# Scenario

**Feature**: `--no-open-in-agent` disables agent from config for this run

```
config agent on + terminal new; wrk --no-open-in-agent
  -> iterm without agent follow-up; outer agent not run
```

## Steps

1. Write config terminal new + agent on (no window).
2. Run bare create with `--no-open-in-agent`.

```go
func Setup(t *testing.T, req *Request) error {
	writeCreateUXConfig(t, req.WrkHome, map[string]interface{}{
		"terminal": map[string]interface{}{"mode": "new"},
		"agent": map[string]interface{}{
			"enabled":         true,
			"runner":          "grok-tty",
			"prompt_template": "/brainstorm ${task}",
			"args":            []string{"--session-id-from-prompt", "--no-submit", "--open"},
		},
	})
	req.TaskDesc = "ship feature"
	req.TaskFlag = "-t"
	req.Args = []string{"--no-open-in-agent"}
	return nil
}
```
