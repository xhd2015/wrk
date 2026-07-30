## Expected

- Exit 0; in-place follow-up success.
- Last `events.jsonl` event has `command: "cd"` (not `"main"`), `exit_code: 0`.
- Event `args` include both `--main` and `--cd`.

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

	assertLastEventPartner(t, req.WrkHome, "cd", 0, "--main", "--cd")
}
```
