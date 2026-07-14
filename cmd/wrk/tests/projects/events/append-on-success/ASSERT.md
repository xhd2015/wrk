## Expected

- Exit code 0.
- `events.jsonl` contains one event with `command: "create"`, `exit_code: 0`.
- Event `main_repo` and `work_dir` match the main repo path.
- Event `args` is empty (no flags).

## Side Effects

- Worktree is created (create mode side effect).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	assertEventsCount(t, req.WrkHome, 1)
	assertLastEvent(t, req.WrkHome, "create", 0, req.MainRepo, req.MainRepo, []string{})
}
```