## Expected

- Exit code 0.
- Last `events.jsonl` event has `command: "projects-dep-graph"`, `exit_code: 0`.
- Event `args` include `--projects-dep-graph`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)

	events := readEvents(t, req.WrkHome)
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	ev := events[len(events)-1]
	if ev.Command != "projects-dep-graph" {
		t.Fatalf("event command: want %q, got %q (all=%v)", "projects-dep-graph", ev.Command, events)
	}
	if ev.ExitCode != 0 {
		t.Fatalf("event exit_code: want 0, got %d", ev.ExitCode)
	}
	found := false
	for _, a := range ev.Args {
		if a == "--projects-dep-graph" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("event args should include --projects-dep-graph, got %v", ev.Args)
	}
}
```
