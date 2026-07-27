## Expected

- `err` is non-nil (no go.mod at moduleRoot).
- No successful plan is required; items/name unchecked beyond error.

## Side Effects

- None (pure discovery; may only stat/read paths).

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertPlanError(t, req, resp, err)
}
```
