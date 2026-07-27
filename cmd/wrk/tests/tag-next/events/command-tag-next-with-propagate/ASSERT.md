
## Expected

- Exit code 0.
- Last `events.jsonl` event has `command: "tag-next"` (primary; **not** `propagate-tags`).
- Event `args` include both `--tag-next` and `--propagate-tags`.
- Event `exit_code` is 0.

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
	if ev.Command != "tag-next" {
		t.Fatalf("event command: want %q, got %q (all=%v)", "tag-next", ev.Command, events)
	}
	if ev.ExitCode != 0 {
		t.Fatalf("event exit_code: want 0, got %d", ev.ExitCode)
	}
	var foundTagNext, foundPropagate bool
	for _, a := range ev.Args {
		if a == "--tag-next" {
			foundTagNext = true
		}
		if a == "--propagate-tags" {
			foundPropagate = true
		}
	}
	if !foundTagNext || !foundPropagate {
		t.Fatalf("event args should include --tag-next and --propagate-tags, got %v", ev.Args)
	}
	wantMain := resolveRepoPath(t, req.MainRepo)
	gotMain := resolveRepoPath(t, ev.MainRepo)
	if gotMain != wantMain {
		t.Fatalf("event main_repo: want %q, got %q", wantMain, gotMain)
	}
}
```
