## Expected

- Exit code 0 (PR path succeeds).
- Last `events.jsonl` event has `command: "pr"`, `exit_code: 0`.
- Event `args` include `--pr`, and preferably `--title` / `--comment`.

## Side Effects

- Event appended for the bare pr primary.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	events := readPrEvents(t, req.WrkHome)
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	ev := events[len(events)-1]
	if ev.Command != "pr" {
		t.Fatalf("event command: want %q, got %q (all=%v)", "pr", ev.Command, events)
	}
	if ev.ExitCode != 0 {
		t.Fatalf("event exit_code: want 0, got %d", ev.ExitCode)
	}
	foundPr := false
	for _, a := range ev.Args {
		if a == "--pr" {
			foundPr = true
			break
		}
	}
	if !foundPr {
		t.Fatalf("event args should include --pr, got %v", ev.Args)
	}
}
```
