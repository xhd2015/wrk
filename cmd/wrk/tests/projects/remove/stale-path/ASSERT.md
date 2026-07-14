## Expected

- Exit code 0.
- Stdout equals the recorded main repo absolute path (single line).
- `projects.json` contains zero entries.
- Last event is `command: "rm"` with `exit_code: 0`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	assertStdoutExactPath(t, resp.Stdout, req.MainRepo)
	assertProjectsCount(t, req.WrkHome, 0)
	assertLastEvent(t, req.WrkHome, "rm", 0, "", req.WorkRoot, []string{"--rm", req.MainRepo})
}
```
