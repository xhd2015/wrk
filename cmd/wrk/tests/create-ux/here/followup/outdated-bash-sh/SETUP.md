# Scenario

**Feature**: `--here` with outdated bash.sh falls back to in-process agent

```
WRK_FOLLOWUP_FILE open; bash.sh lacks agent-run whitelist
wrk -t 'outdated wrapper' --open-in-agent --here
  -> warning: bash integration is outdated …
  -> follow-up: cd only
  -> outer agent-run invoked (in-process fallback)
```

## Steps

1. Enable follow-up channel, then overwrite bash.sh with a pre-`--here` stub.
2. Run `--here --open-in-agent`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	installCreateUXBashIntegrationScript(t, req, false) // outdated (cd only)
	req.TaskDesc = "outdated wrapper"
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent", "--here"}
	return nil
}
```
