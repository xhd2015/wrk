# Scenario

**Feature**: double quotes in task are safe for current-process agent-run argv

```
wrk -t 'fix "quoted" task' --open-in-agent
  -> agent-run last arg is /brainstorm fix "quoted" task (argv element, not shell-split)
```

## Steps

1. Task contains embedded double quotes.
2. Run `--open-in-agent` only (argv path, not shell string).

```go
func Setup(t *testing.T, req *Request) error {
	req.TaskDesc = `fix "quoted" task`
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent"}
	return nil
}
```
