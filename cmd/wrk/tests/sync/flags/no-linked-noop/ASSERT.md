## Expected Output

```
synced: 0 into main, 0 into worktrees, 0 skipped
```

## Expected

- Exit code 0.
- Stdout is exactly the one-line summary above (trailing `\n`).
- Summary mentions sync into main / into worktrees / skipped counts (all zero).
- Phase 1: no `git merge --ff-only`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertOutputExact(t, resp.Stdout, syncStdoutV2(syncSummaryZero))
}
```
