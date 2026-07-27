## Expected

- Exit code 0; stderr empty.
- Stdout matches plain `wrk --status` at the same main root with Dir via statusDirLine
  (identity when inv cwd is main root).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertExitZeroEmptyStderr(t, resp, err)
	assertStdoutMainStatusDirAware(t, req, resp, req.MainRepo)
}
```
