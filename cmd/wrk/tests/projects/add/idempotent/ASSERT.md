## Expected

- Exit code 0.
- `projects.json` contains exactly one entry for the main repo.
- Entry `source` remains `"auto"` (first record wins).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	assertProjectsCount(t, req.WrkHome, 1)
	assertProjectRecorded(t, req.WrkHome, req.MainRepo, "auto")
}
```