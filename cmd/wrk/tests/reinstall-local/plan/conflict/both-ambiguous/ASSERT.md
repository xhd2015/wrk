## Expected

- `err` is nil.
- Items empty (both trees drop bin `foo`).
- Two warnings: `ambiguous-cmd` then `ambiguous-script` (Kind order for same BinName).
- No prefer-script notice.

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
