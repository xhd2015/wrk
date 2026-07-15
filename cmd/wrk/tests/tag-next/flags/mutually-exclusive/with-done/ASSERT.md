## Expected

- Non-zero exit code.
- Stderr mentions "mutually exclusive".
- No new tags created.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertContains(t, resp.Stderr, "mutually exclusive")
	if tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("no tag should be created when mode flags conflict")
	}
}
```