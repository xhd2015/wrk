## Expected

- Exit 0; nested shell at main root (launch success).
- Last `events.jsonl` event has `command: "main"`, `exit_code: 0`.
- Event `args` include `--main`.
- Event `main_repo` resolves to the main checkout when present.

## Side Effects

- Nested shell launched; event appended under `{WRK_HOME}/events.jsonl`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertMinimalLaunchUX(t, resp)
	assertFakeShellLaunched(t, req)
	assertFakeShellCwd(t, req, req.MainRepo)

	assertLastEventCommandMain(t, req.WrkHome, 0)

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
