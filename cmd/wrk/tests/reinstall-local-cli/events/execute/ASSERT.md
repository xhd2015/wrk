
## Expected

- Exit code 0 (successful execute with zero install attempts).
- Last `events.jsonl` event has `command: "reinstall-local"`, `exit_code: 0`.
- Event `args` include `--reinstall-local` and do **not** include `--dry-run`.
- Event `work_dir` resolves to the module root (process cwd).
- Event `ts` is non-empty RFC3339.

## Side Effects

- Skip-only execute (no go install); event appended under `{WRK_HOME}/events.jsonl`.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)

	assertLastEventCommandReinstallLocal(t, req.WrkHome, 0, []string{
		"--reinstall-local",
	})

	events := readEvents(t, req.WrkHome)
	ev := events[len(events)-1]
	if eventArgsContain(ev.Args, "--dry-run") {
		t.Fatalf("execute event args must not include --dry-run, got %v", ev.Args)
	}
	wantWD := resolvePath(t, req.ModuleRoot)
	gotWD := resolvePath(t, ev.WorkDir)
	if gotWD != wantWD {
		t.Fatalf("event work_dir: want %q, got %q", wantWD, gotWD)
	}
}
```
