# Scenario

**Feature**: bare non-TTY `wrk` records events.jsonl `command: "dashboard"`

```
myrepo -> wrk (no args)
  -> exit 0
  -> events.jsonl: command=dashboard, exit_code=0
```

## Steps

1. Init main repo.
2. Run bare `wrk`.
3. Assert last event is dashboard.

```go
func Setup(t *testing.T, req *Request) error {
	setupDashboardMainRepo(t, req)
	req.Args = nil
	return nil
}
```
