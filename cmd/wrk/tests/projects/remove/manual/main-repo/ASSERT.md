## Expected

- Exit code 0.
- Stdout (trimmed) equals the resolved main repo absolute path (single line).
- `projects.json` contains zero entries.
- Last `events.jsonl` event has `command: "rm"`, `exit_code: 0`, `args: ["--rm", "<mainRepo>"]`.

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
	assertLastEvent(t, req.WrkHome, "rm", 0, "", req.WorkRoot, []string{"--rm", req.MainRepo})
}
```
