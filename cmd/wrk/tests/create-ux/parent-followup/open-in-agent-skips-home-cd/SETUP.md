# Scenario

**Feature**: create from FakeHome with `--open-in-agent` skips parent home-gated auto-cd

```
FakeHome (cwd) + WRK_FOLLOWUP_FILE
wrk <mainRepo> -t 'ship feature' --open-in-agent
  -> stdout worktree path\n
  -> agent-run invoked with --dir
  -> follow-up file empty (no parent cd)
```

## Steps

1. Setup create from FakeHome with follow-up channel.
2. Run with `--open-in-agent` and a task.

```go
func Setup(t *testing.T, req *Request) error {
	setupCreateUXFromFakeHome(t, req)
	req.TaskDesc = "ship feature"
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent"}
	return nil
}
```
