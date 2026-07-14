# Scenario

**Feature**: create from FakeHome with `--new-terminal` (no agent) skips parent home-gated auto-cd

```
FakeHome (cwd) + WRK_FOLLOWUP_FILE
wrk <mainRepo> --new-terminal
  -> stdout worktree path\n
  -> iterm ForceNew at worktree
  -> follow-up file empty (terminal UX skips home-gated parent cd)
```

## Steps

1. Setup create from FakeHome with follow-up channel.
2. Run with `--new-terminal` only (no agent).

```go
func Setup(t *testing.T, req *Request) error {
	setupCreateUXFromFakeHome(t, req)
	req.Args = []string{"--new-terminal"}
	return nil
}
```
