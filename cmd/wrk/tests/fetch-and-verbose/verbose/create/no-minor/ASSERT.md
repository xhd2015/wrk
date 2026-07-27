## Expected

- Exit code 0.
- Stderr has **no** `rev-parse` or `status` log lines.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	assertStderrNoGitSubcommand(t, resp.Stderr, "rev-parse")
	assertStderrNoGitSubcommand(t, resp.Stderr, "status")
}
```