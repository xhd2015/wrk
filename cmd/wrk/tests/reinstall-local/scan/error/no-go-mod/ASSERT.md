## Expected

- `err` is non-nil (no scan root / no go.mod).
- Error text contains `go.mod`.

## Side Effects

- None.

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertScanPlanError(t, req, resp, err)
}
```
