## Expected

- Exit code 0.
- Stdout (trimmed) equals the resolved main repo absolute path (single line).
- `projects.json` contains one entry with `source: "manual"`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	assertStdoutExactPath(t, resp.Stdout, resolvePath(t, req.MainRepo))
	assertProjectsCount(t, req.WrkHome, 1)
	assertProjectRecorded(t, req.WrkHome, req.MainRepo, "manual")
}
```