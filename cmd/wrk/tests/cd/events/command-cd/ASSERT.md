
## Expected

- Exit 0.
- Last `events.jsonl` event has `command: "cd"`, `exit_code: 0`.
- Event `args` include `--cd`.
- Event `work_dir` is the resolved absolute target path.

## Side Effects

- Follow-up file written (in-place); event appended.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStdout(t, resp.Stdout)
	assertFollowupCDLine(t, req, req.MainRepo)

	events := readEvents(t, req.WrkHome)
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	ev := events[len(events)-1]
	if ev.Command != "cd" {
		t.Fatalf("event command: want %q, got %q (all=%v)", "cd", ev.Command, events)
	}
	if ev.ExitCode != 0 {
		t.Fatalf("event exit_code: want 0, got %d", ev.ExitCode)
	}
	foundCD := false
	for _, a := range ev.Args {
		if a == "--cd" {
			foundCD = true
			break
		}
	}
	if !foundCD {
		t.Fatalf("event args should include --cd, got %v", ev.Args)
	}
	wantWD := resolvePath(t, req.MainRepo)
	gotWD := resolvePath(t, ev.WorkDir)
	if gotWD != wantWD {
		t.Fatalf("event work_dir: want %q, got %q", wantWD, gotWD)
	}
}
```
