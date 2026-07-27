## Expected

- Exit code 0.
- Stdout matches `git -C <savedRepo> worktree list` exactly.
- Stderr is empty.

## Side Effects

- Worktree list reported for the saved project path resolved via basename fallback, not cwd.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	want := gitWorktreeListIsolated(t, req.MainRepo)
	if resp.Stdout != want {
		t.Fatalf("stdout mismatch:\nwant:\n%q\ngot:\n%q", want, resp.Stdout)
	}

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```