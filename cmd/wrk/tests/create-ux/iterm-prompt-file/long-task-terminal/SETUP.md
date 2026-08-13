# Scenario

**Feature**: long `-t` + terminal+agent spills prompt to `--prompt-file`

```
wrk -t '<700×x>' --new-terminal --open-in-agent
  -> iterm FollowUpCommands use --prompt-file, not the 700-x body
```

## Steps

1. Grouping already set long TaskDesc + mocks.
2. Add terminal+agent flags.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--new-terminal", "--open-in-agent"}
	return nil
}
```
