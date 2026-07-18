## Expected

- Exit 0; worktree created (create mode).
- Last `events.jsonl` event has `command: "create"`, `exit_code: 0`.
- Event `args` includes `--new`.
- Event `main_repo` / `work_dir` match the main repo path (normalized).

## Side Effects

- One events.jsonl line for this invocation.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	wt := wantDashboardCreateWorktree(req)
	assertStdoutExactPath(t, resp.Stdout, wt)
	assertFileExists(t, wt)

	events := readDashboardEvents(t, req.WrkHome)
	if len(events) == 0 {
		t.Fatal("expected events.jsonl entry for wrk --new")
	}
	ev := events[len(events)-1]
	if ev.Command != "create" {
		t.Fatalf("event command: want %q, got %q", "create", ev.Command)
	}
	if ev.ExitCode != 0 {
		t.Fatalf("event exit_code: want 0, got %d", ev.ExitCode)
	}
	foundNew := false
	for _, a := range ev.Args {
		if a == "--new" {
			foundNew = true
			break
		}
	}
	if !foundNew {
		t.Fatalf("event args should include --new, got %v", ev.Args)
	}
	if resolvePath(t, ev.MainRepo) != resolvePath(t, req.MainRepo) {
		t.Fatalf("event main_repo: want %q, got %q", req.MainRepo, ev.MainRepo)
	}
	if resolvePath(t, ev.WorkDir) != resolvePath(t, req.MainRepo) {
		t.Fatalf("event work_dir: want %q, got %q", req.MainRepo, ev.WorkDir)
	}
	if ev.TS == "" {
		t.Fatal("event missing ts")
	}
	if _, err := time.Parse(time.RFC3339, ev.TS); err != nil {
		t.Fatalf("event ts not RFC3339: %q", ev.TS)
	}
}
```
