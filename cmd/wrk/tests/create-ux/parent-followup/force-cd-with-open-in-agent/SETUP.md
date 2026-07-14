# Scenario

**Feature**: `--force-cd` with `--open-in-agent` still writes parent follow-up cd

```
FakeHome (cwd) + WRK_FOLLOWUP_FILE
wrk <mainRepo> -t 'ship feature' --force-cd --open-in-agent
  -> agent-run with --dir
  -> follow-up: cd <abs-worktree>  (--force-cd wins over UX skip)
```

## Steps

1. Setup create from FakeHome with follow-up channel.
2. Run with `--force-cd --open-in-agent` and a task.

```go
func Setup(t *testing.T, req *Request) error {
	setupCreateUXFromFakeHome(t, req)
	req.TaskDesc = "ship feature"
	req.TaskFlag = "-t"
	req.Args = []string{"--force-cd", "--open-in-agent"}
	return nil
}
```
