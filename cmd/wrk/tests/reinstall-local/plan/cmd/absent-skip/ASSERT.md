## Expected

- `err` is nil.
- One plan item for `missing` with `Action: skip` and method `go-install`.
- Skip entries remain in the plan (not silently dropped).

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
