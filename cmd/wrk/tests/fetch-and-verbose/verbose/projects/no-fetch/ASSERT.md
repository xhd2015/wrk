## Expected

- Exit code 0.
- Stdout includes healthy project block.
- Stderr is empty (all minor reads, no fetch).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertContains(t, resp.Stdout, "Dir:")
	assertContains(t, resp.Stdout, "Remote:")
}
```