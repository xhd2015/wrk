## Expected

- Non-zero exit.
- Create path exists (no rollback).
- First external **may** exist (apply of dep1 can complete before dep2 fails).
- Stderr mentions that the second path is not a git repository.

## Errors

- Second bring path exists but is not a git repo.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for second not-git apply fail, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	wt := createBringDefaultWT(req)
	assertFileExists(t, wt)
	assertGitFileIsWorktreeLink(t, wt)
	assertContains(t, resp.Stderr, "git")
}
```
