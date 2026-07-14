# Scenario

**Feature**: terminal + agent only sends agent-run as iTerm follow-up (with `--dir`)

```
wrk -t 'ship feature' --new-terminal --open-in-agent
  -> iterm FollowUpCommands contain agent-run --dir <wt> …
  -> outer agent-run binary NOT executed
```

## Steps

1. Run `--new-terminal --open-in-agent` with task.

```go
func Setup(t *testing.T, req *Request) error {
	req.TaskDesc = "ship feature"
	req.TaskFlag = "-t"
	req.Args = []string{"--new-terminal", "--open-in-agent"}
	return nil
}
```
