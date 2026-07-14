## Expected

- Exit code 0.
- Stderr has **no** `fetch` verbose log line.
- No `Remote:` in stdout.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	assertStdoutNoRemoteField(t, resp.Stdout)
	assertStderrNoGitSubcommand(t, resp.Stderr, "fetch")
}
```