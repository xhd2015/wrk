## Expected

- Exit 0; empty stdout; follow-up `cd <main>\n`.
- Same contract as `--main --cd` (flag order free).

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
