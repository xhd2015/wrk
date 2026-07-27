
## Expected

- Exit code 0.
- Stdout is empty (in-place; no path printed).
- Follow-up file is exactly `cd <abs>\n`.
- No interactive shell launched (follow-up only).

## Side Effects

- Follow-up file written; no shell child required.

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
}
```
