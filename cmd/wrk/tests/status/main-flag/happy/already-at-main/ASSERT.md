## Expected

- Exit code 0; stderr empty.
- Stdout equals plain `wrk --status` at the same main root (no shell notice required beyond status).

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitZeroEmptyStderr(t, resp, err)
	assertStdoutEqualsMainStatus(t, req, resp)
}
```