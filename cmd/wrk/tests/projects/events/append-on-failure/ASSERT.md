## Expected

- Non-zero exit code (`not a linked worktree`).
- `events.jsonl` contains one event with `command: "done"`, `exit_code` matching process exit.
- Event `args` is `["--done"]`.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertEventsCount(t, req.WrkHome, 1)
	assertLastEvent(t, req.WrkHome, "done", resp.ExitCode, req.MainRepo, req.MainRepo, []string{"--done"})
}
```