## Expected

- Exit code 0.
- Stdout is worktree path (non-empty).
- Stderr contains timestamp `worktree add` pre-command log line.
- Stderr contains git `worktree add` subprocess output (`Preparing worktree` or `HEAD is now at`).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stdout == "" {
		t.Fatal("stdout should contain worktree path")
	}
	assertStderrContainsGitSubcommand(t, resp.Stderr, "worktree add")
	assertStderrVerboseFormat(t, resp.Stderr)
	assertStderrContainsWorktreeAddOutput(t, resp.Stderr)
}
```