# Scenario

**Feature**: quoted task is shell-safe in iterm agent follow-up line

```
wrk -t 'fix "quoted" task' --new-terminal --open-in-agent
  -> follow-up command embeds shell-safe prompt token
```

## Steps

1. Same adversarial task.
2. Terminal + agent path.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TaskDesc = `fix "quoted" task`
	req.TaskFlag = "-t"
	req.Args = []string{"--new-terminal", "--open-in-agent"}
	return nil
}
```
