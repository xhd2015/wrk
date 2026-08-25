# Scenario

**Feature**: `--here` create UX — no window/terminal; agent via shell follow-up or nested fallback

```
wrk -t 'task' --open-in-agent --here
  -> create; no space; no iterm
  -> WRK_FOLLOWUP_FILE: cd + agent-run  (or fallback nested shell)
```

## Steps

- Grouping installs repo + mocks; leaves set `--here` variants.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupMainRepoForCreateUX(t, req)
	installCreateUXMocks(t, req, "darwin")
	return nil
}
```
