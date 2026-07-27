## Expected

- `err` is nil.
- One cmd install item for `foo` (`./cmd/foo`, method `go-install`).
- Diagnostics: only `ambiguous-script` warning (no prefer-script).

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
