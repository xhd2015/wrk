## Expected

- Exit code 0; stderr empty.
- Same Dir-aware content as `--main --status` order (invocation cwd = external wt).
- Content fields match status-from-main; Dir via `statusDirLine`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitZeroEmptyStderr(t, resp, err)
	assertStdoutMainStatusDirAware(t, req, resp, req.MainRepo, req.WtDir)
}
```
