## Expected

- Exit code 0.
- Stdout matches `git worktree list` exactly.
- Stderr is empty (`worktree list` is minor).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	want := gitWorktreeListIsolated(t, req.RepoDir)
	if resp.Stdout != want {
		t.Fatalf("stdout mismatch:\nwant:\n%q\ngot:\n%q", want, resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```