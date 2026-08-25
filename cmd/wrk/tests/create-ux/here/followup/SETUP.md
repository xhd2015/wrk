# Scenario

**Feature**: `--here` with `WRK_FOLLOWUP_FILE` emits cd + agent-run (no in-process agent)

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	enableCreateUXFollowup(t, req)
	return nil
}
```
