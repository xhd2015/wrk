## Expected

- Exit code 0.
- Stdout equals the **main repo** path, not the linked worktree path.
- `projects.json` contains zero entries.
- Last event is `command: "rm"` with `args: ["--rm", "<linkedWt>"]`.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	assertStdoutExactPath(t, resp.Stdout, resolvePath(t, req.MainRepo))
	assertProjectsCount(t, req.WrkHome, 0)
	assertLastEvent(t, req.WrkHome, "rm", 0, "", req.WorkRoot, []string{"--rm", req.WtDir})
}
```
