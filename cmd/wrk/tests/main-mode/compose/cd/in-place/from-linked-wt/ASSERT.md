## Expected

- Exit 0.
- Stdout empty (in-place; no path printed).
- Follow-up file is exactly `cd <mainRepo>\n`.
- No bare-main already-at-root notice on stderr.

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
	assertNoBareMainNotice(t, resp.Stderr)
}
```
