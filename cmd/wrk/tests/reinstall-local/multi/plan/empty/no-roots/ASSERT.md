## Expected

- `err` is nil.
- `BinDir` matches the request bin dir.
- `Modules` is empty (length 0).

## Side Effects

- None (pure helper; no module fixtures written).

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertMultiPlanOK(t, req, resp, err)
}
```
