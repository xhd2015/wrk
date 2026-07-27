## Expected

- Exit code 0.
- Stdout is empty.
- `projects.json` is absent or contains zero entries.
- Last event is `command: "rm"` with `exit_code: 0`.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	assertStdoutEmpty(t, resp.Stdout)
	assertProjectsCount(t, req.WrkHome, 0)
	assertLastEvent(t, req.WrkHome, "rm", 0, "", req.WorkRoot, []string{"--rm", req.MainRepo})
}
```
