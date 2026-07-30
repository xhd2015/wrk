## Expected

- Exit code 0.
- Stdout equals the saved project's absolute path plus trailing `\n`.
- Stderr is empty.
- Same result as `wrk --where spl` (operand-then-flag allowed).

## Side Effects

- Read-only lookup; no git or worktree changes.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertStdoutExactPath(t, resp.Stdout, resolvePath(t, req.MainRepo))
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
