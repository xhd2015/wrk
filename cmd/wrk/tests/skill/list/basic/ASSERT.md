## Expected Output

```
wrk
```

## Expected

- Exit code 0.
- Stdout is exactly `wrk` plus trailing newline.
- Stderr is empty.

## Side Effects

- Read-only; no install or projects.json mutation.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate("wrk\n"))
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
