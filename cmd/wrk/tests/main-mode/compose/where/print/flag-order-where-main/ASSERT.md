## Expected

- Exit 0; stdout main abs path + `\n`; empty stderr; no shell.
- Same contract as `--main --where` (flag order free for bools).

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
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertFakeShellNotLaunched(t, req)
}
```
