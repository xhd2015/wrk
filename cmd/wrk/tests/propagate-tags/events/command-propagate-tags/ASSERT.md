
## Expected

- Exit code 0.
- Last `events.jsonl` event has `command: "propagate-tags"`, `exit_code: 0`.
- Event `args` include `--propagate-tags`.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	events := readEvents(t, req.WrkHome)
	if len(events) == 0 {
		t.Fatal("expected at least one event")
	}
	ev := events[len(events)-1]
	if ev.Command != "propagate-tags" {
		t.Fatalf("event command: want %q, got %q (all=%v)", "propagate-tags", ev.Command, events)
	}
	if ev.ExitCode != 0 {
		t.Fatalf("event exit_code: want 0, got %d", ev.ExitCode)
	}
	found := false
	for _, a := range ev.Args {
		if a == "--propagate-tags" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("event args should include --propagate-tags, got %v", ev.Args)
	}
	wantMain := resolvePath(t, req.SourcePath)
	gotMain := resolvePath(t, ev.MainRepo)
	if gotMain != wantMain {
		t.Fatalf("event main_repo: want %q, got %q", wantMain, gotMain)
	}
}
```
