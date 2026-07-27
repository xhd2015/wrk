## Expected

- Exit 0 (successful status of main).
- Last `events.jsonl` event: `command == "status"`, `exit_code == 0`.
- Event `args` include both `--main` and `--status`.

## Side Effects

- Event appended under `{WRK_HOME}/events.jsonl`.
- No nested shell.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	// Do not call assertStdoutEqualsMainStatus here — that would append another event.
	assertExitZeroEmptyStderr(t, resp, err)
	if resp.Stdout == "" {
		t.Fatal("expected non-empty status stdout on success")
	}
	assertLastEventCommandStatusWithMain(t, req.WrkHome, 0)
}
```