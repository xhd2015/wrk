## Expected

- Non-zero exit code (dirty worktree).
- `projects.json` contains the main repo with `source: "auto"` despite command failure.
- An event is appended to `events.jsonl` with non-zero `exit_code`.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertProjectsCount(t, req.WrkHome, 1)
	assertProjectRecorded(t, req.WrkHome, req.MainRepo, "auto")
	assertEventsCount(t, req.WrkHome, 1)
	assertLastEvent(t, req.WrkHome, "done", resp.ExitCode, req.MainRepo, req.WtDir, []string{"--done"})
}
```