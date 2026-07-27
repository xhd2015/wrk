# Scenario

**Feature**: full pipeline window + terminal + agent follow-up

```
wrk -t 'ship feature' --new-window --new-terminal --open-in-agent
  -> create -> space -> iterm(+agent send); outer agent not exec'd
```

## Steps

1. Run full flag set with task.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TaskDesc = "ship feature"
	req.TaskFlag = "-t"
	req.Args = []string{"--new-window", "--new-terminal", "--open-in-agent"}
	return nil
}
```
