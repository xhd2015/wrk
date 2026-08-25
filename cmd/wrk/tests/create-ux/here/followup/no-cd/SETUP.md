# Scenario

**Feature**: `--here --no-cd` emits agent-run only (no cd line)

```
WRK_FOLLOWUP_FILE open
wrk -t 'no cd here' --open-in-agent --here --no-cd
  -> follow-up: agent-run only
```

## Steps

1. Enable follow-up; run with `--here --open-in-agent --no-cd`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TaskDesc = "no cd here"
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent", "--here", "--no-cd"}
	return nil
}
```
