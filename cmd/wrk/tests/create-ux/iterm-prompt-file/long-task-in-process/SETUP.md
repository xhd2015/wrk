# Scenario

**Feature**: long `-t` + in-process agent uses `--prompt-file`

```
wrk -t '<700×x>' --open-in-agent
  -> outer agent-run argv has --prompt-file=<abs>
  -> file body is /brainstorm <full task>
```

## Steps

1. Grouping already set long TaskDesc + mocks.
2. `--open-in-agent` only (no terminal).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--open-in-agent"}
	return nil
}
```
