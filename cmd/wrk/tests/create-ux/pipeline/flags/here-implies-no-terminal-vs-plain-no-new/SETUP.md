# Scenario

**Feature**: plain `--no-new-window --no-new-terminal --open-in-agent` still runs agent in-process (not `--here` follow-up)

```
# channel open but --here absent → still in-process agent (backward compatible)
WRK_FOLLOWUP_FILE open
wrk -t 'plain no-new' --open-in-agent --no-new-window --no-new-terminal
  -> outer agent-run invoked
  -> follow-up empty (agent UX skips home-gated cd; no --here emit)
```

## Steps

1. Enable follow-up; run plain negatives without `--here`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	enableCreateUXFollowup(t, req)
	req.TaskDesc = "plain no-new"
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent", "--no-new-window", "--no-new-terminal"}
	return nil
}
```
