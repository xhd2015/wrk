## Expected

- `err` is nil.
- One script install item for `foo` (`./script/foo/install`).
- Diagnostics: only `ambiguous-cmd` warning (no `prefer-script` notice).

## Side Effects

- None.

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertPlanOK(t, req, resp, err)
}
```
