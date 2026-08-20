# Scenario

**Feature**: `--agent-runner grok` still uses `/brainstorm` prompt (regression guard)

```
wrk --open-in-agent --agent-runner grok
  -> agent-run prompt uses /brainstorm (unchanged default)
```

## Steps

- Empty config; flags only.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	return nil
}
```
