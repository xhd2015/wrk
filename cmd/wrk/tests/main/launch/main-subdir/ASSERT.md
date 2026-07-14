## Expected

- Exit code 0.
- Stdout empty; stderr has no bash-integration install hint.
- Fake interactive shell launched with cwd = main repo root (not the subdir).

## Side Effects

- Nested shell only; no in-place follow-up cd.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertMinimalLaunchUX(t, resp)
	assertFakeShellLaunched(t, req)
	assertFakeShellCwd(t, req, req.MainRepo)
}
```
