## Expected

- Exit code 0 (push succeeds).
- Last `events.jsonl` event has `command: "push"`, `exit_code: 0`.
- Event `args` include `--push`.

## Side Effects

- Event appended for the bare push primary.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	events := readPushEvents(t, req.WrkHome)
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	ev := events[len(events)-1]
	if ev.Command != "push" {
		t.Fatalf("event command: want %q, got %q (all=%v)", "push", ev.Command, events)
	}
	if ev.ExitCode != 0 {
		t.Fatalf("event exit_code: want 0, got %d", ev.ExitCode)
	}
	found := false
	for _, a := range ev.Args {
		if a == "--push" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("event args should include --push, got %v", ev.Args)
	}
}
```
