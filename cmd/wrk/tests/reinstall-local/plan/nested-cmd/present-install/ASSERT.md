## Expected

- `err` is nil.
- One item: bin `tool`, method `go-install`, path `./cmd/nested/tool`,
  action `install`.

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
