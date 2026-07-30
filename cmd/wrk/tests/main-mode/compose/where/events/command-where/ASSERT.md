## Expected

- Exit 0; stdout main path.
- Last `events.jsonl` event has `command: "where"` (not `"main"`), `exit_code: 0`.
- Event `args` include both `--main` and `--where`.
- Event `main_repo` resolves to main checkout when present.

## Side Effects

- Event appended under `{WRK_HOME}/events.jsonl`.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertStdoutMainPath(t, resp.Stdout, req.MainRepo)

	assertLastEventPartner(t, req.WrkHome, "where", 0, "--main", "--where")

	events := readEvents(t, req.WrkHome)
	ev := events[len(events)-1]
	if ev.MainRepo != "" {
		wantMain := resolvePath(t, req.MainRepo)
		gotMain := resolvePath(t, ev.MainRepo)
		if gotMain != wantMain {
			t.Fatalf("event main_repo: want %q, got %q", wantMain, gotMain)
		}
	}
}
```
