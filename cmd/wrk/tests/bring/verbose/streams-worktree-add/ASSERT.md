## Expected

- Exit code 0; stdout abs external path.
- Stderr contains timestamped `git … worktree add` pre-command log.
- Stderr contains git subprocess progress (`Preparing worktree` or `HEAD is now at`).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)

	assertBringStderrContainsGitWorktreeAdd(t, resp.Stderr)
	assertBringStderrContainsWorktreeAddOutput(t, resp.Stderr)
}
```
