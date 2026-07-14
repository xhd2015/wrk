## Expected

- Non-zero exit code.
- Stderr contains `not a linked worktree`.
- Main repo unchanged.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertContains(t, resp.Stderr, "not a linked worktree")
}
```