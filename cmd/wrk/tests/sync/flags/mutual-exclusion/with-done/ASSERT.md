## Expected

- Non-zero exit code.
- Stderr mentions "mutually exclusive".
- Stdout empty.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertEmptyStdout(t, resp.Stdout)
	assertContains(t, resp.Stderr, "mutually exclusive")
}
```
