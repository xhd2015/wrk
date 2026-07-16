# Scenario

**Feature**: successful bare `wrk --push` records events.jsonl `command: "push"`

```
myrepo + origin -> wrk --push
  -> events.jsonl last: command=push, exit_code=0, args include --push
```

## Steps

1. Seed main with bare origin.
2. Run `wrk --push`.
3. Assert last events.jsonl event (do not re-invoke wrk before read).

```go
func Setup(t *testing.T, req *Request) error {
	setupPushMainWithOrigin(t, req)
	req.Args = []string{"--push"}
	return nil
}
```
