## Expected

- Exit 0.
- Stdout empty.
- Follow-up is `cd <main>\n` (still runCd at main root).
- Stderr must **not** be bare-main notice-only (no "already at main repository root" short-circuit without follow-up).
- Fake shell not launched (in-place path).

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
	assertFakeShellNotLaunched(t, req)
}
```
