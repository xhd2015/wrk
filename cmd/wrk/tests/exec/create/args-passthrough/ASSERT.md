## Expected

- Exit code 0.
- Stdout: worktree path line, then `--task` (echo output), trailing `\n`.
- Worktree still created under default naming (no task slug — `--task` was not a wrk flag).
- Stderr empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	// No --task slug: default date-only name.
	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assertPathThenChildStdout(t, resp.Stdout, wantPath, "--task")

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	// Task-slug path must NOT exist.
	taskSlugPath := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, "task", 0)
	assertFileNotExists(t, taskSlugPath)

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
