## Expected

- Exit code 0 (`--projects` never aborts on per-project errors).
- Stdout block includes `Remote:       error:` with fetch failure detail.
- Stderr is empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertContains(t, resp.Stdout, "Remote:")
	assertContains(t, resp.Stdout, "error:")
	assertContains(t, resp.Stdout, "Worktrees:")
}
```