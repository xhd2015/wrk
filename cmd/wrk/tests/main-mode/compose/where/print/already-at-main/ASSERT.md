## Expected

- Exit 0.
- Stdout is main abs path + `\n` (always print when already at main).
- Stderr empty — **no** bare `--main` already-at-root notice.
- Fake shell not launched.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertStdoutMainPath(t, resp.Stdout, req.MainRepo)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty on success, got %q", resp.Stderr)
	}
	assertNoBareMainNotice(t, resp.Stderr)
	assertFakeShellNotLaunched(t, req)
}
```
