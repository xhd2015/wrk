## Expected Output

```
0 projects  ·  0 modules  ·  0 cross-edges
```

## Expected

- Exit code 0.
- Stdout is exactly the zero footer plus trailing newline (no project blocks).
- Stderr is empty.

## Side Effects

- Read-only registry; no writes required.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	body := graphFooter(0, 0, 0) + "\n"
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(body))
}
```
