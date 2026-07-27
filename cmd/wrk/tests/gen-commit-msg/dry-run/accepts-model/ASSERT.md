
## Expected

- Exit code 0.
- Stdout is mock message B for N=1 (model is accepted, unused for dry-run generation).

## Side Effects

- No agent invocation required; pure plan path.

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
