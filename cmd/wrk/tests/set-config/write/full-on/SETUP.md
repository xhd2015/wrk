# Scenario

**Feature**: enable window, terminal, and agent defaults in one write

```
wrk --set-config --create --new-window --new-terminal --open-in-agent
  -> window.mode=new, terminal.mode=new, agent.enabled=true + defaults
```

## Steps

1. No prior config.
2. Run full-on set-config flags.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = setConfigArgs("--create", "--new-window", "--new-terminal", "--open-in-agent")
	return nil
}
```
