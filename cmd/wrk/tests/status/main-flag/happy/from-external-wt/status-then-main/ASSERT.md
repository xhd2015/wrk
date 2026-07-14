## Expected

- Exit code 0; stderr empty.
- Stdout equals `wrk --status` from main (same as `--main --status` order).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitZeroEmptyStderr(t, resp, err)
	assertStdoutEqualsMainStatus(t, req, resp)
}
```