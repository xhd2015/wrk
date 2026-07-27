
## Expected Output

```text
dry-run: would generate commit message for 1 staged file(s)
```

## Expected

- Exit code 0.
- Stdout is exactly mock message B for N=1 (trailing newline).
- Agent is not required (dry-run pure plan).

## Side Effects

- No commit required; index may remain staged (library dry-run does not unstage text).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertMockMessageB(t, resp.Stdout, 1)
}
```
