## Expected Output

Same single linked worktree path as flag-first form.

## Expected

- Exit code 0.
- Stdout is the linked worktree absolute path plus trailing `\n`.
- Stderr is empty.

## Side Effects

- Read-only location lookup.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertStdoutExactPath(t, resp.Stdout, resolvePath(t, req.WtDir))
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
