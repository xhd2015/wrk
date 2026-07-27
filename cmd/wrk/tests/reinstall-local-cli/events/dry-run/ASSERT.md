
## Expected

- Exit code 0 (successful dry-run plan).
- Last `events.jsonl` event has `command: "reinstall-local"`, `exit_code: 0`.
- Event `args` include both `--reinstall-local` and `--dry-run`.
- Event `work_dir` resolves to the module root (process cwd).
- Event `ts` is non-empty RFC3339.

## Side Effects

- Dry-run plan only (no real install); event appended under `{WRK_HOME}/events.jsonl`.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)

	assertLastEventCommandReinstallLocal(t, req.WrkHome, 0, []string{
		"--reinstall-local",
		"--dry-run",
	})

	events := readEvents(t, req.WrkHome)
	ev := events[len(events)-1]
	wantWD := resolvePath(t, req.ModuleRoot)
	gotWD := resolvePath(t, ev.WorkDir)
	if gotWD != wantWD {
		t.Fatalf("event work_dir: want %q, got %q", wantWD, gotWD)
	}
}
```
