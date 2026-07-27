## Expected

- Exit code 0; stderr empty.
- Same Dir-aware content as `--main --status` order (invocation cwd = external wt).
- Content fields match status-from-main; Dir via `statusDirLine`.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertExitZeroEmptyStderr(t, resp, err)
	assertStdoutMainStatusDirAware(t, req, resp, req.MainRepo, req.WtDir)
}
```
