## Expected

- Exit code 0.
- Stdout equals `wrk skill --show wrk-pipeline`.
- Stderr is empty.

## Side Effects

- Read-only.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertSkillShowStdout(t, resp.Stdout, resp.Stderr, resp.ExitCode, "wrk-pipeline")
}
```
