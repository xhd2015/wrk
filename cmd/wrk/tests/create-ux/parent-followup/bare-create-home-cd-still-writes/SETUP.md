# Scenario

**Feature**: bare create from FakeHome still writes home-gated parent auto-cd (regression)

```
FakeHome (cwd) + WRK_FOLLOWUP_FILE
wrk <mainRepo>
  -> stdout worktree path\n
  -> no agent / no terminal UX
  -> follow-up: cd <abs-worktree>
```

## Steps

1. Setup create from FakeHome with follow-up channel.
2. Run bare create (no UX flags).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	setupCreateUXFromFakeHome(t, req)
	req.Args = nil
	return nil
}
```
