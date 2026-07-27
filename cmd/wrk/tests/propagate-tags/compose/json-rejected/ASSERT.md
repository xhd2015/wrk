## Expected

- Non-zero exit code.
- Stderr mentions `json` and `propagate-tags` (or clear “not valid with
  --propagate-tags” class message).
- Stdout is not a JSON object plan (must not silently accept into bare
  `--tag-next --json`).

## Errors

- `--json` cannot be combined with `--propagate-tags` (compose or bare).

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertJSONRejectedWithPropagate(t, resp)
}
```
