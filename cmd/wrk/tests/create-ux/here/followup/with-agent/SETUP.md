# Scenario

**Feature**: `--here --open-in-agent` writes follow-up cd + agent-run

```
WRK_FOLLOWUP_FILE open
wrk -t 'ship here' --open-in-agent --here
  -> worktree path on stdout
  -> follow-up: cd <wt>\nagent-run run --dir <wt> …
  -> outer agent-run not invoked; no iterm/space
```

## Steps

1. Enable follow-up channel; run with `--here --open-in-agent`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TaskDesc = "ship here"
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent", "--here"}
	return nil
}
```
