## Expected Output

```
would: synced: 0 into main, 0 into worktrees, 0 skipped
```

## Expected

- Exit code 0.
- Stdout is exactly the one-line dry-run summary above (trailing `\n`).
- Phase 1: no merges / no worktree mutations (main-only fixture has nothing to change).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertOutputExact(t, resp.Stdout, syncStdoutV2(syncWouldSummaryZero))
}
```
