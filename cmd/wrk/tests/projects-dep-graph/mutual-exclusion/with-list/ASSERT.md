## Expected

- Non-zero exit code.
- Stderr mentions mutual exclusion (or exclusive / wrk: error).
- Stdout is empty.

## Errors

- `--projects-dep-graph` cannot be combined with `--list`.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertMutualExclusion(t, resp)
}
```
